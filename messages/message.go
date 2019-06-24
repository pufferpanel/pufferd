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

package messages

type Message interface {
	Key() string
}

type Transmission struct {
	Message Message `json:"data"`
	Type    string  `json:"type"`
}

type StatMessage struct {
	Memory float64 `json:"memory"`
	Cpu    float64 `json:"cpu"`
}

type ConsoleMessage struct {
	Logs []string `json:"logs"`
}

type PingMessage struct {
}

func (m StatMessage) Key() string {
	return "stat"
}

func (m ConsoleMessage) Key() string {
	return "console"
}

func (m PingMessage) Key() string {
	return "ping"
}

