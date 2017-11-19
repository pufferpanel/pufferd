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

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/programs"
)

func CopyTemplates() {
	os.MkdirAll(programs.TemplateFolder, 0755)

	writeFile("spigot", Spigot)
	writeFile("spigotdocker", SpigotDocker)

	writeFile("bungeecord", Bungeecord)
	writeFile("bungeecorddocker", BungeecordDocker)

	writeFile("fakecraftbukkit", CraftbukkitBySpigot)
	writeFile("fakecraftbukkitdocker", CraftbukkitBySpigotDocker)

	writeFile("vanillaminecraft", Vanilla)
	writeFile("vanillaminecraftdocker", VanillaDocker)

	writeFile("forge", Forge)
	writeFile("forgedocker", ForgeDocker)

	writeFile("spongeforge", Sponge)
	writeFile("spongeforgedocker", SpongeDocker)

	writeFile("srcds", SRCDS)

	writeFile("tf2", TF2)

	writeFile("csgo", CSGO)

	writeFile("gmod", GMOD)

	writeFile("pocketmine", Pocketmine)

	writeFile("factorio", Factorio)

	writeFile("arma3", Arma3)
}

func writeFile(name string, data string) {
	jsonData := []byte(data)
	var prettyJson bytes.Buffer
	json.Indent(&prettyJson, jsonData, "", "  ")
	err := ioutil.WriteFile(common.JoinPath(programs.TemplateFolder, name+".json"), prettyJson.Bytes(), 0664)
	if err != nil {
		logging.Error("Error writing template "+name, err)
	}
}