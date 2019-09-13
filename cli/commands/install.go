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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pufferpanel/apufferi/v3/logging"
	"github.com/pufferpanel/pufferd/v2/config"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strings"
)

var authUrl string
var clientId string
var clientSecret string

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		err := runInstall()
		if err != nil {
			logging.Exception("error running", err)
		}
	},
}

func init() {
	InstallCmd.Flags().StringVar(&authUrl, "authUrl", "", "Base URL to authorization server")
	InstallCmd.Flags().StringVar(&clientId, "clientId", "", "Client ID for authorization server")
	InstallCmd.Flags().StringVar(&clientSecret, "clientSecret", "", "Client secret for authorization server")
	InstallCmd.MarkFlagRequired("authUrl")
	InstallCmd.MarkFlagRequired("clientId")
	InstallCmd.MarkFlagRequired("clientSecret")
}

func runInstall() error {
	cfgData := config.Get()

	url := strings.TrimSuffix(authUrl, "/")
	cfgData.Auth.AuthURL = url + "/oauth2/token"
	cfgData.Auth.InfoURL = url + "/oauth2/info"
	cfgData.Auth.ClientID = clientId
	cfgData.Auth.ClientSecret = clientSecret

	configData, err := json.Marshal(cfgData)
	if err != nil {
		return err
	}

	var prettyJson bytes.Buffer
	err = json.Indent(&prettyJson, configData, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(config.GetPath(), prettyJson.Bytes(), 0664)

	if err != nil {
		return err
	}

	fmt.Println("Config saved")
	return nil
}
