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
	"errors"
	"flag"
	"fmt"
	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/pufferd/config"
	"github.com/pufferpanel/pufferd/data"
	"io/ioutil"
	"strings"
)

type Install struct {
	Command
	install      bool
	authUrl      string
	clientId     string
	clientSecret string
}

func (i *Install) Load() {
	flag.BoolVar(&i.install, "install", false, "Install the daemon")
	flag.StringVar(&i.authUrl, "authUrl", "", "Base URL to authorization server")
	flag.StringVar(&i.clientId, "clientId", "", "Client ID for authorization server")
	flag.StringVar(&i.clientSecret, "clientSecret", "", "Client secret for authorization server")
}

func (i *Install) ShouldRun() bool {
	return i.install
}

func (*Install) ShouldRunNext() bool {
	return false
}

func (i *Install) Run() error {
	if i.authUrl == "" {
		return errors.New("authUrl must be provided")
	}

	if i.clientId == "" {
		return errors.New("clientId must be provided")
	}

	if i.clientSecret == "" {
		return errors.New("clientSecret must be provided")
	}

	cfgData := data.CONFIG

	replacements := make(map[string]interface{})
	replacements["authUrl"] = strings.TrimSuffix(i.authUrl, "/")
	replacements["clientId"] = i.clientId
	replacements["clientSecret"] = i.clientSecret

	configData := []byte(common.ReplaceTokens(cfgData, replacements))

	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, configData, "", "  ")
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
