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

package main

import (
	"flag"
	"fmt"
	"github.com/pufferpanel/pufferd/environments"
	"github.com/pufferpanel/pufferd/install"
	"os"

	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/config"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/commands"
	"github.com/pufferpanel/pufferd/data"
	"github.com/pufferpanel/pufferd/programs"
	"github.com/pufferpanel/pufferd/routing"
	"github.com/pufferpanel/pufferd/sftp"
	"github.com/pufferpanel/pufferd/shutdown"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
)

var (
	VERSION = "nightly"
	GITHASH = "unknown"
)

var runService = true
var configPath string

func main() {
	var loggingLevel string
	var authRoot string
	var authToken string
	var runInstaller bool
	var version bool
	var license bool
	var shutdownPid int
	var runDaemon bool
	var reloadPid int
	flag.StringVar(&loggingLevel, "logging", "INFO", "Lowest logging level to display")
	flag.StringVar(&authRoot, "auth", "", "Base URL to the authorization server")
	flag.StringVar(&authToken, "token", "", "Authorization token")
	flag.BoolVar(&runInstaller, "install", false, "If installing instead of running")
	flag.BoolVar(&version, "version", false, "Get the version")
	flag.BoolVar(&license, "license", false, "View license")
	flag.StringVar(&configPath, "config", "config.json", "Path to config.json")
	flag.IntVar(&shutdownPid, "shutdown", 0, "PID to shut down")
	flag.IntVar(&reloadPid, "reload", 0, "PID to shut down")
	flag.BoolVar(&runDaemon, "run", false, "Runs the daemon")
	flag.Parse()

	versionString := fmt.Sprintf("pufferd %s (%s)", VERSION, GITHASH)

	if shutdownPid != 0 {
		logging.Info("Shutting down")
		commands.Shutdown(shutdownPid)
	}

	if reloadPid != 0 {
		logging.Info("Reloading")
		commands.Reload(reloadPid)
	}

	if version {
		fmt.Println(versionString)
	}

	if license {
		fmt.Println(data.LICENSE)
	}

	if license || version || shutdownPid != 0 || reloadPid != 0 {
		return
	}

	if !runInstaller {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultPath := "config.json"
			if runtime.GOOS == "linux" {
				defaultPath = "/etc/pufferd/config.json"
			}
			if _, err := os.Stat(defaultPath); err == nil {
				logging.Infof("No config passed, defaulting to %s", defaultPath)
				configPath = "/etc/pufferd/config.json"
			} else {
				logging.Error("Cannot find a config file!")
				return
			}
		}
	}

	if runInstaller {
		install.Install(configPath, authRoot, authToken)
	}

	if runInstaller || !runDaemon {
		return
	}

	config.Load(configPath)

	logging.SetLevelByString(loggingLevel)
	var defaultLogFolder = "logs"
	if runtime.GOOS == "linux" {
		defaultLogFolder = "/var/log/pufferd"
	}
	var logPath = config.GetStringOrDefault("logPath", defaultLogFolder)
	logging.SetLogFolder(logPath)
	logging.Init()
	gin.SetMode(gin.ReleaseMode)

	logging.Info(versionString)
	logging.Info("Logging set to " + loggingLevel)

	environments.LoadModules()
	programs.Initialize()

	if _, err := os.Stat(programs.TemplateFolder); os.IsNotExist(err) {
		logging.Info("No template directory found, creating")
		err = os.MkdirAll(programs.TemplateFolder, 0755)
		if err != nil {
			logging.Error("Error creating template folder", err)
		}
	}

	if _, err := os.Stat(programs.ServerFolder); os.IsNotExist(err) {
		logging.Info("No server directory found, creating")
		err = os.MkdirAll(programs.ServerFolder, 0755)
		if err != nil {
			logging.Error("Error creating server folder directory", err)
			return
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

	CreateHook()

	for runService {
		runServices()
	}

	shutdown.Shutdown()
}

func runServices() {
	defer recoverPanic()

	r := routing.ConfigureWeb()

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
		err = manners.ListenAndServeTLS(web, httpsPem, httpsKey, r)
	} else {
		err = manners.ListenAndServe(web, r)
	}
	if err != nil {
		logging.Error("Error starting web service", err)
		runService = false
	}
}

func CreateHook() {
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
				config.Load(configPath)
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
		logging.Critical("Unhandled error", err)
	}
}
