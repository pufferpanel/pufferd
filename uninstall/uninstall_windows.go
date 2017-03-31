package uninstall

import (
	"time"
	"io/ioutil"
	"os/exec"
	"github.com/pufferpanel/pufferd/logging"
	"syscall"
)


func StartProcess(){
	logging.Error("No configured service uninstaller for this OS")
}
