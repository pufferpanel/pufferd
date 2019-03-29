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

package sftp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/sftp"
	"github.com/pufferpanel/apufferi/common"
	"github.com/pufferpanel/apufferi/logging"
	pufferdConfig "github.com/pufferpanel/pufferd/config"
	"github.com/pufferpanel/pufferd/oauth2"
	"github.com/pufferpanel/pufferd/programs"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
)

var sftpServer net.Listener

func Run() {
	e := runServer()
	if e != nil {
		logging.Build(logging.ERROR).WithMessage("Error starting SFTP server").WithError(e).Log()
	}
}

func Stop() {
	sftpServer.Close()
}

func runServer() error {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return oauth2.ValidateSSH(c.User(), string(pass))
		},
	}

	serverKeyFile := pufferdConfig.Get().Listener.SFTPKey

	_, e := os.Stat(serverKeyFile)

	if e != nil && os.IsNotExist(e) {
		logging.Debug("Generating new key")
		var key *rsa.PrivateKey
		key, e = rsa.GenerateKey(rand.Reader, 2048)
		if e != nil {
			return e
		}

		data := x509.MarshalPKCS1PrivateKey(key)
		block := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   data,
		}
		e = ioutil.WriteFile(serverKeyFile, pem.EncodeToMemory(&block), 0700)
		if e != nil {
			return e
		}
	} else if e != nil {
		return e
	}

	logging.Debug("Loading existing key")
	var data []byte
	data, e = ioutil.ReadFile(serverKeyFile)
	if e != nil {
		return e
	}

	hkey, e := ssh.ParsePrivateKey(data)

	if e != nil {
		return e
	}

	config.AddHostKey(hkey)

	bind := pufferdConfig.Get().Listener.SFTP

	sftpServer, e = net.Listen("tcp", bind)
	if e != nil {
		return e
	}
	logging.Info("Started SFTP Server on %s", bind)

	go func() {
		for {
			conn, _ := sftpServer.Accept()
			if conn != nil {
				go HandleConn(conn, config)
			}
		}
	}()

	return nil
}

func HandleConn(conn net.Conn, config *ssh.ServerConfig) {
	defer common.Close(conn)
	logging.Debug("SFTP connection from %s", conn.RemoteAddr().String())
	e := handleConn(conn, config)
	if e != nil {
		if e.Error() != "EOF" {
			logging.Error("sftpd connection error: %s", e.Error())
		}
	}
}
func handleConn(conn net.Conn, config *ssh.ServerConfig) error {
	sc, chans, reqs, e := ssh.NewServerConn(conn, config)
	defer common.Close(sc)
	if e != nil {
		return e
	}

	// The incoming Request channel must be serviced.
	go PrintDiscardRequests(reqs)

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of an SFTP session, this is "subsystem"
		// with a payload string of "<length=4>sftp"
		if newChannel.ChannelType() != "session" {
			err := newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			if err != nil {
				return err
			}
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			return err
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "subsystem" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false
				switch req.Type {
				case "subsystem":
					if string(req.Payload[4:]) == "sftp" {
						ok = true
					}
				}
				req.Reply(ok, nil)
			}
		}(requests)

		fs := CreateRequestPrefix(filepath.Join(programs.ServerFolder, sc.Permissions.Extensions["server_id"]))

		server := sftp.NewRequestServer(channel, fs)

		if err := server.Serve(); err != nil {
			return err
		}
	}
	return nil
}

func PrintDiscardRequests(in <-chan *ssh.Request) {
	for req := range in {
		if req.WantReply {
			req.Reply(false, nil)
		}
	}
}
