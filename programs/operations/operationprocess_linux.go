/*
 Copyright 2018 Padduck, LLC

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

package operations

import (
	"io/ioutil"
	"os"
	"path"
	"plugin"
	"reflect"

	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/config"
	"github.com/pufferpanel/pufferd/programs/operations/ops"
)

func loadOpModules() {
	var directory = path.Join(config.Get().Data.ModuleFolder, "operations")

	files, err := ioutil.ReadDir(directory)
	if err != nil && os.IsNotExist(err) {
		return
	} else if err != nil {
		logging.Error("Error reading directory", err)
	}

	for _, file := range files {
		logging.Info("Loading operation module: %s", file.Name())
		p, e := plugin.Open(path.Join(directory, file.Name()))
		if err != nil {
			logging.Error("Unable to load module", e)
			continue
		}

		factory, e := p.Lookup("Factory")
		if err != nil {
			logging.Error("Unable to load module", e)
			continue
		}

		fty, ok := factory.(ops.OperationFactory)
		if !ok {
			logging.Error("Expected OperationFactory, but found %s", reflect.TypeOf(factory).Name())
			continue
		}

		commandMapping[fty.Key()] = fty

		logging.Info("Loaded operation module: %s", fty.Key())
	}
}
