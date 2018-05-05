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

const SRCDS = `{
  "pufferd": {
    "type": "srcds",
    "install": {
      "commands": [
        {
          "files": "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz",
          "type": "download"
        },
        {
          "target": "steamcmd",
          "type": "mkdir"
        },
        {
          "commands": [
            "tar --no-same-owner -xzvf steamcmd_linux.tar.gz -C steamcmd",
            "steamcmd/steamcmd.sh +login anonymous +force_install_dir ${rootdir} +app_update ${appid} +quit",
            "mkdir -p .steam/sdk32",
            "cp steamcmd/linux32/steamclient.so .steam/sdk32/steamclient.so"
          ],
          "type": "command"
        }
      ]
    },
    "run": {
      "stop": "exit",
      "pre": [],
      "post": [],
      "arguments": [
      	"+ip",
      	"${ip}",
      	"+port",
      	"${port}",
      	"-game ${gametype}",
      	"-console",
        "+map ${map}",
      	"-norestart"
      ],
      "program": "./srcds_run"
    },
    "environment": {
      "type": "tty"
    },
    "data": {
      "appid": {
        "value": "232250",
        "required": true,
        "desc": "App ID",
        "display": "Application ID",
        "internal": false,
        "type": "integer"
      },
      "gametype": {
        "value": "tf",
        "required": false,
        "desc": "Game Type",
        "display": "tf, csgo, etc.",
        "internal": false
      },
      "map": {
      	"value": "ctf_2fort",
      	"required": false,
      	"desc": "Map",
      	"display": "Map to load",
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
        "internal": false,
        "type": "integer"
      }
    }
  }
}`

const TF2 = `{
  "pufferd": {
    "type": "srcds",
    "display": "Team Fortress 2",
    "install": {
      "commands": [
        {
          "files": "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz",
          "type": "download"
        },
        {
          "target": "steamcmd",
          "type": "mkdir"
        },
        {
          "commands": [
            "tar --no-same-owner -xzvf steamcmd_linux.tar.gz -C steamcmd",
            "steamcmd/steamcmd.sh +login anonymous +force_install_dir ${rootdir} +app_update 232250 +quit",
            "mkdir -p .steam/sdk32",
            "cp steamcmd/linux32/steamclient.so .steam/sdk32/steamclient.so"
          ],
          "type": "command"
        }
      ]
    },
    "run": {
      "stop": "exit",
      "pre": [],
      "post": [],
      "arguments": [
      	"+ip",
      	"${ip}",
      	"+port",
      	"${port}",
      	"-game tf",
      	"-console",
        "+map ${map}",
      	"-norestart"
      ],
      "program": "./srcds_run"
    },
    "environment": {
      "type": "tty"
    },
    "data": {
      "map": {
      	"value": "ctf_2fort",
      	"required": true,
      	"desc": "TF2 Map",
      	"display": "Team Fortess 2 Map to load",
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
        "internal": false,
        "type": "integer"
      }
    }
  }
}`

const CSGO = `{
  "pufferd": {
    "type": "srcds",
    "display": "Counter Strike: Global Offensive",
    "install": {
      "commands": [
        {
          "files": "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz",
          "type": "download"
        },
        {
          "target": "steamcmd",
          "type": "mkdir"
        },
        {
          "commands": [
            "tar --no-same-owner -xzvf steamcmd_linux.tar.gz -C steamcmd",
            "steamcmd/steamcmd.sh +login anonymous +force_install_dir ${rootdir} +app_update 740 +quit",
            "mkdir -p .steam/sdk32",
            "cp steamcmd/linux32/steamclient.so .steam/sdk32/steamclient.so"
          ],
          "type": "command"
        }
      ]
    },
    "run": {
      "stop": "exit",
      "pre": [],
      "post": [],
      "arguments": [
        "-game",
        "csgo",
        "-usercon",
        "-strictportbind",
        "-ip",
        "${ip}",
        "-port",
        "${port}",
        "+clientport",
        "${clientport}",
        "+tv_port",
        "${tvport}",
        "+map",
        "${map}",
        "-maxplayers_override",
        "${maxplayers}",
        "+sv_setsteamaccount",
        "${gslt}",
        "-tickrate",
        "${tickrate}",
        "+mapgroup",
        "${mapgroup}",
        "+game_mode",
        "${gamemode}",
        "+game_type",
        "${gametype}",
        "-net_port_try 1",
        "-norestart"
      ],
      "program": "./srcds_run"
    },
    "environment": {
      "type": "tty"
    },
    "data": {
      "map": {
        "value": "de_dust2",
        "required": true,
        "desc": "Map to load by default",
        "display": "Map",
        "internal": false
      },
      "maxplayers": {
        "value": "20",
        "required": true,
        "desc": "Max players allowed on server",
        "display": "Max Players",
        "internal": false,
        "type": "integer"
      },
      "gslt": {
        "value": "GSLTTOKEN",
        "required": true,
        "desc": "It is required to register your server, otherwise it can only be played on LAN. Visit http://steamcommunity.com/dev/managegameservers to get a token",
        "display": "GSLT Token",
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
        "internal": false,
        "type": "integer"
      },
      "clientport": {
        "value": "27016",
        "required": true,
        "desc": "What port to bind the clientport to",
        "display": "Client Port",
        "internal": false,
        "type": "integer"
      },
      "tvport": {
        "value": "27017",
        "required": true,
        "desc": "What port to bind the GOTV to",
        "display": "GOTV Port",
        "internal": false,
        "type": "integer"
      },
      "tickrate": {
        "value": "64",
        "required": true,
        "desc": "Server tick rate",
        "display": "Tickrate",
        "internal": false,
        "type": "integer"
      },
      "mapgroup": {
        "value": "mg_active",
        "required": true,
        "desc": "(Advanced) Map group",
        "display": "Map group",
        "internal": false,
        "type": "integer"
      },
      "gamemode": {
        "value": "0",
        "required": true,
        "desc": "(Advanced) Gamemode",
        "display": "Gamemode",
        "internal": false,
        "type": "integer"
      },
      "gametype": {
        "value": "0",
        "required": true,
        "desc": "(Advanced) Gametype",
        "display": "Gametype",
        "internal": false,
        "type": "integer"
      }
    }
  }
}`

const GMOD = `{
  "pufferd": {
    "type": "srcds",
    "display": "Garry's Mod",
    "install": {
      "commands": [
        {
          "files": "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz",
          "type": "download"
        },
        {
          "target": "steamcmd",
          "type": "mkdir"
        },
        {
          "commands": [
            "tar --no-same-owner -xzvf steamcmd_linux.tar.gz -C steamcmd",
            "steamcmd/steamcmd.sh +login anonymous +force_install_dir ${rootdir} +app_update 4020 +quit",
            "mkdir -p .steam/sdk32",
            "cp steamcmd/linux32/steamclient.so .steam/sdk32/steamclient.so"
          ],
          "type": "command"
        }
      ]
    },
    "run": {
      "stop": "exit",
      "pre": [],
      "post": [],
      "arguments": [
        "+ip",
        "${ip}",
        "+port",
        "${port}",
        "-game",
        "garrysmod",
        "-console",
        "+map",
        "${map}",
        "+maxplayers",
        "${maxplayers}",
        "-norestart"
      ],
      "program": "./srcds_run"
    },
    "environment": {
      "type": "tty"
    },
    "data": {
      "map": {
        "value": "gm_construct",
        "required": true,
        "desc": "Map to load by default",
        "display": "Map",
        "internal": false
      },
      "maxplayers": {
        "value": "20",
        "required": true,
        "desc": "Max players allowed on server",
        "display": "Max Players",
        "internal": false,
        "type": "integer"
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
        "internal": false,
        "type": "integer"
      }
    }
  }
}`
