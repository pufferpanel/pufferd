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

const Arma3 = `{
  "pufferd": {
    "type": "srcds",
    "install": {
      "commands": [
        {
          "files": "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz",
          "type": "download"
        },
        {
          "commands": [
            "mkdir steamcmd",
            "tar --no-same-owner -xzvf steamcmd_linux.tar.gz -C steamcmd",
            "steamcmd/steamcmd.sh +login ${steam_user} ${steam_password} +force_install_dir ${rootdir} +app_update 233780 +quit",
            "mkdir -p .steam/sdk32",
            "cp steamcmd/linux32/steamclient.so .steam/sdk32/steamclient.so"
          ],
          "type": "command"
        }
      ]
    },
    "run": {
      "stop": "",
      "pre": [],
      "post": [],
      "arguments": [
        "-ip=${ip}",
        "-port=${port}"
      ],
      "program": "./arma3server"
    },
    "environment": {
      "type": "tty"
    },
    "data": {
      "steam_user": {
        "value": "anonymous",
        "required": true,
        "desc": "Username for steam login",
        "display": "Steam User",
        "internal": false
      },
      "steam_password": {
        "value": "",
        "desc": "Password for steam login",
        "display": "Steam Password",
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
        "value": "2302",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      }
    }
  }
}`
