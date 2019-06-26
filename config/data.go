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

type Base struct {
	Auth     Auth     `json:"auth"`
	Listener Listener `json:"listen"`
	Console  Console  `json:"console"`
	Data     Data     `json:"data"`
}

type Data struct {
	ServerFolder             string `json:"servers"`
	TemplateFolder           string `json:"templates"`
	CacheFolder              string `json:"cache"`
	ModuleFolder             string `json:"modules"`
	CrashLimit               int    `json:"crashLimit"`
	BasePath                 string `json:"base"`
	LogFolder                string `json:"logs"`
	MaxWebsocketDownloadSize int64  `json:"maxWSDownloadSize"`
}

type Console struct {
	Buffer  int  `json:"buffer"`
	Forward bool `json:"forward"`
}

type Auth struct {
	AuthURL      string `json:"authUrl"`
	InfoURL      string `json:"infoUrl"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type Listener struct {
	Web     string `json:"web"`
	SFTP    string `json:"sftp"`
	SFTPKey string `json:"serverKey"`
}
