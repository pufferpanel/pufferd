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

package config

import (
	"github.com/spf13/viper"
	"runtime"
	"strings"
)

var path = "config.json"

func init() {
	//env configuration
	viper.SetEnvPrefix("PUFFERPANEL")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	//defaults we can set at this point in time
	viper.SetDefault("console.buffer", 50)
	viper.SetDefault("console.forward", false)

	viper.SetDefault("listen.web", "0.0.0.0:5656")
	viper.SetDefault("listen.socket", "/var/run/pufferd.sock")
	viper.SetDefault("listen.webCert", "https.pem")
	viper.SetDefault("listen.webKey", "https.key")
	viper.SetDefault("listen.sftp", "0.0.0.0:5657")
	viper.SetDefault("listen.sftpKey", "sftp.key")

	viper.SetDefault("auth.publicKey", "panel.pem")
	if runtime.GOOS == "windows" {
		viper.SetDefault("auth.url", "http://localhost:8080/oauth2/token")
	} else {
		//TODO: Support unix sockets for authorization endpoint
		//viper.SetDefault("auth.url", "/var/run/pufferpanel.sock")
		viper.SetDefault("auth.url", "http://localhost:8080/oauth2/token")
	}

	viper.SetDefault("auth.clientId", "")
	viper.SetDefault("auth.clientSecret", "")

	viper.SetDefault("data.cache", "cache")
	viper.SetDefault("data.servers", "servers")
	viper.SetDefault("data.modules", "modules")
	viper.SetDefault("data.logs", "logs")
	viper.SetDefault("data.crashLimit", 3)
	viper.SetDefault("data.maxWSDownloadSize", int64(1024*1024*20)) //1024 bytes (1KB) * 1024 (1MB) * 50 (50MB))
}

func LoadConfig(p string) error {
	if p != "" {
		path = p
	}

	if path != "" {
		viper.SetConfigName("config")

		if runtime.GOOS != "windows" {
			viper.AddConfigPath("/etc/pufferd/")
		}

		viper.AddConfigPath(".")
	} else {
		viper.SetConfigFile(path)
		path = p
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			//this is just a missing config, since ENV is supported, ignore
		} else {
			return err
		}
	}

	return nil
}
