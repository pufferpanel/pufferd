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

package httphandlers

import (
	"bytes"
	"encoding/json"
	"github.com/pufferpanel/apufferi/common"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/config"
	pufferdHttp "github.com/pufferpanel/apufferi/http"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/programs"
)

type oauthCache struct {
	oauthToken string
	scopes     map[string][]string
	expireTime int64
}

var cache = make([]*oauthCache, 20)

func OAuth2Handler(scope string, requireServer bool) gin.HandlerFunc {
	return func(gin *gin.Context) {
		failure := true
		defer func() {
			if failure && !gin.IsAborted() {
				gin.Abort()
			}
		}()
		authHeader := gin.Request.Header.Get("Authorization")
		var authToken string
		if authHeader == "" {
			authToken = gin.Query("accessToken")
			if authToken == "" {
				pufferdHttp.Respond(gin).Fail().Code(pufferdHttp.NOTAUTHORIZED).Status(400).Message("no access token provided").Send()
				gin.Abort()
				return
			}
		} else {
			authArr := strings.SplitN(authHeader, " ", 2)
			if len(authArr) < 2 || authArr[0] != "Bearer" {
				pufferdHttp.Respond(gin).Code(pufferdHttp.NOTAUTHORIZED).Fail().Status(400).Message("invalid access token format").Send()
				gin.Abort()
				return
			}
			authToken = authArr[1]
		}

		if !validateToken(authToken, gin) {
			gin.Abort()
			return
		}

		serverId := gin.Param("id")
		internalMap, _ := gin.Get("serverScopes")
		scopes := internalMap.(map[string][]string)

		var scopeSet []string

		if requireServer {
			scopeSet = scopes[serverId]
			if scopeSet == nil || len(scopeSet) == 0 {
				pufferdHttp.Respond(gin).Fail().Status(403).Code(pufferdHttp.NOTAUTHORIZED).Message("invalid access").Send()
				return
			}

			var program programs.Program
			program, _ = programs.Get(serverId)
			if program == nil {
				pufferdHttp.Respond(gin).Fail().Status(404).Code(pufferdHttp.NOSERVER).Message("no server with id " + serverId).Send()
				return
			}

			gin.Set("server", program)
		} else {
			scopeSet = scopes[""]
			if scopeSet == nil || len(scopeSet) == 0 {
				pufferdHttp.Respond(gin).Fail().Status(403).Code(pufferdHttp.NOTAUTHORIZED).Message("invalid access").Send()
				return
			}
		}

		if !common.ContainsValue(scopeSet, scope) {
			pufferdHttp.Respond(gin).Fail().Status(403).Code(pufferdHttp.NOTAUTHORIZED).Message("missing scope " + scope).Send()
			return
		}

		failure = false
	}
}

func validateToken(accessToken string, gin *gin.Context) bool {
	authUrl := config.GetString("infoServer")
	token := config.GetString("authToken")
	client := &http.Client{}
	data := url.Values{}
	data.Set("token", accessToken)
	request, _ := http.NewRequest("POST", authUrl, bytes.NewBufferString(data.Encode()))
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	response, err := client.Do(request)
	if err != nil {
		logging.Error("Error talking to auth server", err)
		pufferdHttp.Respond(gin).Message(err.Error()).Fail().Status(500).Send()
		gin.Abort()
		return false
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		logging.Error("Unexpected response code from auth server", response.StatusCode)
		pufferdHttp.Respond(gin).Message(fmt.Sprintf("unexpected response code %d", response.StatusCode)).Fail().Status(500).Send()
		gin.Abort()
		return false
	}
	var respArr map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&respArr)

	if  err != nil {
		logging.Error("Error parsing response from auth server", err)
		pufferdHttp.Respond(gin).Message(err.Error()).Fail().Status(500).Send()
		gin.Abort()
		return false
	} else if respArr["error"] != nil {
		pufferdHttp.Respond(gin).Message(respArr["error"].(string)).Fail().Status(500).Send()
		gin.Abort()
		return false
	}

	active, ok := respArr["active"].(bool)

	if !ok || !active {
		gin.AbortWithStatus(401)
		return false
	}

	serverMapping := respArr["servers"].(map[string]interface{})

	mapping := make(map[string][]string)

	for k, v := range serverMapping {
		mapping[k] = common.ToStringArray(v)
	}

	gin.Set("serverScopes", mapping)
	return true
}
