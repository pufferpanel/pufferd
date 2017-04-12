package uninstaller

import (
	"github.com/pufferpanel/pufferd/config"
	"github.com/pufferpanel/pufferd/logging"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func StartProcess() {
	killDaemon()
	deleteFiles()
	deleteUser()
	return
}

func killDaemon() {
	exec.Command("systemctl", "stop", "pufferd").Run()
	logging.Info("Attempting to kill all pufferd process...")
	time.Sleep(time.Second * 15)                         //Giving 5 seconds to kill "correctly" all process
	exec.Command("killall", "-9", "-u", "pufferd").Run() //"Hard killing" anything
}

func deleteUser() {
	cmd := exec.Command("userdel", "-Z", "-r", "-f", "pufferd")
	err := cmd.Run() //Delete pufferd and it's home dir (/var/lib/pufferd)

	if err != nil {
		flag := false //flag which indicate if the
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {

				switch status.ExitStatus() {
				case 6:
					logging.Error("The pufferd user don't exist" + err)
					flag = true
				case 8:
					logging.Error("The pufferd user is logged in" + err)
					flag = true
				case 12:
					logging.Error("Couldn't remove pufferd directory" + err)
					flag = true
				case 10:
					logging.Error("Couldn't update group file" + err)
					flag = true

				}
			}
		}
		if !flag {
			logging.Error("Couldn't delete the pufferd user", err)
		}
	}

}

func deleteFiles() {

	//disable service
	cmd := exec.Command("systemctl", "disable", "pufferd")
	_, err := cmd.CombinedOutput()
	if err != nil {
		logging.Error("Error disabling pufferd service, is it installed?", err)
	}

	//delete service
	err = os.Remove("/etc/systemd/system/pufferd.service")
	if err != nil {
		logging.Error("Error deleting the pufferd service file:", err)
	}

	err = os.RemoveAll(config.Get("serverfolder"))
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
