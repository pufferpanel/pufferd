package uninstall

import (
	"time"
	"io/ioutil"
	"os/exec"
	"github.com/pufferpanel/pufferd/logging"
	"syscall"
)


func StartProcess(){
	killDaemon()
	deleteFiles()
	deleteUser()
}

func killDaemon(){
	exec.Command("killall", "-15" "-u", "pufferd").Run()
	logging.Info("Attempting to kill all pufferd process...")
	time.Sleep(time.Second * 5)//Giving 5 seconds to kill "correctly" all process
	exec.Command("killall", "-9" "-u", "pufferd").Run()//"Hard killing" anything
}

func deleteUser(){
	cmd := exec.Command("userdel", "-Z", "-r", "-f", "pufferd")
	err := cmd.Run() //Delete pufferd and it's home dir (/var/lib/pufferd)

	if err != nil{

		if exitErr, ok := err.(*exec.ExitError); ok {

			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				switch status.ExitStatus(){
					case 6:
						logging.Error("The pufferd user don't exist", err)
					case 8:
						logging.Error("The pufferd user is logged in", err)
					case 12://Wrong way, I know
						logging.Error("	Couldn't remove pufferd directory", err)
					case 10://Wrong way, I know
						logging.Error("Couldn't update group file", err)
					
				}
			}
		}
		logging.Error("Couldn't delete the pufferd user")
	}



}

func deleteFiles(){

	//disable service
	cmd = exec.Command("systemctl", "disable", "pufferd")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logging.Error("Error disabling pufferd service, is it installed?", err)
	}

	//delete service
	err :=os.Remove("/etc/systemd/system/pufferd.service")
	if err != nil{
		logging.Error("Error deleting the pufferd service file:", err)
	}
}


