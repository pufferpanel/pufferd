package docker

import (
	"github.com/pufferpanel/apufferi/cache"
	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/pufferd/environments/envs"
	"github.com/pufferpanel/pufferd/utils"
	"sync"
)

type EnvironmentFactory struct {
	envs.EnvironmentFactory
}

func (ef EnvironmentFactory) Create(folder, id string, environmentSection map[string]interface{}, rootDirectory string, cache cache.Cache, wsManager utils.WebSocketManager) envs.Environment {
	imageName := common.GetStringOrDefault(environmentSection, "image", "")
	enforceNetwork := common.GetBooleanOrDefault(environmentSection, "enforceNetwork", false)

	if imageName == "" {
		imageName = "pufferpanel/generic"
	}

	d := &docker{BaseEnvironment: &envs.BaseEnvironment{Type: "docker"}, ContainerId: id, ImageName: imageName, enforceNetwork: enforceNetwork}
	d.BaseEnvironment.ExecutionFunction = d.dockerExecuteAsync
	d.BaseEnvironment.WaitFunction = d.WaitForMainProcess
	d.wait = &sync.WaitGroup{}

	d.RootDirectory = rootDirectory
	d.ConsoleBuffer = cache
	d.WSManager = wsManager
	return d
}

func (ef EnvironmentFactory) Key() string {
	return "docker"
}
