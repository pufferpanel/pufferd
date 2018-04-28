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

	ppError "github.com/pufferpanel/pufferd/errors"
	"github.com/docker/docker/client"
	"context"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/network"
	"io"
	"io/ioutil"
	"time"
	"runtime"
	"os"
	"fmt"
	"syscall"
)

type docker struct {
	*BaseEnvironment
	ContainerId string `json:"-"`
	ImageName   string `json:"image"`
	connection  types.HijackedResponse
	cli			*client.Client
	downloadingImage bool
}

func createDocker(containerId, imageName string) *docker {
	if imageName == "" {
		imageName = "pufferpanel/generic"
	}
	d := &docker{BaseEnvironment: &BaseEnvironment{Type: "docker"}, ContainerId: containerId, ImageName: imageName}
	d.BaseEnvironment.executeAsync = d.dockerExecuteAsync
	d.BaseEnvironment.waitForMainProcess = d.WaitForMainProcess
	return d
}

func (d *docker) dockerExecuteAsync(cmd string, args []string, env map[string]string, callback func(graceful bool)) (error) {
	running, err := d.IsRunning()
	if err != nil {
		return err

	}
	if running {
		return errors.New("container is already running")
	}

	if d.downloadingImage {
		return errors.New("container image is downloading, cannot execute")
	}

	client, err := d.getClient()
	ctx := context.Background()

	exists, err := d.doesContainerExist(client, ctx)

	if err != nil {
		return err
	}

	//container does not exist
	if !exists {
		err = d.createContainer(client, ctx, cmd, args, env, d.RootDirectory)
		if err != nil {
			return err
		}
	}

	config := types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	}

	d.connection, err = client.ContainerAttach(ctx, d.ContainerId, config)
	if err != nil {
		return err
	}

	d.wait.Add(1)

	go func() {
		defer d.connection.Close()
		wrapper := d.createWrapper()
		io.Copy(wrapper, d.connection.Reader)
		c, _ := d.getClient()
		c.ContainerStop(context.Background(), d.ContainerId, nil)
		time.Sleep(1 * time.Second)
		d.wait.Done()
	}()

	startOpts := types.ContainerStartOptions{
	}

	err = client.ContainerStart(ctx, d.ContainerId, startOpts)
	if err != nil {
		return err
	}
	return err
}

func (d *docker) ExecuteInMainProcess(cmd string) (err error) {
	running, err := d.IsRunning()
	if err != nil {
		return
	}
	if !running {
		err = errors.New("main process has not been started")
		return
	}

	d.connection.Conn.Write([]byte(cmd + "\n"))
	return
}

func (d *docker) Kill() (err error) {
	running, err := d.IsRunning()
	if err != nil {
		return err
	}

	if !running {
		return
	}

	client, err := d.getClient()
	ctx := context.Background()
	err = client.ContainerKill(ctx, d.ContainerId, "SIGKILL")
	return
}

func (d *docker) Create() error {
	err := os.Mkdir(d.RootDirectory, 0755)
	if err != nil {
		return err
	}

	go func() {
		cli, err := d.getClient()
		if err != nil {
			return
		}
		err = d.pullImage(cli, context.Background(), false)
	}()

	return err
}

func (d *docker) IsRunning() (bool, error) {
	client, err := d.getClient()
	if err != nil {
		logging.Error("Error checking run status", err)
		return false, err
	}

	ctx := context.Background()

	exists, err := d.doesContainerExist(client, ctx)
	if !exists {
		return false, err
	}

	stats, err := client.ContainerInspect(ctx, d.ContainerId)
	if err != nil {
		logging.Error("Error checking run status", err)
		return false, err
	}
	return stats.State.Running, nil
}

func (d *docker) GetStats() (map[string]interface{}, error) {
	running, err := d.IsRunning()
	if err != nil {
		return nil, err
	}

	if !running {
		return nil, ppError.NewServerOffline()
	}

	resultMap := make(map[string]interface{})
	resultMap["memory"] = 0
	resultMap["cpu"] = 0
	return resultMap, nil
}

func (e *docker) WaitForMainProcess() error {
	return e.WaitForMainProcessFor(0)
}

func (e *docker) WaitForMainProcessFor(timeout int) (err error) {
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

func (d *docker) getClient() (*client.Client, error) {
	var err error = nil
	if d.cli == nil {
		d.cli, err = client.NewEnvClient()
		ctx := context.Background()
		d.cli.NegotiateAPIVersion(ctx)
	}
	return d.cli, err
}

func (d *docker) doesContainerExist(client *client.Client, ctx context.Context) (bool, error) {
	opts := types.ContainerListOptions{
		Filters: filters.NewArgs(),
	}

	opts.All = true
	opts.Filters.Add("name", d.ContainerId)

	existingContainers, err := client.ContainerList(ctx, opts)

	logging.Debugf("Does container (%s) exist?: %t", d.ContainerId, len(existingContainers) > 0)

	if len(existingContainers) == 0 {
		return false, err
	} else {
		return true, err
	}
}

func (d *docker) pullImage(client *client.Client, ctx context.Context, force bool) error {
	exists := false

	opts := types.ImageListOptions{
		All:     true,
		Filters: filters.NewArgs(),
	}
	opts.Filters.Add("reference", d.ImageName)
	images, err := client.ImageList(ctx, opts)

	if err != nil {
		return err
	}

	if len(images) >= 1 {
		exists = true
	}

	logging.Debugf("Does image %v exist? %v", d.ImageName, exists)

	if exists && !force {
		return nil
	}

	op := types.ImagePullOptions{
	}

	logging.Debugf("Downloading image %v", d.ImageName)
	d.DisplayToConsole("Downloading image for container, please wait\n")

	d.downloadingImage = true

	r, err := client.ImagePull(ctx, d.ImageName, op)
	defer r.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(ioutil.Discard, r)

	d.downloadingImage = false
	logging.Debugf("Downloaded image %v", d.ImageName)
	d.DisplayToConsole("Downloaded image for container\n")
	return err
}

func (d *docker) createContainer(client *client.Client, ctx context.Context, cmd string, args []string, env map[string]string, root string) error {
	err := d.pullImage(client, ctx, false)

	if err != nil {
		return err
	}

	cmdSlice := strslice.StrSlice{}

	cmdSlice = append(cmdSlice, cmd)

	for _, v := range args {
		cmdSlice = append(cmdSlice, v)
	}

	newEnv := os.Environ()
	//newEnv["home"] = root
	newEnv = append(newEnv, "HOME=" + root)
	for k, v := range env {
		newEnv = append(newEnv, fmt.Sprintf("%s=%s", k, v))
	}

	config := &container.Config{
		AttachStderr:    true,
		AttachStdin:     true,
		AttachStdout:    true,
		Tty:             true,
		OpenStdin:       true,
		NetworkDisabled: false,
		Cmd:             cmdSlice,
		Image:           d.ImageName,
		WorkingDir:		 root,
		Env:			 newEnv,
	}

	if runtime.GOOS == "linux" {
		config.User = fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
	}

	hostConfig := &container.HostConfig{
		AutoRemove:  true,
		NetworkMode: "host",
		Resources: container.Resources{
		},
		Binds: make([]string, 0),
	}
	hostConfig.Binds = append(hostConfig.Binds, root+":"+root)

	networkConfig := &network.NetworkingConfig{
	}

	_, err = client.ContainerCreate(ctx, config, hostConfig, networkConfig, d.ContainerId)
	return err
}

func (e *docker) SendCode(code int) error {
	running, err := e.IsRunning()

	if err != nil || !running {
		return err
	}

	client, err := e.getClient()

	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.ContainerKill(ctx, e.ContainerId, syscall.Signal(code).String())
}