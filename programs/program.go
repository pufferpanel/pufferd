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
	"container/list"
	"sync"
	"time"
	"github.com/pufferpanel/pufferd/environments"
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

var queue *list.List
var lock = sync.Mutex{}
var ticker *time.Ticker
var running = false

func InitService() {
	queue = list.New()
	ticker = time.NewTicker(1 * time.Second)
	running = true
	go func() {
		for range ticker.C {
			lock.Lock()
			next := queue.Front()
			lock.Unlock()
			if next == nil {
				continue
			}
			program := next.Value.(Program)
			if run, _ := program.IsRunning(); !run {
				program.Start()
			}
		}
	}()
}

func StartViaService (p Program) {
	lock.Lock()
	defer func() {
		lock.Unlock()
	}()

	if running {
		queue.PushBack(p)
	}
}

func ShutdownService() {
	lock.Lock()
	defer func() {
		lock.Unlock()
	}()

	running = false
	ticker.Stop()
}