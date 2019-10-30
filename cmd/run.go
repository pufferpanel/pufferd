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

package main

import (
	"fmt"
	"github.com/braintree/manners"
	"github.com/pufferpanel/apufferi/v4"
	"github.com/pufferpanel/apufferi/v4/logging"
	"github.com/pufferpanel/pufferd/v2"
	"github.com/pufferpanel/pufferd/v2/environments"
	"github.com/pufferpanel/pufferd/v2/programs"
	"github.com/pufferpanel/pufferd/v2/routing"
	"github.com/pufferpanel/pufferd/v2/sftp"
	"github.com/pufferpanel/pufferd/v2/shutdown"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
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
	_ = pufferd.LoadConfig()

	var logPath = viper.GetString("data.logs")
	_ = logging.WithLogDirectory(logPath, logging.DEBUG, nil)

	logging.Info(pufferd.Display)

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
			element.GetEnvironment().DisplayToConsole(true, "Daemon has been started\n")
			if element.IsAutoStart() {
				logging.Info("Queued server %s", element.Id())
				element.GetEnvironment().DisplayToConsole(true, "Server has been queued to start\n")
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

	httpsPem := viper.GetString("listen.webCert")
	httpsKey := viper.GetString("listen.webKey")

	if _, err := os.Stat(httpsPem); os.IsNotExist(err) {
		logging.Warn("No HTTPS.PEM found in data folder, will use http instead")
	} else if _, err := os.Stat(httpsKey); os.IsNotExist(err) {
		logging.Warn("No HTTPS.KEY found in data folder, will use http instead")
	} else {
		useHttps = true
	}

	sftp.Run()

	web := viper.GetString("listen.web")

	logging.Info("Starting web access on %s", web)
	var err error
	if useHttps {
		err = manners.ListenAndServeTLS(web, httpsPem, httpsKey, router)
	} else {
		err = manners.ListenAndServe(web, router)
	}

	if runtime.GOOS != "windows" {
		go func() {
			file := viper.GetString("listen.socket")

			if file == "" || !strings.HasPrefix(file, "unix:") {
				return
			}

			file = strings.TrimPrefix(file, "unix:")

			err := os.Remove(file)
			if err != nil && !os.IsNotExist(err) {
				logging.Exception(fmt.Sprintf("Error deleting %s", file), err)
				return
			}

			listener, err := net.Listen("unix", file)
			defer apufferi.Close(listener)
			if err != nil {
				logging.Exception(fmt.Sprintf("Error listening on %s", file), err)
				return
			}

			err = os.Chmod(file, 0777)
			if err != nil {
				logging.Exception(fmt.Sprintf("Error listening on %s", file), err)
				return
			}

			logging.Info("Listening for socket requests")
			err = http.Serve(listener, router)
			if err != nil {
				logging.Exception(fmt.Sprintf("Error listening on %s", file), err)
				return
			}
		}()
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
				//manners.Close()
				//sftp.Stop()
				_ = pufferd.LoadConfig()
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
		fmt.Printf("CRITICAL: %s", err.Error())
		logging.Critical("Unhandled error: %s", err.Error())
	}
}
