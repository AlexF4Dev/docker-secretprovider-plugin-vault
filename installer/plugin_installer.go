package main

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/swarm/runtime"
	dockerclient "github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.New()
)

func main() {
	cli, err := dockerclient.NewEnvClient()
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}
	serviceName := "docker-secretprovider-plugin-vault"
	pluginName := "sirlatrom/docker-secretprovider-plugin-vault:latest"
	if override, exists := os.LookupEnv("plugin_name"); exists {
		pluginName = override
		re := regexp.MustCompile("^(?:.+/|)([^:$]+)(?::.*|)$")
		serviceName = re.ReplaceAllString(pluginName, "${1}")
	}
	remote := pluginName
	if override, exists := os.LookupEnv("remote"); exists {
		remote = override
	}
	service, err := cli.ServiceCreate(context.Background(), swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: serviceName,
		},
		TaskTemplate: swarm.TaskSpec{
			PluginSpec: &runtime.PluginSpec{
				Name:     pluginName,
				Remote:   remote,
				Disabled: false,
				Privileges: []*runtime.PluginPrivilege{
					&runtime.PluginPrivilege{
						Name:        "network",
						Description: "permissions to access a network",
						Value:       []string{"host"},
					},
					&runtime.PluginPrivilege{
						Name:        "mount",
						Description: "host path to mount",
						Value:       []string{"/var/run/docker.sock"},
					},
					&runtime.PluginPrivilege{
						Name:        "capabilities",
						Description: "list of additional capabilities required",
						Value:       []string{"CAP_SYS_ADMIN"},
					},
				},
				Env: []string{
					"policy-template={{ .ServiceName }},{{ .TaskImage }},{{ ServiceLabel \"com.docker.ucp.access.label\" }}",
					"DOCKER_API_VERSION=1.37",
					"baz",
					"foo=bar",
				},
			},
			Placement: &swarm.Placement{
				Constraints: []string{"node.role == manager"},
			},
			Runtime: swarm.RuntimePlugin,
		},
	}, types.ServiceCreateOptions{})
	if err != nil {
		log.Fatalf("Failed to create plugin service: %v", err)
	}
	fmt.Println(service.ID)
}
