package uninstaller

import (
	"github.com/pufferpanel/pufferd/logging"
	"github.com/pufferpanel/pufferd/config"
	"os"
)


func StartProcess(){
	deleteFiles()
}

func deleteFiles(){
err := os.RemoveAll(config.Get("serverfolder"))
	if err != nil {
		logging.Error("Error deleting pufferd server folder, stored in %s",config.Get("serverfolder") , err)
	}

	err = os.RemoveAll(config.Get("templatefolder"))
	if err != nil {
		logging.Error("Error deleting pufferd template folder, stored in %s",config.Get("templatefolder") , err)
	}

	err = os.RemoveAll(config.Get("datafolder"))
	if err != nil {
		logging.Error("Error deleting pufferd data folder, stored in %s",config.Get("datafolder") , err)
	}
}