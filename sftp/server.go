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
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	configuration "github.com/pufferpanel/apufferi/config"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/programs"
	"golang.org/x/crypto/ssh"
	"path/filepath"
)

var sftpServer net.Listener

func Run() {
	e := runServer()
	if e != nil {
		logging.Error("Error starting SFTP", e)
	}
}

func Stop() {
	sftpServer.Close()
}

func runServer() error {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return validateSSH(c.User(), string(pass))
		},
	}

	serverKeyFile := path.Join(configuration.GetStringOrDefault("dataFolder", "data"), "server.key")

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

	bind := configuration.GetStringOrDefault("sftp", "0.0.0.0:5657")

	sftpServer, e = net.Listen("tcp", bind)
	if e != nil {
		return e
	}
	logging.Infof("Started SFTP Server on %s", bind)

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
	defer conn.Close()
	logging.Debugf("SFTP connection from %s", conn.RemoteAddr().String())
	e := handleConn(conn, config)
	if e != nil {
		if e.Error() != "EOF" {
			logging.Error("sftpd connection error:", e)
		}
	}
}
func handleConn(conn net.Conn, config *ssh.ServerConfig) error {
	sc, chans, reqs, e := ssh.NewServerConn(conn, config)
	defer func() {
		if sc != nil {
			sc.Close()
		}
	}()
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

var client = &http.Client{}

func validateSSH(username string, password string) (*ssh.Permissions, error) {
	authUrl := configuration.GetString("authServer")
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", username)
	data.Set("password", password)
	data.Set("scope", "sftp")
	token := configuration.GetString("authToken")
	request, _ := http.NewRequest("POST", authUrl, bytes.NewBufferString(data.Encode()))
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	response, err := client.Do(request)
	defer func() {
		if response != nil {
			response.Body.Close()
		}
	}()
	if err != nil {
		logging.Error("Error talking to auth server", err)
		return nil, errors.New("Invalid response from authorization server")
	}

	//we should only get a 200, if we get any others, we have a problem
	if response.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(response.Body)

		logging.Errorf("Error talking to auth server: [%d] [%s]", response.StatusCode, msg)
		return nil, errors.New("Invalid response from authorization server")
	}

	var respArr map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&respArr)
	if err != nil {
		return nil, err
	}
	if respArr["error"] != nil {
		return nil, errors.New("Incorrect username or password")
	}
	sshPerms := &ssh.Permissions{}
	scopes := strings.Split(respArr["scope"].(string), " ")
	if len(scopes) != 2 {
		return nil, errors.New("Invalid response from authorization server")
	}
	for _, v := range scopes {
		if v != "sftp" {
			sshPerms.Extensions = make(map[string]string)
			sshPerms.Extensions["server_id"] = v
			return sshPerms, nil
		}
	}
	return nil, errors.New("Incorrect username or password")
}
