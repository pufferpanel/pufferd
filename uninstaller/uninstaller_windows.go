package uninstaller

import (
	"github.com/pufferpanel/pufferd/config"
	"github.com/pufferpanel/pufferd/logging"
	"os"
)

func StartProcess() {
	deleteFiles()
}

func deleteFiles() {
	err := os.RemoveAll(config.Get("serverfolder"))
	if err != nil {
		logging.Error("Error deleting pufferd server folder, stored in " + config.Get("serverfolder"), err)
	}

	err = os.RemoveAll(config.Get("templatefolder"))
	if err != nil {
		logging.Error("Error deleting pufferd template folder, stored in " + config.Get("templatefolder"), err)
	}

	err = os.RemoveAll(config.Get("datafolder"))
	if err != nil {
		logging.Error("Error deleting pufferd data folder, stored in " + config.Get("datafolder"), err)
	}
}
