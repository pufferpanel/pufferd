/*
 Copyright 2018 Padduck, LLC

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

package programs

type ProgramTemplate struct {
	Core ProgramTemplateData `json:"pufferd"`
}

type ProgramTemplateData struct {
	ProgramData
	SupportedEnvironments map[string]interface{} `json:"supportedEnvironments,omitEmpty"`
}

func (pt ProgramTemplate) Create(environment map[string]interface{}) Program {
	newPrg := &ProgramData{}
	newPrg.CopyFrom(&pt.Core.ProgramData)
	newPrg.EnvironmentData = environment

	return newPrg
}
