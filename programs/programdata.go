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

package programs

import (
	"encoding/json"
	"errors"
	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/apufferi/config"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/environments/envs"
	"github.com/pufferpanel/pufferd/programs/operations"
	"io/ioutil"
	"os"
)

type ServerJson struct {
	ProgramData ProgramData `json:"pufferd"`
}

type ProgramData struct {
	Data            map[string]DataObject  `json:"data,omitempty"`
	Display         string                 `json:"display,omitempty"`
	EnvironmentData map[string]interface{} `json:"environment,omitempty"`
	InstallData     InstallSection         `json:"install,omitempty"`
	UninstallData   InstallSection         `json:"uninstall,omitempty"`
	Type            string                 `json:"type,omitempty"`
	Identifier      string                 `json:"id,omitempty"`
	RunData         RunObject              `json:"run,omitempty"`
	Template        string                 `json:"template,omitempty"`

	Environment  envs.Environment `json:"-"`
	CrashCounter int                      `json:"-"`
}

type DataObject struct {
	Description  string      `json:"desc,omitempty"`
	Display      string      `json:"display,omitempty"`
	Internal     bool        `json:"internal,omitempty"`
	Required     bool        `json:"required,omitempty"`
	Value        interface{} `json:"value,omitempty"`
	UserEditable bool        `json:"userEdit,omitempty"`
}

type RunObject struct {
	Arguments               []string                 `json:"arguments,omitempty"`
	Program                 string                   `json:"program,omitempty"`
	Stop                    string                   `json:"stop,omitempty"`
	Enabled                 bool                     `json:"enabled,omitempty"`
	AutoStart               bool                     `json:"autostart,omitempty"`
	AutoRestartFromCrash    bool                     `json:"autorecover,omitempty"`
	AutoRestartFromGraceful bool                     `json:"autorestart,omitempty"`
	Pre                     []map[string]interface{} `json:"pre,omitempty"`
	Post                    []map[string]interface{} `json:"post,omitempty"`
	StopCode                int                      `json:"stopCode,omitempty"`
	EnvironmentVariables    map[string]string        `json:"environmentVars,omitempty"`
}

type InstallSection struct {
	Operations []map[string]interface{} `json:"commands,,omitempty"`
}

func (p ProgramData) DataToMap() map[string]interface{} {
	var result = make(map[string]interface{}, len(p.Data))

	for k, v := range p.Data {
		result[k] = v.Value
	}

	return result
}

func CreateProgram() ProgramData {
	return ProgramData{
		RunData: RunObject{
			Enabled:                 true,
			AutoStart:               false,
			AutoRestartFromCrash:    false,
			AutoRestartFromGraceful: false,
			Pre:                  make([]map[string]interface{}, 0),
			Post:                 make([]map[string]interface{}, 0),
			EnvironmentVariables: make(map[string]string, 0),
		},
		Type:    "standard",
		Data:    make(map[string]DataObject, 0),
		Display: "Unknown server",
		InstallData: InstallSection{
			Operations: make([]map[string]interface{}, 0),
		},
	}
}

//Starts the program.
//This includes starting the environment if it is not running.
func (p *ProgramData) Start() (err error) {
	if !p.IsEnabled() {
		logging.Error("Server %s is not enabled, cannot start", p.Id())
		return errors.New("server not enabled")
	}
	logging.Debug("Starting server %s", p.Id())
	p.Environment.DisplayToConsole("Starting server\n")
	data := make(map[string]interface{})
	for k, v := range p.Data {
		data[k] = v.Value
	}

	process := operations.GenerateProcess(p.RunData.Pre, p.Environment, p.DataToMap(), p.RunData.EnvironmentVariables)
	err = process.Run(p.Environment)
	if err != nil {
		p.Environment.DisplayToConsole("Error running pre execute, check daemon logs\n")
		return
	}

	//HACK: add rootDir stuff
	data["rootDir"] = p.Environment.GetRootDirectory()

	err = p.Environment.ExecuteAsync(p.RunData.Program, common.ReplaceTokensInArr(p.RunData.Arguments, data), common.ReplaceTokensInMap(p.RunData.EnvironmentVariables, data), p.afterExit)
	if err != nil {
		logging.Error("Error starting server", err)
		p.Environment.DisplayToConsole("Failed to start server\n")
	}

	return
}

//Stops the program.
//This will also stop the environment it is ran in.
func (p *ProgramData) Stop() (err error) {
	logging.Debug("Stopping server %s", p.Id())
	if p.RunData.StopCode != 0 {
		err = p.Environment.SendCode(p.RunData.StopCode)
	} else {
		err = p.Environment.ExecuteInMainProcess(p.RunData.Stop)
	}
	if err != nil {
		p.Environment.DisplayToConsole("Failed to stop server\n")
	} else {
		p.Environment.DisplayToConsole("Server stopped\n")
	}
	return
}

//Kills the program.
//This will also stop the environment it is ran in.
func (p *ProgramData) Kill() (err error) {
	logging.Debug("Killing server %s", p.Id())
	err = p.Environment.Kill()
	if err != nil {
		p.Environment.DisplayToConsole("Failed to kill server\n")
	} else {
		p.Environment.DisplayToConsole("Server killed\n")
	}
	return
}

//Creates any files needed for the program.
//This includes creating the environment.
func (p *ProgramData) Create() (err error) {
	logging.Debug("Creating server %s", p.Id())
	p.Environment.DisplayToConsole("Allocating server\n")
	err = p.Environment.Create()
	p.Environment.DisplayToConsole("Server allocated\n")
	p.Environment.DisplayToConsole("Ready to be installed\n")
	return
}

//Destroys the server.
//This will delete the server, environment, and any files related to it.
func (p *ProgramData) Destroy() (err error) {
	logging.Debug("Destroying server %s", p.Id())
	process := operations.GenerateProcess(p.UninstallData.Operations, p.Environment, p.DataToMap(), p.RunData.EnvironmentVariables)
	err = process.Run(p.Environment)
	if err != nil {
		p.Environment.DisplayToConsole("Error running uninstall, check daemon logs\n")
		return
	}
	err = p.Environment.Delete()
	return
}

func (p *ProgramData) Install() (err error) {
	if !p.IsEnabled() {
		logging.Error("Server %s is not enabled, cannot install", p.Id())
		return errors.New("server not enabled")
	}

	logging.Debug("Installing server %s", p.Id())
	running, err := p.IsRunning()
	if err != nil {
		logging.Error("Error stopping server to install: ", err)
		p.Environment.DisplayToConsole("Error stopping server\n")
		return
	}

	if running {
		err = p.Stop()
	}

	if err != nil {
		logging.Error("Error stopping server to install: ", err)
		p.Environment.DisplayToConsole("Error stopping server\n")
		return
	}

	p.Environment.DisplayToConsole("Installing server\n")

	os.MkdirAll(p.Environment.GetRootDirectory(), 0755)

	var process operations.OperationProcess

	if len(p.InstallData.Operations) == 0 && p.Template != "" {
		logging.Debug("Server %s has no defined install data, using template", p.Id())
		templateData, err := ioutil.ReadFile(common.JoinPath(TemplateFolder, p.Template+".json"))
		if err != nil {
			logging.Error("Error reading template for "+p.Template, err)
			p.Environment.DisplayToConsole("Error running installer, check daemon logs\n")
			return err
		}

		templateJson := ServerJson{}
		err = json.Unmarshal(templateData, &templateJson)
		if err != nil {
			logging.Error("Malformed json for program "+p.Template, err)
			p.Environment.DisplayToConsole("Error running installer, check daemon logs\n")
			return err
		}

		process = operations.GenerateProcess(templateJson.ProgramData.InstallData.Operations, p.GetEnvironment(), p.DataToMap(), p.RunData.EnvironmentVariables)

	} else {
		logging.Debug("Server %s has defined install data", p.Id())
		process = operations.GenerateProcess(p.InstallData.Operations, p.GetEnvironment(), p.DataToMap(), p.RunData.EnvironmentVariables)
	}

	err = process.Run(p.Environment)
	if err != nil {
		p.Environment.DisplayToConsole("Error running installer, check daemon logs\n")
	} else {
		p.Environment.DisplayToConsole("Server installed\n")
	}
	return
}

//Determines if the server is running.
func (p *ProgramData) IsRunning() (isRunning bool, err error) {
	isRunning, err = p.Environment.IsRunning()
	return
}

//Sends a command to the process
//If the program supports input, this will send the arguments to that.
func (p *ProgramData) Execute(command string) (err error) {
	err = p.Environment.ExecuteInMainProcess(command)
	return
}

func (p *ProgramData) SetEnabled(isEnabled bool) (err error) {
	p.RunData.Enabled = isEnabled
	return
}

func (p *ProgramData) IsEnabled() (isEnabled bool) {
	isEnabled = p.RunData.Enabled
	return
}

func (p *ProgramData) SetEnvironment(environment envs.Environment) (err error) {
	p.Environment = environment
	return
}

func (p *ProgramData) Id() string {
	return p.Identifier
}

func (p *ProgramData) GetEnvironment() envs.Environment {
	return p.Environment
}

func (p *ProgramData) SetAutoStart(isAutoStart bool) (err error) {
	p.RunData.AutoStart = isAutoStart
	return
}

func (p *ProgramData) IsAutoStart() (isAutoStart bool) {
	isAutoStart = p.RunData.AutoStart
	return
}

func (p *ProgramData) Save(file string) (err error) {
	logging.Debug("Saving server %s", p.Id())

	endResult := ServerJson{}
	endResult.ProgramData = *p

	data, err := json.MarshalIndent(endResult, "", "  ")
	if err != nil {
		return
	}

	err = ioutil.WriteFile(file, data, 0664)
	return
}

func (p *ProgramData) Edit(data map[string]interface{}) (err error) {
	for k, v := range data {
		if v == nil || v == "" {
			delete(p.Data, k)
		}

		var elem DataObject

		if _, ok := p.Data[k]; ok {
			elem = p.Data[k]
		} else {
			elem = DataObject{}
		}
		elem.Value = v

		p.Data[k] = elem
	}
	err = Save(p.Id())
	return
}

func (p *ProgramData) GetData() map[string]DataObject {
	return p.Data
}

func (p *ProgramData) GetNetwork() string {
	data := p.GetData()
	ip := "0.0.0.0"
	port := "0"

	if ipData, ok := data["ip"]; ok {
		ip = ipData.Value.(string)
	}

	if portData, ok := data["port"]; ok {
		port = portData.Value.(string)
	}

	return ip + ":" + port
}

func (p *ProgramData) CopyFrom(s *ProgramData) {
	p.Data = s.Data
	p.RunData = s.RunData
	p.Display = s.Display
	p.EnvironmentData = s.EnvironmentData
	p.InstallData = s.InstallData
	p.Type = s.Type
	p.Template = s.Template
}

func (p *ProgramData) afterExit(graceful bool) {
	if graceful {
		p.CrashCounter = 0
	}

	mapping := p.DataToMap()
	mapping["success"] = graceful

	processes := operations.GenerateProcess(p.RunData.Post, p.Environment, mapping, p.RunData.EnvironmentVariables)

	p.Environment.DisplayToConsole("Running post-execution steps\n")
	logging.Debug("Running post execution steps: %s", p.Id())

	err := processes.Run(p.Environment)
	if err != nil {
		logging.Error("Error running post processing")
		p.Environment.DisplayToConsole("Error executing post steps\n")
		return
	}

	if !p.RunData.AutoRestartFromCrash && !p.RunData.AutoRestartFromGraceful {
		return
	}

	if graceful && p.RunData.AutoRestartFromGraceful {
		StartViaService(p)
	} else if !graceful && p.RunData.AutoRestartFromCrash && p.CrashCounter < config.GetIntOrDefault("crashlimit", 3) {
		p.CrashCounter++
		StartViaService(p)
	}
}
