package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-plugins-helpers/secrets"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.New()
)

type MySecretsDriver struct {
	VaultToken string
}

func (d MySecretsDriver) Get(req secrets.Request) secrets.Response {
	vaultConfig := vaultapi.DefaultConfig()
	c, err := vaultapi.NewClient(vaultConfig)

	if err != nil {
		return secrets.Response{
			Err: "Ohnoes",
		}
	}
	c.SetToken(d.VaultToken)
	var client *http.Client
	cli, err := dockerclient.NewClient("unix:///var/run/docker.sock", "1.35", client, nil)
	if err != nil {
		log.Errorf("Failed to create Docker client: %v", err)
	}
	swarmSecret, _, err := cli.SecretInspectWithRaw(context.Background(), req.SecretName)
	if err != nil {
		log.Errorf("Error inspecting secret in Swarm: %v", err)
		return secrets.Response{Err: fmt.Sprintf("Error inspecting secret in Swarm: %v", err)}
	}
	var specialTypeLabel string
	if v, ok := swarmSecret.Spec.Labels["dk.almbrand.docker.plugin.secretprovider.vault.type"]; ok {
		specialTypeLabel = v
	}
	var secret *vaultapi.Secret
	isVaultToken := specialTypeLabel == "vault_token" // || req.SecretName == "VAULT_TOKEN"
	if isVaultToken {
		secret, err = c.Auth().Token().Create(&vaultapi.TokenCreateRequest{
			Lease:    "1h",
			Policies: []string{"default"},
		})
	} else {
		secret, err = c.Logical().Read(fmt.Sprintf("secret/data/%s", req.SecretName))
	}
	if err != nil {
		log.Errorf("Error getting something from Vault: %v", err)
		return secrets.Response{
			Value: []byte(fmt.Sprint(err)),
			Err:   fmt.Sprint(err),
		}
	}

	if isVaultToken {
		return secrets.Response{
			Value: []byte(fmt.Sprint(secret.Auth.ClientToken)),
		}
	}
	if secret == nil || secret.Data == nil {
		return secrets.Response{Err: "Data is nil"}
	}

	data := secret.Data["data"]
	if dataMap, ok := data.(map[string]interface{}); ok {
		return secrets.Response{
			Value: []byte(fmt.Sprintf("%v", dataMap["value"])),
		}
	} else {
		return secrets.Response{Err: "Invalid data map"}
	}

}

func main() {
	var client *http.Client
	cli, err := dockerclient.NewClient("unix:///var/run/docker.sock", "1.35", client, nil)
	if err != nil {
		log.Errorf("Failed to create Docker client: %v", err)
	}
	vaultHelperServiceName := os.Getenv("vault-helper-service")
	secretZeroName := os.Getenv("secret-zero-name")
	service, _, err := cli.ServiceInspectWithRaw(context.Background(), vaultHelperServiceName, types.ServiceInspectOptions{})
	if err != nil {
		log.Errorf("Error inspecting helper service %q: %v", vaultHelperServiceName, err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("Error getting hostname: %v", err)
	}
	args := filters.NewArgs(filters.Arg("name", vaultHelperServiceName), filters.Arg("node", hostname))
	tasks, err := cli.TaskList(context.Background(), types.TaskListOptions{
		Filters: args,
	})
	if err != nil {
		log.Errorf("Error listing tasks for service %q: %v", vaultHelperServiceName, err)
	}
	var vaultToken string
	for i, task := range tasks {
		if task.ServiceID != service.ID {
			tasks = append(tasks[0:i], tasks[i+1:]...)
		}
		containerStatus := task.Status.ContainerStatus
		if containerStatus != nil {
			id := containerStatus.ContainerID
			execConfig := types.ExecConfig{
				AttachStdout: true,
				Detach:       false,
				Tty:          false,
				Cmd:          []string{"cat", fmt.Sprintf("/run/secrets/%s", secretZeroName)},
			}
			response, err := cli.ContainerExecCreate(context.Background(), id, execConfig)
			if err != nil {
				log.Errorf("Error creating exec: %v", err)
			}
			execID := response.ID
			if execID == "" {
				log.Errorf("exec ID empty")
			}
			execStartCheck := types.ExecStartCheck{
				Detach: false,
				Tty:    false,
			}
			resp, err := cli.ContainerExecAttach(context.Background(), execID, execStartCheck)
			if err != nil {
				log.Errorf("Error attaching to exec: %v", err)
			}
			defer resp.Close()
			buf := new(bytes.Buffer)
			if _, err := stdcopy.StdCopy(buf, buf, resp.Reader); err != nil {
				if err != nil {
					log.Errorf("Error reading secret zero: %v", err)
				}
			}
			vaultToken = buf.String()
		}
	}

	d := MySecretsDriver{
		VaultToken: vaultToken,
	}
	h := secrets.NewHandler(d)
	if err := h.ServeUnix("secrets-plugin", 0); err != nil {
		log.Errorf("Error serving plugin: %v", err)
	}
}
