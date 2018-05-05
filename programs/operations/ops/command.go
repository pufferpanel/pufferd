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

package ops

import (
	"fmt"
	"strings"

	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/environments"
)

type Command struct {
	Command     string
	Environment environments.Environment
}

func (c *Command) Run() error {
	logging.Debugf("Executing command: %s", c.Command)
	c.Environment.DisplayToConsole(fmt.Sprintf("Executing: %s\n", c.Command))
	parts := strings.Split(c.Command, " ")
	cmd := parts[0]
	args := parts[1:]
	env := make(map[string]string)
	_, err := c.Environment.Execute(cmd, args, env, nil)
	return err
}
