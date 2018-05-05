package environments

import (
	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/cache"
	"github.com/pufferpanel/pufferd/utils"
)

func LoadEnvironment(environmentType, folder, id string, environmentSection map[string]interface{}) Environment {
	serverRoot := common.JoinPath(folder, id)
	rootDirectory := common.GetStringOrDefault(environmentSection, "root", serverRoot)
	cache := cache.CreateCache()
	wsManager := utils.CreateWSManager()
	switch environmentType {
	case "tty":
		logging.Debugf("Loading server as tty")
		t := createTty()
		t.RootDirectory = rootDirectory
		t.ConsoleBuffer = cache
		t.WSManager = wsManager
		return t
	case "docker":
		logging.Debugf("Loading server as docker")
		d := createDocker(id, common.GetStringOrDefault(environmentSection, "image", "pufferpanel/generic"))
		d.RootDirectory = rootDirectory
		d.ConsoleBuffer = cache
		d.WSManager = wsManager
		return d
	default:
		logging.Debugf("Loading server as standard")
		s := createStandard()
		s.RootDirectory = rootDirectory
		s.ConsoleBuffer = cache
		s.WSManager = wsManager
		return s
	}
}
