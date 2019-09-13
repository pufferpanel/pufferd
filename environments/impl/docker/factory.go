/*
 Copyright 2019 Padduck, LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 	http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package docker

import (
	"github.com/pufferpanel/apufferi/v3"
	"github.com/pufferpanel/pufferd/v2/environments/envs"
	"github.com/pufferpanel/pufferd/v2/utils"
	"sync"
)

type EnvironmentFactory struct {
	envs.EnvironmentFactory
}

func (ef EnvironmentFactory) Create(folder, id string, environmentSection map[string]interface{}, rootDirectory string, cache apufferi.Cache, wsManager utils.WebSocketManager) envs.Environment {
	imageName := apufferi.GetStringOrDefault(environmentSection, "image", "")
	enforceNetwork := apufferi.GetBooleanOrDefault(environmentSection, "enforceNetwork", false)

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
