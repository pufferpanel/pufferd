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

package commands

import (
	"flag"
	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/config"
	"github.com/pufferpanel/apufferi/logging"
	config2 "github.com/pufferpanel/pufferd/config"
	"github.com/pufferpanel/pufferd/environments"
	"github.com/pufferpanel/pufferd/programs"
	"github.com/pufferpanel/pufferd/routing"
	"github.com/pufferpanel/pufferd/sftp"
	"github.com/pufferpanel/pufferd/shutdown"
	"github.com/pufferpanel/pufferd/version"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
)

type Run struct {
	Command
	run        bool
	runService bool
}

func (r *Run) Load() {
	r.runService = true
	flag.BoolVar(&r.run, "run", false, "Runs the daemon")
}

func (r *Run) ShouldRun() bool {
	return r.run
}

func (*Run) ShouldRunNext() bool {
	return false
}

func (r *Run) Run() error {
	err := config2.LoadConfig()

	if err != nil {
		return err
	}

	logging.Init()

	gin.SetMode(gin.ReleaseMode)
	logging.Info(version.Display)

	environments.LoadModules()
	programs.Initialize()

	if _, err = os.Stat(programs.TemplateFolder); os.IsNotExist(err) {
		logging.Info("No template directory found, creating")
		err = os.MkdirAll(programs.TemplateFolder, 0755)
		if err != nil {
			logging.Error("Error creating template folder", err)
		}

	}

	if _, err = os.Stat(programs.ServerFolder); os.IsNotExist(err) {
		logging.Info("No server directory found, creating")
		err = os.MkdirAll(programs.ServerFolder, 0755)
		if err != nil {
			logging.Error("Error creating server folder directory", err)
			return nil
		}
	}

	programs.LoadFromFolder()

	programs.InitService()

	for _, element := range programs.GetAll() {
		if element.IsEnabled() {
			element.GetEnvironment().DisplayToConsole("Daemon has been started\n")
			if element.IsAutoStart() {
				logging.Info("Queued server " + element.Id())
				element.GetEnvironment().DisplayToConsole("Server has been queued to start\n")
				programs.StartViaService(element)
			}
		}
	}

	defer recoverPanic()

	r.createHook()

	for r.runService {
		r.runServices()
	}

	shutdown.Shutdown()

	return nil
}

func (r *Run) runServices() {
	router := routing.ConfigureWeb()

	useHttps := false

	dataFolder := config.GetStringOrDefault("dataFolder", "data")
	httpsPem := filepath.Join(dataFolder, "https.pem")
	httpsKey := filepath.Join(dataFolder, "https.key")

	if _, err := os.Stat(httpsPem); os.IsNotExist(err) {
		logging.Warn("No HTTPS.PEM found in data folder, will use http instead")
	} else if _, err := os.Stat(httpsKey); os.IsNotExist(err) {
		logging.Warn("No HTTPS.KEY found in data folder, will use http instead")
	} else {
		useHttps = true
	}

	sftp.Run()

	web := config.GetStringOrDefault("web", config.GetStringOrDefault("webHost", "0.0.0.0")+":"+config.GetStringOrDefault("webPort", "5656"))

	logging.Infof("Starting web access on %s", web)
	var err error
	if useHttps {
		err = manners.ListenAndServeTLS(web, httpsPem, httpsKey, router)
	} else {
		err = manners.ListenAndServe(web, router)
	}
	if err != nil {
		logging.Error("Error starting web service", err)
	}
}

func (r *Run) createHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGPIPE)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logging.Errorf("Error: %+v\n%s", err, debug.Stack())
			}
		}()

		var sig os.Signal

		for sig != syscall.SIGTERM {
			sig = <-c
			switch sig {
			case syscall.SIGHUP:
				manners.Close()
				sftp.Stop()
				config.Load(config2.GetPath())
			case syscall.SIGPIPE:
				//ignore SIGPIPEs for now, we're somehow getting them and it's causing issues
			}
		}

		r.runService = false
		shutdown.CompleteShutdown()
	}()
}

func recoverPanic() {
	if rec := recover(); rec != nil {
		err := rec.(error)
		logging.Critical("Unhandled error", err)
	}
}
