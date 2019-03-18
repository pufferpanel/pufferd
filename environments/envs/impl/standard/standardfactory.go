package standard

import (
	"github.com/pufferpanel/apufferi/cache"
	"github.com/pufferpanel/pufferd/environments/envs"
	"github.com/pufferpanel/pufferd/utils"
	"sync"
)

type EnvironmentFactory struct {
	envs.EnvironmentFactory
}

func (ef EnvironmentFactory) Create(folder, id string, environmentSection map[string]interface{}, rootDirectory string, cache cache.Cache, wsManager utils.WebSocketManager) envs.Environment {
	s := &standard{BaseEnvironment: &envs.BaseEnvironment{Type: "standard"}}
	s.BaseEnvironment.ExecutionFunction = s.standardExecuteAsync
	s.BaseEnvironment.WaitFunction = s.WaitForMainProcess
	s.wait = &sync.WaitGroup{}

	s.RootDirectory = rootDirectory
	s.ConsoleBuffer = cache
	s.WSManager = wsManager
	return s
}

func (ef EnvironmentFactory) Key() string {
	return "standard"
}
