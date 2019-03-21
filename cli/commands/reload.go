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
	"errors"
	"flag"
	"github.com/pufferpanel/apufferi/cli"
	"os"
	"syscall"
)

type Reload struct {
	cli.Command
	pid int
}

func (r *Reload) Load() {
	flag.IntVar(&r.pid, "reload", 0, "PID to reload")
}

func (r *Reload) ShouldRun() bool {
	return r.pid != 0
}

func (*Reload) ShouldRunNext() bool {
	return false
}

func (r *Reload) Run() error {
	proc, err := os.FindProcess(r.pid)
	if err != nil || proc == nil {
		if err == nil && proc == nil {
			err = errors.New("no process found")
		}
		return err
	}
	return proc.Signal(syscall.Signal(1))
}
