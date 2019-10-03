package envs

import (
	"os/exec"
)

func Chroot(process *exec.Cmd, rootDir string) {
	process.SysProcAttr.Chroot = rootDir
}
