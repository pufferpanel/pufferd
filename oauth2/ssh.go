package oauth2

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/pufferpanel/apufferi/config"
	"github.com/pufferpanel/apufferi/logging"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func ValidateSSH(username string, password string) (*ssh.Permissions, error) {
	return validateSSH(username, password, true)
}

func validateSSH(username string, password string, recurse bool) (*ssh.Permissions, error) {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", username)
	data.Set("password", password)
	data.Set("scope", "sftp")
	encodedData := data.Encode()
	request, _ := http.NewRequest("POST", config.GetString("authServer"), bytes.NewBufferString(encodedData))

	RefreshIfStale()

	atLocker.RLock()
	request.Header.Add("Authorization", "Bearer "+daemonToken)
	atLocker.RUnlock()
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))
	response, err := client.Do(request)
	if err != nil {
		logging.Error("Error talking to auth server", err)
		return nil, errors.New("Invalid response from authorization server")
	}
	defer response.Body.Close()

	//we should only get a 200, if we get any others, we have a problem
	if response.StatusCode != 200 {
		if response.StatusCode == 401 {
			if recurse && RefreshToken() {
				response.Body.Close()
				return validateSSH(username, password, false)
			}
		}

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