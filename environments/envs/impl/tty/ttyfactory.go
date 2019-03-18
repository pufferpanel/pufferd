// +build !windows

package tty

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
	t := &tty{BaseEnvironment: &envs.BaseEnvironment{Type: "tty"}}
	t.BaseEnvironment.ExecutionFunction = t.ttyExecuteAsync
	t.BaseEnvironment.WaitFunction = t.WaitForMainProcess
	t.wait = &sync.WaitGroup{}

	t.RootDirectory = rootDirectory
	t.ConsoleBuffer = cache
	t.WSManager = wsManager
	return t
}

func (ef EnvironmentFactory) Key() string {
	return "tty"
}
