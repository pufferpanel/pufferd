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

const Bungeecord = `{
  "pufferd": {
    "type": "java",
    "display": "Bungeecord - Minecraft",
    "install": {
      "commands": [
        {
          "files": "http://ci.md-5.net/job/BungeeCord/lastSuccessfulBuild/artifact/bootstrap/target/BungeeCord.jar",
          "type": "download"
        },
        {
          "type": "writefile",
          "text": "listeners:\n- query_port: 25577\n  host: ${ip}:${port}\n  motd: '${motd}'\n",
          "target": "config.yml"
        }
      ]
    },
    "run": {
      "stop": "end",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
       "-Xms${memory}M",
      	"-jar",
      	"BungeeCord.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "standard"
    },
    "data": {
      "memory": {
      	"value": "512",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the BungeeCord server. It is generally recommended to allocated 512MB RAM for every 500 players. Please keep in mind that additional plugins may impact the amount needed in your case.",
      	"display": "Memory (MB)",
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
        "value": "25577",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "motd": {
        "value": "A BungeeCord Server &9hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const Spigot = `{
  "pufferd": {
    "type": "java",
    "display": "Spigot - Minecraft",
    "install": {
      "commands": [
        {
          "files": "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar",
          "type": "download"
        },
        {
          "commands": [
            "java -jar BuildTools.jar --rev ${version}"
          ],
          "type": "command"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "source": "spigot-*.jar",
          "target": "server.jar",
          "type": "move"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
        "-Djline.terminal=jline.UnsupportedTerminal",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "standard"
    },
    "data": {
      "version": {
      	"value": "latest",
      	"required": true,
      	"desc": "Version of Minecraft to install",
      	"display": "Version",
      	"internal": false
      },
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you agree to the Minecraft EULA?",
        "display": "EULA Agreement (true/false)",
        "internal": false,
        "type": "boolean"
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const CraftbukkitBySpigot = `{
  "pufferd": {
    "type": "java",
    "display": "CraftBukkit by Spigot - Minecraft",
    "install": {
      "commands": [
        {
          "files": "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar",
          "type": "download"
        },
        {
          "commands": [
            "java -jar BuildTools.jar --rev ${version}"
          ],
          "type": "command"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "source": "craftbukkit-*.jar",
          "target": "server.jar",
          "type": "move"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "standard"
    },
    "data": {
      "version": {
      	"value": "latest",
      	"required": true,
      	"desc": "Version of Minecraft to install",
      	"display": "Version",
      	"internal": false
      },
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you (or the server owner) agree to the <a href='https://account.mojang.com/documents/minecraft_eula'>Minecraft EULA?</a>",
        "display": "EULA Agreement",
        "internal": false,
        "type": "boolean"
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const Vanilla = `{
  "pufferd": {
    "type": "java",
    "display": "Vanilla - Minecraft",
    "install": {
      "commands": [
        {
          "type": "download",
          "files": "https://s3.amazonaws.com/Minecraft.Download/versions/${version}/minecraft_server.${version}.jar"
        },
        {
          "source": "minecraft_server.*.jar",
          "target": "server.jar",
          "type": "move"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "standard"
    },
    "data": {
      "version": {
      	"value": "1.12.2",
      	"required": true,
      	"desc": "Version of Minecraft to install",
      	"display": "Version",
      	"internal": false
      },
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you (or the server owner) agree to the <a href='https://account.mojang.com/documents/minecraft_eula'>Minecraft EULA?</a>",
        "display": "EULA Agreement",
        "internal": false,
        "type": "boolean"
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const Forge = `{
  "pufferd": {
    "type": "java",
    "display": "MinecraftForge - Minecraft",
    "install": {
      "commands": [
        {
          "type": "download",
          "files": "http://files.minecraftforge.net/maven/net/minecraftforge/forge/${version}/forge-${version}-installer.jar"
        },
        {
          "source": "forge-*.jar",
          "target": "installer.jar",
          "type": "move"
        },
        {
          "commands": [
            "java -jar installer.jar --installServer"
          ],
          "type": "command"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        },
        {
          "source": "forge-*-universal.jar",
          "target": "server.jar",
          "type": "move"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "standard"
    },
    "data": {
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you (or the server owner) agree to the <a href='https://account.mojang.com/documents/minecraft_eula'>Minecraft EULA?</a>",
        "display": "EULA Agreement (true/false)",
        "internal": false,
        "type": "boolean"
      },
      "version": {
      	"value": "1.12.2-14.23.1.2555",
      	"required": true,
      	"desc": "Version of Forge to install (may be located <a href='http://files.minecraftforge.net/#Downloads'>here</a>",
      	"display": "Version",
      	"internal": false
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const Sponge = `{
  "pufferd": {
    "type": "java",
    "display": "SpongeForge - Minecraft",
    "install": {
      "commands": [
        {
          "type": "download",
          "files": [
          	"http://files.minecraftforge.net/maven/net/minecraftforge/forge/${forgeversion}/forge-${forgeversion}-installer.jar",
          	"http://files.minecraftforge.net/maven/org/spongepowered/spongeforge/${spongeversion}/spongeforge-${spongeversion}.jar"
          ]
        },
        {
          "source": "forge-*.jar",
          "target": "installer.jar",
          "type": "move"
        },
        {
          "target": "mods",
          "type": "mkdir"
        },
        {
          "source": "spongeforge-*.jar",
          "target": "mods",
          "type": "move"
        },
        {
          "commands": [
            "java -jar installer.jar --installServer"
          ],
          "type": "command"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "source": "forge-*-universal.jar",
          "target": "server.jar",
          "type": "move"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "standard"
    },
    "data": {
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you (or the server owner) agree to the <a href='https://account.mojang.com/documents/minecraft_eula'>Minecraft EULA?</a>",
        "display": "EULA Agreement (true/false)",
        "internal": false,
        "type": "boolean"
      },
      "spongeversion": {
      	"value": "1.12.2-2555-7.1.0-BETA-2844",
      	"required": true,
      	"desc": "Version of Sponge to install (may be located <a href='https://www.spongepowered.org/downloads/spongeforge/stable/'>here</a>",
      	"display": "Sponge Version",
      	"internal": false
      },
      "forgeversion": {
      	"value": "1.12.2-14.23.1.2555",
      	"required": true,
      	"desc": "Version of Forge to install (use version specified by Sponge)",
      	"display": "Forge Version",
      	"internal": false
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const BungeecordDocker = `{
  "pufferd": {
    "type": "java",
    "display": "Bungeecord - Minecraft (Docker)",
    "install": {
      "commands": [
        {
          "files": "http://ci.md-5.net/job/BungeeCord/lastSuccessfulBuild/artifact/bootstrap/target/BungeeCord.jar",
          "type": "download"
        },
        {
          "type": "writefile",
          "text": "listeners:\n- query_port: 25577\n  host: ${ip}:${port}\n  motd: '${motd}'\n",
          "target": "config.yml"
        }
      ]
    },
    "run": {
      "stop": "end",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
       "-Xms${memory}M",
      	"-jar",
      	"BungeeCord.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "docker",
      "image": "pufferpanel/minecraft"
    },
    "data": {
      "memory": {
      	"value": "512",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the BungeeCord server. It is generally recommended to allocated 512MB RAM for every 500 players. Please keep in mind that additional plugins may impact the amount needed in your case.",
      	"display": "Memory (MB)",
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
        "value": "25577",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "motd": {
        "value": "A BungeeCord Server &9hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const SpigotDocker = `{
  "pufferd": {
    "type": "java",
    "display": "Spigot - Minecraft (Docker)",
    "install": {
      "commands": [
        {
          "files": "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar",
          "type": "download"
        },
        {
          "commands": [
            "java -jar BuildTools.jar --rev ${version}"
          ],
          "type": "command"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "source": "spigot-*.jar",
          "target": "server.jar",
          "type": "move"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
        "-Djline.terminal=jline.UnsupportedTerminal",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "docker",
      "image": "pufferpanel/minecraft"
    },
    "data": {
      "version": {
      	"value": "latest",
      	"required": true,
      	"desc": "Version of Minecraft to install",
      	"display": "Version",
      	"internal": false
      },
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you agree to the Minecraft EULA?",
        "display": "EULA Agreement (true/false)",
        "internal": false,
        "type": "boolean"
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const CraftbukkitBySpigotDocker = `{
  "pufferd": {
    "type": "java",
    "display": "CraftBukkit by Spigot - Minecraft (Docker)",
    "install": {
      "commands": [
        {
          "files": "https://hub.spigotmc.org/jenkins/job/BuildTools/lastSuccessfulBuild/artifact/target/BuildTools.jar",
          "type": "download"
        },
        {
          "commands": [
            "java -jar BuildTools.jar --rev ${version}"
          ],
          "type": "command"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "source": "craftbukkit-*.jar",
          "target": "server.jar",
          "type": "move"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "docker",
      "image": "pufferpanel/minecraft"
    },
    "data": {
      "version": {
      	"value": "latest",
      	"required": true,
      	"desc": "Version of Minecraft to install",
      	"display": "Version",
      	"internal": false
      },
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you (or the server owner) agree to the <a href='https://account.mojang.com/documents/minecraft_eula'>Minecraft EULA?</a>",
        "display": "EULA Agreement",
        "internal": false,
        "type": "boolean"
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const VanillaDocker = `{
  "pufferd": {
    "type": "java",
    "display": "Vanilla - Minecraft (Docker)",
    "install": {
      "commands": [
        {
          "type": "download",
          "files": "https://s3.amazonaws.com/Minecraft.Download/versions/${version}/minecraft_server.${version}.jar"
        },
        {
          "source": "minecraft_server.*.jar",
          "target": "server.jar",
          "type": "move"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "docker",
      "image": "pufferpanel/minecraft"
    },
    "data": {
      "version": {
      	"value": "1.12.2",
      	"required": true,
      	"desc": "Version of Minecraft to install",
      	"display": "Version",
      	"internal": false
      },
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you (or the server owner) agree to the <a href='https://account.mojang.com/documents/minecraft_eula'>Minecraft EULA?</a>",
        "display": "EULA Agreement",
        "internal": false,
        "type": "boolean"
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const ForgeDocker = `{
  "pufferd": {
    "type": "java",
    "display": "MinecraftForge - Minecraft (Docker)",
    "install": {
      "commands": [
        {
          "type": "download",
          "files": "http://files.minecraftforge.net/maven/net/minecraftforge/forge/${version}/forge-${version}-installer.jar"
        },
        {
          "source": "forge-*.jar",
          "target": "installer.jar",
          "type": "move"
        },
        {
          "commands": [
            "java -jar installer.jar --installServer"
          ],
          "type": "command"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        },
        {
          "source": "forge-*-universal.jar",
          "target": "server.jar",
          "type": "move"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "docker",
      "image": "pufferpanel/minecraft"
    },
    "data": {
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you (or the server owner) agree to the <a href='https://account.mojang.com/documents/minecraft_eula'>Minecraft EULA?</a>",
        "display": "EULA Agreement (true/false)",
        "internal": false,
        "type": "boolean"
      },
      "version": {
      	"value": "1.12.2-14.23.1.2555",
      	"required": true,
      	"desc": "Version of Forge to install (may be located <a href='http://files.minecraftforge.net/#Downloads'>here</a>",
      	"display": "Version",
      	"internal": false
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`

const SpongeDocker = `{
  "pufferd": {
    "type": "java",
    "display": "SpongeForge - Minecraft (Docker)",
    "install": {
      "commands": [
        {
          "type": "download",
          "files": [
          	"http://files.minecraftforge.net/maven/net/minecraftforge/forge/${forgeversion}/forge-${forgeversion}-installer.jar",
          	"http://files.minecraftforge.net/maven/org/spongepowered/spongeforge/${spongeversion}/spongeforge-${spongeversion}.jar"
          ]
        },
        {
          "source": "forge-*.jar",
          "target": "installer.jar",
          "type": "move"
        },
        {
          "target": "mods",
          "type": "mkdir"
        },
        {
          "source": "spongeforge-*.jar",
          "target": "mods",
          "type": "move"
        },
        {
          "commands": [
            "java -jar installer.jar --installServer"
          ],
          "type": "command"
        },
        {
          "type": "writefile",
          "text": "eula=${eula}",
          "target": "eula.txt"
        },
        {
          "type": "writefile",
          "text": "server-ip=${ip}\nserver-port=${port}\nmotd=${motd}\n",
          "target": "server.properties"
        },
        {
          "source": "forge-*-universal.jar",
          "target": "server.jar",
          "type": "move"
        }
      ]
    },
    "run": {
      "stop": "stop",
      "pre": [],
      "post": [],
      "arguments": [
      	"-Xmx${memory}M",
      	"-jar",
      	"server.jar"
      ],
      "program": "java"
    },
    "environment": {
      "type": "docker",
      "image": "pufferpanel/minecraft"
    },
    "data": {
      "memory": {
      	"value": "1024",
      	"required": true,
      	"desc": "How much memory in MB to allocate to the Java Heap",
      	"display": "Memory (MB)",
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
        "value": "25565",
        "required": true,
        "desc": "What port to bind the server to",
        "display": "Port",
        "internal": false,
        "type": "integer"
      },
      "eula": {
        "value": "false",
        "required": true,
        "desc": "Do you (or the server owner) agree to the <a href='https://account.mojang.com/documents/minecraft_eula'>Minecraft EULA?</a>",
        "display": "EULA Agreement (true/false)",
        "internal": false,
        "type": "boolean"
      },
      "spongeversion": {
      	"value": "1.12.2-2555-7.1.0-BETA-2844",
      	"required": true,
      	"desc": "Version of Sponge to install (may be located <a href='https://www.spongepowered.org/downloads/spongeforge/stable/'>here</a>",
      	"display": "Sponge Version",
      	"internal": false
      },
      "forgeversion": {
      	"value": "1.12.2-14.23.1.2555",
      	"required": true,
      	"desc": "Version of Forge to install (use version specified by Sponge)",
      	"display": "Forge Version",
      	"internal": false
      },
      "motd": {
        "value": "A Minecraft Server\\n\\u00A79 hosted on PufferPanel",
        "required": true,
        "desc": "This is the message that is displayed in the server list of the client, below the name. The MOTD does support <a href='https://minecraft.gamepedia.com/Formatting_codes' target='_blank'>color and formatting codes</a>.",
        "display": "MOTD message of the day",
        "internal": false
      }
    }
  }
}`
