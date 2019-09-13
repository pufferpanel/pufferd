/*
 Copyright 2016 Padduck, LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 distributed under the License is distributed on an "AS IS" BASIS,
 You may obtain a copy of the License at

 	http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package programs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pufferpanel/pufferd/v2/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pufferpanel/apufferi/v3"
	"github.com/pufferpanel/apufferi/v3/logging"
	"github.com/pufferpanel/pufferd/v2/environments"
	"github.com/pufferpanel/pufferd/v2/programs/operations"
)

var (
	allPrograms  = make([]*Program, 0)
	ServerFolder string
)

func Initialize() {
	ServerFolder = config.Get().Data.ServerFolder

	operations.LoadOperations()
}

func LoadFromFolder() {
	err := os.Mkdir(ServerFolder, 0755)
	if err != nil && !os.IsExist(err) {
		logging.Critical("Error creating server data folder: %s", err)
	}
	programFiles, err := ioutil.ReadDir(ServerFolder)
	if err != nil {
		logging.Critical("Error reading from server data folder: %s", err)
	}
	var program *Program
	for _, element := range programFiles {
		if element.IsDir() {
			continue
		}
		id := strings.TrimSuffix(element.Name(), filepath.Ext(element.Name()))
		program, err = Load(id)
		if err != nil {
			logging.Exception(fmt.Sprintf("Error loading server details from json (%s)", element.Name()), err)
			continue
		}
		logging.Info("Loaded server %s", program.Id())
		allPrograms = append(allPrograms, program)
	}
}

func Get(id string) (program *Program, err error) {
	program = GetFromCache(id)
	if program == nil {
		program, err = Load(id)
	}
	return
}

func GetAll() []*Program {
	return allPrograms
}

func Load(id string) (program *Program, err error) {
	var data []byte
	data, err = ioutil.ReadFile(apufferi.JoinPath(ServerFolder, id+".json"))
	if len(data) == 0 || err != nil {
		return
	}
	program, err = LoadFromData(id, data)
	return
}

func LoadFromData(id string, source []byte) (*Program, error) {
	data := CreateProgram()
	err := json.Unmarshal(source, &data)
	if err != nil {
		return nil, err
	}

	data.Identifier = id

	environmentType := data.Server.Environment.Type
	data.Environment, err = environments.Create(environmentType, ServerFolder, id, data.Server.Environment)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Create(program *Program) bool {
	if GetFromCache(program.Id()) != nil {
		return false
	}

	var err error

	defer func() {
		if err != nil {
			//revert since we have an error
			_ = os.Remove(apufferi.JoinPath(ServerFolder, program.Id()+".json"))
			if program.Environment != nil {
				_ = program.Environment.Delete()
			}
		}
	}()

	f, err := os.Create(apufferi.JoinPath(ServerFolder, program.Id()+".json"))
	defer apufferi.Close(f)
	if err != nil {
		logging.Exception("error writing server", err)
		return false
	}

	encoder := json.NewEncoder(f)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(program)

	if err != nil {
		logging.Exception("error writing server", err)
		return false
	}

	environmentType := program.Server.Environment.Type
	program.Environment, err = environments.Create(environmentType, ServerFolder, program.Id(), program.Server.Environment)

	err = program.Create()
	if err != nil {
		return false
	}

	allPrograms = append(allPrograms, program)
	return true
}

func Delete(id string) (err error) {
	var index int
	var program *Program
	for i, element := range allPrograms {
		if element.Id() == id {
			program = element
			index = i
			break
		}
	}
	if program == nil {
		return
	}
	running, err := program.IsRunning()

	if err != nil {
		return
	}

	if running {
		err = program.Stop()
		if err != nil {
			return
		}
	}

	err = program.Destroy()
	if err != nil {
		return
	}
	err = os.Remove(apufferi.JoinPath(ServerFolder, program.Id()+".json"))
	if err != nil {
		logging.Exception("error removing server", err)
	}
	allPrograms = append(allPrograms[:index], allPrograms[index+1:]...)
	return
}

func GetFromCache(id string) *Program {
	for _, element := range allPrograms {
		if element != nil && element.Id() == id {
			return element
		}
	}
	return nil
}

func Save(id string) (err error) {
	program := GetFromCache(id)
	if program == nil {
		err = errors.New("no server with given id")
		return
	}
	err = program.Save(apufferi.JoinPath(ServerFolder, id+".json"))
	return
}

func Reload(id string) (err error) {
	program := GetFromCache(id)
	if program == nil {
		err = errors.New("server does not exist")
		return
	}
	logging.Info("Reloading server %s", program.Id())
	newVersion, err := Load(id)
	if err != nil {
		logging.Exception("error reloading server", err)
		return
	}

	program.CopyFrom(newVersion)
	return
}
