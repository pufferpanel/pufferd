/*
 Copyright 2019 Padduck, LLC

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

package oauth2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/v3"
	"github.com/pufferpanel/apufferi/v3/response"
	"github.com/pufferpanel/apufferi/v3/logging"
	"github.com/pufferpanel/pufferd/v2/commons"
	"github.com/pufferpanel/pufferd/v2/config"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

func ValidateToken(accessToken string, gin *gin.Context) bool {
	return validateToken(accessToken, gin, true)
}

func validateToken(accessToken string, gin *gin.Context, recurse bool) bool {
	authUrl := config.Get().Auth.InfoURL
	data := url.Values{}
	data.Set("token", accessToken)
	encodedData := data.Encode()
	request, _ := http.NewRequest("POST", authUrl, bytes.NewBufferString(encodedData))

	RefreshIfStale()

	atLocker.RLock()
	request.Header.Add("Authorization", "Bearer "+daemonToken)
	atLocker.RUnlock()
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))
	oauthResponse, err := client.Do(request)
	defer commons.CloseResponse(oauthResponse)
	if err != nil {
		logging.Exception("error talking to auth server", err)
		response.Respond(gin).Message(err.Error()).Fail().Status(500).Send()
		gin.Abort()
		return false
	}

	if oauthResponse.StatusCode != 200 {
		if oauthResponse.StatusCode == 401 {
			//refresh token and repeat call
			//if we didn't refresh, then there's no reason to try again
			if recurse && RefreshToken() {
				commons.CloseResponse(oauthResponse)
				return validateToken(accessToken, gin, false)
			}
		}

		logging.Error("Unexpected response code from auth server: %s", oauthResponse.StatusCode)
		response.Respond(gin).Message(fmt.Sprintf("unexpected response code %d", oauthResponse.StatusCode)).Fail().Status(500).Send()
		gin.Abort()
		return false
	}

	var respArr map[string]interface{}
	err = json.NewDecoder(oauthResponse.Body).Decode(&respArr)

	if err != nil {
		logging.Exception("error parsing auth server response", err)
		response.Respond(gin).Message(err.Error()).Fail().Status(500).Send()
		gin.Abort()
		return false
	} else if respArr["error"] != nil {
		errStr, ok := respArr["error"].(string)
		if !ok {
			err = errors.New(fmt.Sprintf("error is %s instead of string", reflect.TypeOf(respArr["error"])))
			logging.Exception("error parsing auth server response", err)
		} else {
			err = errors.New(errStr)
		}
		response.Respond(gin).Message(err.Error()).Fail().Status(500).Send()
		gin.Abort()
		return false
	}

	active, ok := respArr["active"].(bool)

	if !ok || !active {
		gin.AbortWithStatus(401)
		return false
	}

	serverMapping, ok := respArr["servers"].(map[string]interface{})
	if !ok {
		err = errors.New(fmt.Sprintf("auth server did not respond in the format expected, got %s instead of map[string]interface{} for servers", reflect.TypeOf(respArr["servers"])))
		logging.Exception("error parsing auth server response", err)
		response.Respond(gin).Message(err.Error()).Fail().Status(500).Send()
		gin.Abort()
		return false
	}

	mapping := make(map[string][]string)

	for k, v := range serverMapping {
		mapping[k] = apufferi.ToStringArray(v)
	}

	gin.Set("serverScopes", mapping)
	return true
}
