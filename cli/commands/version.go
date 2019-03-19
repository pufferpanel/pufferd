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
	"fmt"

	"github.com/pufferpanel/pufferd/version"
)

type Version struct {
	Command
	version bool
}

func (v *Version) Load() {
	flag.BoolVar(&v.version, "version", false, "Get the version")
}

func (v *Version) ShouldRun() bool {
	return v.version
}

func (*Version) ShouldRunNext() bool {
	return false
}

func (v *Version) Run() error {
	fmt.Printf(version.Display)
	return nil
}
