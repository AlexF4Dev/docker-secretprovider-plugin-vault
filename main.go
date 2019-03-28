package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-plugins-helpers/secrets"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

const (
	// Can be set to "vault_token" to return a vault token
	typeLabel      = "dk.almbrand.docker.plugin.secretprovider.vault.type"
	vaultTokenType = "vault_token"
	// Can be set to "true" to wrap the contents of the secret
	wrapLabel = "dk.almbrand.docker.plugin.secretprovider.vault.wrap"

	// Read secret from this path
	pathLabel = "dk.almbrand.docker.plugin.secretprovider.vault.path"
	// Read this field from the secret (defaults to "value")
	fieldLabel = "dk.almbrand.docker.plugin.secretprovider.vault.field"
	// If using v2 key/value backend, use this version of the secret
	versionLabel = "dk.almbrand.docker.plugin.secretprovider.vault.version"
	// Return JSON encoded map of secret if set to "true"
	jsonLabel = "dk.almbrand.docker.plugin.secretprovider.vault.json"
)

var (
	log        = logrus.New()
	secretZero string
)

type vaultSecretsDriver struct {
	dockerClient *dockerclient.Client
	vaultClient  *vaultapi.Client
}

func (d vaultSecretsDriver) Get(req secrets.Request) secrets.Response {
	errorResponse := func(s string, err error) secrets.Response {
		log.Errorf("Error getting secret %q: %s: %v", req.SecretName, s, err)
		return secrets.Response{
			Value: []byte("-"),
			Err:   fmt.Sprintf("%s: %v", s, err),
		}
	}
	valueResponse := func(s string) secrets.Response {
		return secrets.Response{
			Value:      []byte(s),
			DoNotReuse: true,
		}
	}

	// First use secret zero client to create a service token
	serviceToken, err := d.vaultClient.Auth().Token().Create(&vaultapi.TokenCreateRequest{
		Policies: []string{req.ServiceName},
	})
	if err != nil {
		return errorResponse(fmt.Sprintf("Error creating service token with policies like %q", req.ServiceName), err)
	}

	// Create a Vault client
	var vaultClient *vaultapi.Client
	vaultConfig := vaultapi.DefaultConfig()
	if c, err := vaultapi.NewClient(vaultConfig); err != nil {
		log.Fatalf("Error creating Vault client: %v", err)
	} else {
		c.SetToken(serviceToken.Auth.ClientToken)
		vaultClient = c
		defer vaultClient.Auth().Token().RevokeSelf(serviceToken.Auth.ClientToken)
	}

	vaultClient.SetToken(serviceToken.Auth.ClientToken)

	// Inspect the secret to read its labels
	var vaultWrapValue bool
	if v, exists := req.SecretLabels[wrapLabel]; exists {
		if v, err := strconv.ParseBool(v); err == nil {
			vaultWrapValue = v
		} else {
			return errorResponse(fmt.Sprintf("Error parsing boolean value of label %q", wrapLabel), err)
		}
	}

	switch req.SecretLabels[typeLabel] {
	case vaultTokenType:
		// Create a token
		// TODO: Set reasonable default values, and allow configuring them through secret labels
		secret, err := vaultClient.Auth().Token().Create(&vaultapi.TokenCreateRequest{
			Lease:    "1h",
			Policies: []string{"default"},
			Metadata: map[string]string{
				"created_by": os.Args[0],
				// TODO: Add any other interesting metadata
			},
		})
		if err != nil {
			return errorResponse("Error creating token in Vault", err)
		}
		return valueResponse(secret.Auth.ClientToken)
	default:
		var secret *vaultapi.Secret
		// Read from KV secrets mount
		field := ""
		if fieldName, exists := req.SecretLabels[fieldLabel]; exists {
			field = fieldName
		}
		useJSON := false
		if value, exists := req.SecretLabels[jsonLabel]; exists {
			b, err := strconv.ParseBool(value)
			if err != nil {
				return errorResponse(fmt.Sprintf("Error parsing %q as bool", value), err)
			}
			useJSON = b
		}
		path := fmt.Sprintf("secret/data/%s", req.SecretName)
		if v, exists := req.SecretLabels[pathLabel]; exists {
			path = v
		}
		if v, exists := req.SecretLabels[versionLabel]; exists {
			path = fmt.Sprintf("%s?version=%s", path, v)
		}
		secret, err = vaultClient.Logical().Read(path)
		if err != nil {
			return errorResponse(fmt.Sprintf("Error getting kv secret from Vault at path %q", path), err)
		}
		if secret == nil || secret.Data == nil {
			return errorResponse(fmt.Sprintf("Data is nil at path %q (secret: %#v)", path, secret), err)
		}

		data := secret.Data["data"]
		if dataMap, ok := data.(map[string]interface{}); ok {
			if !vaultWrapValue {
				if useJSON {
					var result string
					if len(field) == 0 {
						resultBytes, err := json.Marshal(dataMap)
						if err != nil {
							return errorResponse("Error marshalling secret data map", err)
						}
						result = string(resultBytes)
					} else {
						resultBytes, err := json.Marshal(dataMap[field])
						if err != nil {
							return errorResponse(fmt.Sprintf("Error marshalling secret data field %q", field), err)
						}
						result = string(resultBytes)
					}
					return valueResponse(fmt.Sprintf("%v", result))
				}
				if len(field) == 0 {
					field = "value"
				}
				return valueResponse(fmt.Sprintf("%v", dataMap[field]))
			}
			// Wrap data map
			wrappedSecret, err := vaultClient.Logical().Write("sys/wrapping/wrap", dataMap)
			if err != nil {
				return errorResponse("Error wrapping secret data", err)
			}
			return valueResponse(wrappedSecret.WrapInfo.Token)
		}
		return errorResponse("Invalid data map", err)
	}
}

// Read "secret zero" from the file system of a helper service task container, then serve the plugin.
func main() {
	// Create Docker client
	var client *http.Client
	cli, err := dockerclient.NewClient("unix:///var/run/docker.sock", "1.35", client, nil)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Read plugin configuration from environment
	vaultHelperServiceName := os.Getenv("vault-helper-service")
	secretZeroName := os.Getenv("secret-zero-name")

	// Inspect the helper service
	service, _, err := cli.ServiceInspectWithRaw(context.Background(), vaultHelperServiceName, types.ServiceInspectOptions{})
	if err != nil {
		log.Fatalf("Error inspecting helper service %q: %v", vaultHelperServiceName, err)
	}

	// Look up hostname to filter tasks
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Error getting hostname: %v", err)
	}

	// Find a task on this node, as otherwise we will not be able to exec inside its container
	args := filters.NewArgs(filters.Arg("name", vaultHelperServiceName), filters.Arg("node", hostname))
	tasks, err := cli.TaskList(context.Background(), types.TaskListOptions{
		Filters: args,
	})
	if err != nil {
		log.Fatalf("Error listing tasks for service %q: %v", vaultHelperServiceName, err)
	}

	// Look for a task from the helper service
	var secretZero string
	for _, task := range tasks {
		// avoid services with the name as a shared prefix but different ID
		if task.ServiceID != service.ID {
			continue
		}
		// Use a task that has a container
		containerStatus := task.Status.ContainerStatus
		if containerStatus != nil {
			// Create an exec to later read its output
			response, err := cli.ContainerExecCreate(context.Background(), containerStatus.ContainerID, types.ExecConfig{
				AttachStdout: true,
				Detach:       false,
				Tty:          false,
				Cmd:          []string{"cat", fmt.Sprintf("/run/secrets/%s", secretZeroName)},
			})
			if err != nil {
				log.Fatalf("Error creating exec: %v", err)
			}
			// Start and attach to exec to read its output
			resp, err := cli.ContainerExecAttach(context.Background(), response.ID, types.ExecStartCheck{
				Detach: false,
				Tty:    false,
			})
			if err != nil {
				log.Fatalf("Error attaching to exec: %v", err)
			}
			defer resp.Close()
			// Read the output into a buffer and convert to a string
			buf := new(bytes.Buffer)
			if _, err := stdcopy.StdCopy(buf, buf, resp.Reader); err != nil {
				if err != nil {
					log.Fatalf("Error reading secret zero: %v", err)
				}
			}
			secretZero = buf.String()
			break
		}
	}
	if len(secretZero) == 0 {
		log.Fatalf("Failed to read a Vault token from the helper service %q", vaultHelperServiceName)
	}

	// Create a Vault client
	var vaultClient *vaultapi.Client
	vaultConfig := vaultapi.DefaultConfig()
	if c, err := vaultapi.NewClient(vaultConfig); err != nil {
		log.Fatalf("Error creating Vault client: %v", err)
	} else {
		c.SetToken(secretZero)
		vaultClient = c
	}

	// Create the driver
	d := vaultSecretsDriver{
		dockerClient: cli,
		vaultClient:  vaultClient,
	}
	h := secrets.NewHandler(d)

	// Serve plugin
	if err := h.ServeUnix("plugin", 0); err != nil {
		log.Errorf("Error serving plugin: %v", err)
	}
}
