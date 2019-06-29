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
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"runtime"
)

var filePath string
var global = Base{}

func init() {
	var baseDataPath string
	var logPath string

	if runtime.GOOS == "windows" {
		filePath = "config.json"
		baseDataPath = "data"
		logPath = "logs"
	} else {
		filePath = "/etc/pufferd/config.json"
		baseDataPath = "/var/lib/pufferd"
		logPath = "/var/log/pufferd"
	}

	global.Console = Console{
		Buffer:  50,
		Forward: false,
	}

	global.Auth = Auth{
	}

	global.Listener = Listener{
		Web:     "0.0.0.0:5656",
		SFTP:    "0.0.0.0:5657",
		SFTPKey: path.Join(baseDataPath, "server.key"),
	}

	global.Data = Data{
		CacheFolder:    path.Join(baseDataPath, "cache"),
		ServerFolder:   path.Join(baseDataPath, "servers"),
		TemplateFolder: path.Join(baseDataPath, "templates"),
		ModuleFolder:   path.Join(baseDataPath, "modules"),
		BasePath:       baseDataPath,
		LogFolder:      logPath,
		CrashLimit:     3,
		MaxWebsocketDownloadSize: 1024 * 1024 * 20, //1024 bytes (1KB) * 1024 (1MB) * 50 (50MB)
	}
}

func Get() Base {
	return global
}

func SetPath(newPath string) {
	//only change path if the new path is actually valid
	if newPath != "" {
		filePath = newPath
	}
}

func GetPath() string {
	return filePath
}

func LoadConfig() error {
	file, err := ioutil.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = json.Unmarshal(file, &global)
	return err
}
