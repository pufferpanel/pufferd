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
	"io/ioutil"
	"os"
	"path/filepath"

	"fmt"

	"net/http"
	"strings"

	"github.com/braintree/manners"
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/config"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/data"
	"github.com/pufferpanel/pufferd/data/templates"
	"github.com/pufferpanel/pufferd/install"
	"github.com/pufferpanel/pufferd/migration"
	"github.com/pufferpanel/pufferd/programs"
	"github.com/pufferpanel/pufferd/routing"
	"github.com/pufferpanel/pufferd/sftp"
	"github.com/pufferpanel/pufferd/shutdown"
	"github.com/pufferpanel/pufferd/uninstaller"
	"runtime"
)

var (
	VERSION      = "nightly"
	MAJORVERSION = "nightly"
	GITHASH      = "unknown"
)

func main() {
	var loggingLevel string
	var authRoot string
	var authToken string
	var runInstaller bool
	var version bool
	var license bool
	var regenerate bool
	var migrate bool
	var uninstall bool
	var configPath string
	var pid int
	var runDaemon bool
	flag.StringVar(&loggingLevel, "logging", "INFO", "Lowest logging level to display")
	flag.StringVar(&authRoot, "auth", "", "Base URL to the authorization server")
	flag.StringVar(&authToken, "token", "", "Authorization token")
	flag.BoolVar(&runInstaller, "install", false, "If installing instead of running")
	flag.BoolVar(&version, "version", false, "Get the version")
	flag.BoolVar(&license, "license", false, "View license")
	flag.BoolVar(&regenerate, "regenerate", false, "Regenerate pufferd templates")
	flag.BoolVar(&migrate, "migrate", false, "Migrate Scales data to pufferd")
	flag.BoolVar(&uninstall, "uninstall", false, "Uninstall pufferd")
	flag.StringVar(&configPath, "config", "config.json", "Path to pufferd config.json")
	flag.IntVar(&pid, "shutdown", 0, "PID to shut down")
	flag.BoolVar(&runDaemon, "run", false, "Runs the daemon")
	flag.Parse()

	versionString := fmt.Sprintf("pufferd %s (%s)", VERSION, GITHASH)

	if pid != 0 {
		logging.Info("Shutting down")
		shutdown.Command(pid)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) && !runInstaller {
		if _, err := os.Stat("/etc/pufferd/config.json"); err == nil {
			logging.Info("No config passed, defaulting to /etc/pufferd/config.json")
			configPath = "/etc/pufferd/config.json"
		} else {
			logging.Error("Cannot find a config file!")
			return
		}
	}

	if uninstall {
		fmt.Println("This option will UNINSTALL pufferd, are you sure? Please enter \"yes\" to proceed [no]")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "yes" || strings.ToLower(response) == "y" {
			if os.Geteuid() != 0 {
				logging.Error("To uninstall pufferd you need to have sudo or root privileges")
			} else {
				config.Load(configPath)
				uninstaller.StartProcess()
				logging.Info("pufferd is now uninstalled.")
			}
		} else {
			logging.Info("Uninstall process aborted")
			logging.Info("Exiting")
		}
		return
	}

	if version || !runDaemon {
		os.Stdout.WriteString(versionString + "\r\n")
	}

	if license {
		os.Stdout.WriteString(data.LICENSE + "\r\n")
	}

	if regenerate {
		config.Load(configPath)
		programs.Initialize()

		if _, err := os.Stat(programs.TemplateFolder); os.IsNotExist(err) {
			logging.Info("No template directory found, creating")
			err = os.MkdirAll(programs.TemplateFolder, 0755)
			if err != nil {
				logging.Error("Error creating template folder", err)
			}
		}
		// Overwrite existing templates
		templates.CopyTemplates()
		logging.Info("Templates regenerated")
	}

	if migrate {
		config.Load(configPath)
		migration.MigrateFromScales()
	}

	if license || version || regenerate || migrate || pid != 0 {
		return
	}

	config.Load(configPath)

	logging.SetLevelByString(loggingLevel)
	var defaultLogFolder = "logs"
	if runtime.GOOS == "linux" {
		defaultLogFolder = "/var/log/pufferd"
	}
	var logPath = config.GetOrDefault("logPath", defaultLogFolder)
	logging.SetLogFolder(logPath)
	logging.Init()
	gin.SetMode(gin.ReleaseMode)

	logging.Info(versionString)
	logging.Info("Logging set to " + loggingLevel)

	if runInstaller {
		install.Install(configPath, authRoot, authToken)
	}

	if runInstaller || !runDaemon {
		return
	}

	programs.Initialize()

	if _, err := os.Stat(programs.TemplateFolder); os.IsNotExist(err) {
		logging.Info("No template directory found, creating")
		err = os.MkdirAll(programs.TemplateFolder, 0755)
		if err != nil {
			logging.Error("Error creating template folder", err)
		}

	}
	if files, _ := ioutil.ReadDir(programs.TemplateFolder); len(files) == 0 {
		logging.Info("Templates being copied to " + programs.TemplateFolder)
		templates.CopyTemplates()
	}

	if _, err := os.Stat(programs.ServerFolder); os.IsNotExist(err) {
		logging.Info("No server directory found, creating")
		os.MkdirAll(programs.ServerFolder, 0755)
	}

	programs.LoadFromFolder()

	programs.InitService()

	for _, element := range programs.GetAll() {
		if element.IsEnabled() && element.IsAutoStart() {
			logging.Info("Queued server " + element.Id())
			programs.StartViaService(element)
		}
	}

	r := routing.ConfigureWeb()

	useHttps := false

	dataFolder := config.GetOrDefault("datafolder", "data")
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

	//check if there's an update
	if config.GetOrDefault("update-check", "true") == "true" {
		go func() {
			url := "https://dl.pufferpanel.com/pufferd/" + MAJORVERSION + "/version.txt"
			logging.Debug("Checking for updates using " + url)
			resp, err := http.Get(url)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			onlineVersion, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}
			if string(onlineVersion) != GITHASH {
				logging.Infof("DL server reports a different hash than this version, an update may be available")
				logging.Infof("Installed: %s", GITHASH)
				logging.Infof("Online: %s", onlineVersion)
			}
		}()
	}

	web := config.GetOrDefault("web", config.GetOrDefault("webhost", "0.0.0.0")+":"+config.GetOrDefault("webport", "5656"))

	shutdown.CreateHook()

	logging.Infof("Starting web access on %s", web)
	var err error
	if useHttps {
		err = manners.ListenAndServeTLS(web, httpsPem, httpsKey, r)
	} else {
		err = manners.ListenAndServe(web, r)
	}
	if err != nil {
		logging.Error("Error starting web service", err)
	}

	shutdown.Shutdown()
}
