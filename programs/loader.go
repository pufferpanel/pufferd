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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/apufferi/config"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/environments"
)

var (
	allPrograms    []Program = make([]Program, 0)
	ServerFolder   string
	TemplateFolder string
)

func Initialize() {
	ServerFolder = config.GetOrDefault("serverfolder", common.JoinPath("data", "servers"))
	TemplateFolder = config.GetOrDefault("templatefolder", common.JoinPath("data", "templates"))
}

func LoadFromFolder() {
	os.Mkdir(ServerFolder, 0755)
	var programFiles, err = ioutil.ReadDir(ServerFolder)
	if err != nil {
		logging.Critical("Error reading from server data folder", err)
	}
	var program Program
	for _, element := range programFiles {
		if element.IsDir() {
			continue
		}
		id := strings.TrimSuffix(element.Name(), filepath.Ext(element.Name()))
		program, err = Load(id)
		if err != nil {
			logging.Error(fmt.Sprintf("Error loading server details from json (%s)", element.Name()), err)
			continue
		}
		logging.Infof("Loaded server %s", program.Id())
		allPrograms = append(allPrograms, program)
	}
}

func Get(id string) (program Program, err error) {
	program = GetFromCache(id)
	if program == nil {
		program, err = Load(id)
	}
	return
}

func GetAll() []Program {
	return allPrograms
}

func Load(id string) (program Program, err error) {
	var data []byte
	data, err = ioutil.ReadFile(common.JoinPath(ServerFolder, id+".json"))
	if len(data) == 0 || err != nil {
		return
	}
	program, err = LoadFromData(id, data)
	return
}

func LoadFromData(id string, source []byte) (program Program, err error) {
	var data struct {
		ProgramData ProgramData `json:"pufferd"`
	}
	data.ProgramData = CreateProgram()
	err = json.Unmarshal(source, &data)
	if err != nil {
		return
	}

	data.ProgramData.Identifier = id

	environmentType := common.GetStringOrDefault(data.ProgramData.EnvironmentData, "type", "standard")

	data.ProgramData.Environment = environments.LoadEnvironment(environmentType, ServerFolder, id, data.ProgramData.EnvironmentData)
	program = &data.ProgramData
	return
}

func Create(id string, serverType string, data map[string]interface{}) bool {
	if GetFromCache(id) != nil {
		return false
	}

	templateData, err := ioutil.ReadFile(common.JoinPath(TemplateFolder, serverType+".json"))
	if err != nil {
		logging.Error("Error reading template file for type "+serverType, err)
		return false
	}

	var templateJson struct {
		ProgramData ProgramData `json:"pufferd"`
	}
	templateJson.ProgramData = CreateProgram()
	err = json.Unmarshal(templateData, &templateJson)

	if err != nil {
		logging.Error("Error reading template file for type "+serverType, err)
		return false
	}

	if data != nil {
		mapper := templateJson.ProgramData.Data
		if mapper == nil {
			mapper = make(map[string]DataObject, 0)
		}
		for k, v := range data {
			if d, ok := mapper[k]; ok {
				d.Value= v
				mapper[k] = d
			} else {
				newMap := DataObject{
					Value: v,
					Description: "No Description",
					Display: k,
					Required: false,
					Internal: true,
				}
				mapper[k] = newMap
			}
		}
		templateJson.ProgramData.Data = mapper
	}

	f, err := os.Create(common.JoinPath(ServerFolder, id+".json"))

	if err != nil {
		logging.Error("Error writing server file", err)
		return false
	}

	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(templateJson)

	if err != nil {
		logging.Error("Error writing server file", err)
		return false
	}

	newData, err := json.Marshal(templateJson)

	if err != nil {
		logging.Error("Error regenerating file", err)
		return false
	}

	program, _ := LoadFromData(id, newData)
	allPrograms = append(allPrograms, program)
	program.Create()
	return true
}

func Delete(id string) (err error) {
	var index int
	var program Program
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
	os.Remove(common.JoinPath(ServerFolder, program.Id()+".json"))
	allPrograms = append(allPrograms[:index], allPrograms[index+1:]...)
	return
}

func GetFromCache(id string) Program {
	for _, element := range allPrograms {
		if element.Id() == id {
			return element
		}
	}
	return nil
}

func Save(id string) (err error) {
	program := GetFromCache(id)
	if program == nil {
		err = errors.New("No server with given id")
		return
	}
	err = program.Save(common.JoinPath(ServerFolder, id+".json"))
	return
}

func Reload(id string) (err error) {
	temp := GetFromCache(id)
	if temp == nil {
		err = errors.New("server does not exist")
		return
	}
	logging.Infof("Reloading server %s", temp.Id())
	//have to cast it for this to work
	program, _ := temp.(*ProgramData)

	newVersion, err := Load(id)
	if err != nil {
		logging.Error("error reloading server", err)
		return
	}

	newV2 := newVersion.(*ProgramData)

	program.CopyFrom(newV2)
	return
}

func GetPlugins() map[string]interface{} {

	temps, _ := ioutil.ReadDir(TemplateFolder)

	mapping := make(map[string]interface{})

	for _, element := range temps {
		if element.IsDir() {
			continue
		}
		name := strings.TrimSuffix(element.Name(), filepath.Ext(element.Name()))
		data, err := GetPlugin(name)
		if err == nil {
			mapping[name] = data
		}
	}

	return mapping
}

func GetPlugin(name string) (interface{}, error) {
	templateData, err := ioutil.ReadFile(common.JoinPath(TemplateFolder, name+".json"))
	if err != nil {
		return nil, err
	}

	var templateJson map[string]interface{}
	err = json.Unmarshal(templateData, &templateJson)
	if err != nil {
		logging.Error("Malformed json for program "+name, err)
		return nil, err
	}
	segment := common.GetMapOrNull(templateJson, "pufferd")
	dataSec := make(map[string]interface{})
	dataSec["variables"] = segment["data"].(map[string]interface{})
	dataSec["display"] = segment["display"]
	return dataSec, nil
}
