// +build !windows

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

package environments

import (
	"errors"
	"github.com/kr/pty"
	"github.com/pufferpanel/apufferi/logging"
	ppError "github.com/pufferpanel/pufferd/errors"
	"github.com/shirou/gopsutil/process"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"
	"strings"
	"fmt"
)

type tty struct {
	*BaseEnvironment
	mainProcess *exec.Cmd
	stdInWriter io.Writer
}

func createTty() *tty {
	t := &tty{BaseEnvironment: &BaseEnvironment{Type: "tty"}}
	t.BaseEnvironment.executeAsync = t.ttyExecuteAsync
	t.BaseEnvironment.waitForMainProcess = t.WaitForMainProcess
	return t
}

func (s *tty) ttyExecuteAsync(cmd string, args []string, env map[string]string, callback func(graceful bool)) (err error) {
	running, err := s.IsRunning()
	if err != nil {
		return
	}
	if running {
		err = errors.New("process is already running (" + strconv.Itoa(s.mainProcess.Process.Pid) + ")")
		return
	}
	process := exec.Command(cmd, args...)
	process.Dir = s.RootDirectory
	process.Env = append(os.Environ(), "HOME="+s.RootDirectory)
	for k, v := range env {
		s.mainProcess.Env = append(s.mainProcess.Env, fmt.Sprintf("%s=%s", k, v))
	}

	wrapper := s.createWrapper()
	s.wait = sync.WaitGroup{}
	s.wait.Add(1)
	process.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setsid: true}
	s.mainProcess = process
	logging.Debug("Starting process: %s %s", s.mainProcess.Path, strings.Join(s.mainProcess.Args, " "))
	tty, err := pty.Start(process)
	s.stdInWriter = tty
	go func() {
		io.Copy(wrapper, tty)
		process.Wait()
		s.wait.Done()
		if callback != nil {
			if s.mainProcess == nil || s.mainProcess.ProcessState == nil {
				callback(false)
			} else {
				callback(s.mainProcess.ProcessState.Success())
			}
		}
	}()
	if err != nil {
		logging.Error("Error starting process", err)
	}
	return
}

func (s *tty) ExecuteInMainProcess(cmd string) (err error) {
	running, err := s.IsRunning()
	if err != nil {
		return err
	}
	if !running {
		err = errors.New("main process has not been started")
		return
	}
	stdIn := s.stdInWriter
	_, err = io.WriteString(stdIn, cmd+"\n")
	return
}

func (s *tty) Kill() (err error) {
	running, err := s.IsRunning()
	if err != nil {
		return err
	}
	if running {
		return
	}
	err = s.mainProcess.Process.Kill()
	s.mainProcess.Process.Release()
	s.mainProcess = nil
	return
}

func (s *tty) IsRunning() (isRunning bool, err error) {
	isRunning = s.mainProcess != nil && s.mainProcess.Process != nil
	if isRunning {
		process, pErr := os.FindProcess(s.mainProcess.Process.Pid)
		if process == nil || pErr != nil {
			isRunning = false
		} else if process.Signal(syscall.Signal(0)) != nil {
			isRunning = false
		}
	}
	return
}

func (s *tty) GetStats() (map[string]interface{}, error) {
	running, err := s.IsRunning()
	if err != nil {
		return nil, err
	}
	if !running {
		return nil, ppError.NewServerOffline()
	}
	process, err := process.NewProcess(int32(s.mainProcess.Process.Pid))
	if err != nil {
		return nil, err
	}
	resultMap := make(map[string]interface{})
	memMap, _ := process.MemoryInfo()
	resultMap["memory"] = memMap.RSS
	cpu, _ := process.Percent(time.Millisecond * 50)
	resultMap["cpu"] = cpu
	return resultMap, nil
}

func (e *tty) Create() error {
	return os.Mkdir(e.RootDirectory, 0755)
}

func (e *tty) WaitForMainProcess() error {
	return e.WaitForMainProcessFor(0)
}

func (e *tty) WaitForMainProcessFor(timeout int) (err error) {
	running, err := e.IsRunning()
	if err != nil {
		return
	}
	if running {
		if timeout > 0 {
			var timer = time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
				err = e.Kill()
			})
			e.wait.Wait()
			timer.Stop()
		} else {
			e.wait.Wait()
		}
	}
	return
}

func (e *tty) SendCode(code int) error {
	running, err := e.IsRunning()

	if err != nil || !running {
		return err
	}

	return e.mainProcess.Process.Signal(syscall.Signal(code))
}