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

package templates

const FACTORIO = `{
  "pufferd": {
    "type": "factorio",
    "display":"Factorio",
    "install": {
      "commands": [
        {
          "commands": [
          "curl -L -o factorio.tar.xz https://www.factorio.com/get-download/${version}/headless/linux64",
            "mkdir factorio",
            "tar --no-same-owner -xvf factorio.tar.xz",
            "cp factorio/data/server-settings.example.json factorio/data/server-settings.json",
            "./factorio/bin/x64/factorio --create saves/default.zip"
          ],
          "type": "command"
        }
      ]
    },
    "run": {
      "stop": "/quit",
      "pre": [],
      "post": [],
      "arguments": [
        "--port",
        "${port}",
        "--bind",
        "${ip}",
        "--start-server",
        "${save}",
        "--server-settings",
        "${server-settings}"
      ],
      "program": "./factorio/bin/x64/factorio"
    },
    "environment": {
      "type": "tty"
    },
    "data": {
      "version": {
        "value": "0.15.19",
        "required": true,
        "desc": "Version",
        "display": "Version to Install",
        "internal": true
      },
      "save": {
        "value": "saves/default.zip",
        "required": true,
        "desc": "Save File to Use",
        "display": "Save File to Use",
        "internal": false
      },
      "ip": {
        "value": "0.0.0.0",
        "required": true,
        "desc": "What IP to bind the server to",
        "display": "IP",
        "internal": false
      },
      "port": {
        "value": "27015",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false
      },
      "server-settings": {
        "value": "factorio/data/server-settings.json",
        "required": true,
        "desc": "Server Settings File Location",
        "display": "Server Settings JSON",
        "internal": false
      }
    }
  }
}`
