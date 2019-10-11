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

package cli

import (
	"fmt"
	"github.com/pufferpanel/apufferi/v3/logging"
	"github.com/pufferpanel/pufferd/v2/cli/commands"
	"github.com/pufferpanel/pufferd/v2/config"
	"github.com/pufferpanel/pufferd/v2/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "pufferd",
	Short: "pufferpanel daemon",
}

var configPath = "config.json"
var loggingLevel = "INFO"

func init() {
	cobra.OnInitialize(load)

	rootCmd.AddCommand(
		commands.LicenseCmd,
		commands.ShutdownCmd,
		commands.RunCmd,
		commands.ReloadCmd,
		commands.MigrateCmd)

	rootCmd.PersistentFlags().StringVar(&configPath, "config", configPath, "Path to the config to use")
	rootCmd.PersistentFlags().StringVar(&loggingLevel, "logging", loggingLevel, "Logging level to print to stdout")
	rootCmd.SetVersionTemplate(version.Display)
}

func Execute() error {
	return rootCmd.Execute()
}

func load() {
	err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %s", err)
	}

	level := logging.GetLevel(loggingLevel)
	if level == nil {
		level = logging.INFO
	}

	logging.SetLevel(os.Stdout, level)

	var logPath = viper.GetString("data.logs")
	_ = logging.WithLogDirectory(logPath, logging.DEBUG, nil)
}
