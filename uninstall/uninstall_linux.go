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
	err := exec.Command("userdel", "-Z", "-r", "-f", "pufferd").Run() //Delete pufferd and it's home dir (/var/lib/pufferd)

	/*
	Need to use status.ExitStatus()
	*/
	if err != nil{
		errCode := err.Error()[len(err.Error())-1]
		switch errCode{
			case '6':
				logging.Error("The pufferd user don't exist.")
			case '8':
				logging.Error("The pufferd user is logged in")
			case '2'://Wrong way, I know
				logging.Error("	Couldn't remove pufferd directory.")
			case '0'://Wrong way, I know
				logging.Error("Couldn't delete pufferd user: couldncan't update group file.")
			default:
				logging.Error("Couldn't delete the pufferd group")
		}
	}



	err = exec.Command("groupdel", "pufferd").Run()
	if err != nil{
		errCode := err.Error()[len(err.Error())-1]
		switch errCode{
			case '6':
				logging.Error("The pufferd group don't exist.")
			case '0'://Wrong way, I know
				logging.Error("Couldn't delete pufferd group: couldn't update group file.")
			default:
				logging.Error("Couldn't delete the pufferd group")
		}
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


