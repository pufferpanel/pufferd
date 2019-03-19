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
	"flag"
	"github.com/pufferpanel/pufferd/config"
)

type Config struct {
	Command
	path string
}

func (c *Config) Load() {
	flag.StringVar(&c.path, "config", "", "Path to config.json")
}

func (*Config) ShouldRun() bool {
	return true
}

func (*Config) ShouldRunNext() bool {
	return true
}

func (c *Config) Run() error {
	config.SetPath(c.path)
	return nil
}
