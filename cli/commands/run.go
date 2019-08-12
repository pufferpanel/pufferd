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
	"github.com/braintree/manners"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/config"
	"github.com/pufferpanel/pufferd/environments"
	"github.com/pufferpanel/pufferd/programs"
	"github.com/pufferpanel/pufferd/routing"
	"github.com/pufferpanel/pufferd/sftp"
	"github.com/pufferpanel/pufferd/shutdown"
	"github.com/pufferpanel/pufferd/version"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		err := runRun()
		if err != nil {
			logging.Exception("error running", err)
		}
	},
}

var runService = true

func runRun() error {
	logging.Info(version.Display)

	environments.LoadModules()
	programs.Initialize()

	var err error

	if _, err = os.Stat(programs.ServerFolder); os.IsNotExist(err) {
		logging.Info("No server directory found, creating")
		err = os.MkdirAll(programs.ServerFolder, 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}

	programs.LoadFromFolder()

	programs.InitService()

	for _, element := range programs.GetAll() {
		if element.IsEnabled() {
			element.GetEnvironment().DisplayToConsole("Daemon has been started\n")
			if element.IsAutoStart() {
				logging.Info("Queued server %s", element.Id())
				element.GetEnvironment().DisplayToConsole("Server has been queued to start\n")
				programs.StartViaService(element)
			}
		}
	}

	defer recoverPanic()

	createHook()

	for runService && err == nil {
		err = runServices()
	}

	shutdown.Shutdown()

	return err
}

func runServices() error {
	router := routing.ConfigureWeb()

	useHttps := false

	dataFolder := config.Get().Data.BasePath
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

	web := config.Get().Listener.Web

	logging.Info("Starting web access on %s", web)
	var err error
	if useHttps {
		err = manners.ListenAndServeTLS(web, httpsPem, httpsKey, router)
	} else {
		err = manners.ListenAndServe(web, router)
	}

	return err
}

func createHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGPIPE)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logging.Error("%+v\n%s", err, debug.Stack())
			}
		}()

		var sig os.Signal

		for sig != syscall.SIGTERM {
			sig = <-c
			switch sig {
			case syscall.SIGHUP:
				manners.Close()
				sftp.Stop()
				config.LoadConfig()
			case syscall.SIGPIPE:
				//ignore SIGPIPEs for now, we're somehow getting them and it's causing issues
			}
		}

		runService = false
		shutdown.CompleteShutdown()
	}()
}

func recoverPanic() {
	if rec := recover(); rec != nil {
		err := rec.(error)
		logging.Critical("Unhandled error: %s", err.Error())
	}
}
