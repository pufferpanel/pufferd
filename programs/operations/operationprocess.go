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

package operations

import (
	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/environments"
	"github.com/pufferpanel/pufferd/programs/operations/ops"
)

func GenerateProcess(directions []map[string]interface{}, environment environments.Environment, dataMapping map[string]interface{}) OperationProcess {
	datamap := make(map[string]interface{})
	for k, v := range dataMapping {
		datamap[k] = v
	}
	datamap["rootdir"] = environment.GetRootDirectory()
	operationList := make([]ops.Operation, 0)
	for _, mapping := range directions {
		switch mapping["type"] {
		case "command":
			for _, element := range common.ToStringArray(mapping["commands"]) {
				operationList = append(operationList, &ops.Command{Command: common.ReplaceTokens(element, datamap), Environment: environment})
			}
		case "download":
			for _, element := range common.ToStringArray(mapping["files"]) {
				operationList = append(operationList, &ops.Download{File: common.ReplaceTokens(element, datamap), Environment: environment})
			}
		case "move":
			source := mapping["source"].(string)
			target := mapping["target"].(string)
			operationList = append(operationList, &ops.Move{SourceFile: source, TargetFile: target, Environment: environment})
		case "mkdir":
			target := mapping["target"].(string)
			operationList = append(operationList, &ops.Mkdir{TargetFile: target, Environment: environment})
		case "writefile":
			text := mapping["text"].(string)
			target := mapping["target"].(string)
			operationList = append(operationList, &ops.WriteFile{TargetFile: target, Environment: environment, Text: common.ReplaceTokens(text, datamap)})
		}
	}
	return OperationProcess{processInstructions: operationList}
}

type OperationProcess struct {
	processInstructions []ops.Operation
}

func (p *OperationProcess) Run() (err error) {
	for p.HasNext() {
		err = p.RunNext()
		if err != nil {
			logging.Error("Error running process: ", err)
			break
		}
	}
	return
}

func (p *OperationProcess) RunNext() error {
	var op ops.Operation
	op, p.processInstructions = p.processInstructions[0], p.processInstructions[1:]
	err := op.Run()
	return err
}

func (p *OperationProcess) HasNext() bool {
	return len(p.processInstructions) != 0 && p.processInstructions[0] != nil
}
