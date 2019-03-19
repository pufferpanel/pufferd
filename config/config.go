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
	"github.com/pufferpanel/apufferi/config"
	"os"
	"runtime"
)

var path string

func init() {
	if runtime.GOOS == "windows" {
		path = "config.json"
	} else {
		path = "/etc/pufferd/config.json"
	}
}

func SetPath(newPath string) {
	//only change path if the new path is actually valid
	if newPath != "" {
		path = newPath
	}
}

func GetPath() string {
	return path
}

func LoadConfig() error {
	if _, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	config.Load(path)
	return nil
}
