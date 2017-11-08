/*
 Copyright 2016 Padduck, LLC

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
	"io/ioutil"
	"os"

	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/environments"
	"github.com/pufferpanel/pufferd/programs/operations"
	"errors"
)

type Program interface {
	//Starts the program.
	//This includes starting the environment if it is not running.
	Start() (err error)

	//Stops the program.
	//This will also stop the environment it is ran in.
	Stop() (err error)

	//Kills the program.
	//This will also stop the environment it is ran in.
	Kill() (err error)

	//Creates any files needed for the program.
	//This includes creating the environment.
	Create() (err error)

	//Destroys the server.
	//This will delete the server, environment, and any files related to it.
	Destroy() (err error)

	Install() (err error)

	//Determines if the server is running.
	IsRunning() (isRunning bool, err error)

	//Sends a command to the process
	//If the program supports input, this will send the arguments to that.
	Execute(command string) (err error)

	SetEnabled(isEnabled bool) (err error)

	IsEnabled() (isEnabled bool)

	SetAutoStart(isAutoStart bool) (err error)

	IsAutoStart() (isAutoStart bool)

	SetEnvironment(environment environments.Environment) (err error)

	Id() string

	GetEnvironment() environments.Environment

	Save(file string) (err error)

	Edit(data map[string]interface{}) (err error)

	GetData() map[string]DataObject

	GetNetwork() string
}

//Starts the program.
//This includes starting the environment if it is not running.
func (p *ProgramData) Start() (err error) {
	if !p.IsEnabled() {
		logging.Errorf("Server %s is not enabled, cannot start", p.Id())
		return errors.New("server not enabled")
	}
	logging.Debugf("Starting server %s", p.Id())
	p.Environment.DisplayToConsole("Starting server\n")
	data := make(map[string]interface{})
	for k, v := range p.Data {
		data[k] = v.Value
	}

	process := operations.GenerateProcess(p.RunData.Pre, p.Environment, p.DataToMap())
	err = process.Run()
	if err != nil {
		p.Environment.DisplayToConsole("Error running pre execute, check daemon logs")
		return
	}
	
	err = p.Environment.ExecuteAsync(p.RunData.Program, common.ReplaceTokensInArr(p.RunData.Arguments, data), func(graceful bool) {
	})
	if err != nil {
		logging.Error("Error starting server", err)
		p.Environment.DisplayToConsole("Failed to start server\n")
	}

	return
}

//Stops the program.
//This will also stop the environment it is ran in.
func (p *ProgramData) Stop() (err error) {
	logging.Debugf("Stopping server %s", p.Id())
	err = p.Environment.ExecuteInMainProcess(p.RunData.Stop)
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
	logging.Debugf("Killing server %s", p.Id())
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
	logging.Debugf("Creating server %s", p.Id())
	p.Environment.DisplayToConsole("Allocating server\n")
	err = p.Environment.Create()
	p.Environment.DisplayToConsole("Server allocated\n")
	p.Environment.DisplayToConsole("Ready to be installed\n")
	return
}

//Destroys the server.
//This will delete the server, environment, and any files related to it.
func (p *ProgramData) Destroy() (err error) {
	logging.Debugf("Destroying server %s", p.Id())
	err = p.Environment.Delete()
	return
}

func (p *ProgramData) Install() (err error) {
	if !p.IsEnabled() {
		logging.Errorf("Server %s is not enabled, cannot install", p.Id())
		return errors.New("server not enabled")
	}

	logging.Debugf("Installing server %s", p.Id())
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

	process := operations.GenerateProcess(p.InstallData.Operations, p.Environment, p.DataToMap())
	err = process.Run()
	if err != nil {
		p.Environment.DisplayToConsole("Error running installer, check daemon logs")
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

func (p *ProgramData) SetEnvironment(environment environments.Environment) (err error) {
	p.Environment = environment
	return
}

func (p *ProgramData) Id() string {
	return p.Identifier
}

func (p *ProgramData) GetEnvironment() environments.Environment {
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
	logging.Debugf("Saving server %s", p.Id())

	endResult := make(map[string]interface{})
	endResult["pufferd"] = p

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
