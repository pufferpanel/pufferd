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
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/pufferd/config"
	"github.com/pufferpanel/pufferd/logging"
	"time"
	"github.com/pufferpanel/pufferd/utils"
	pufferdHttp "github.com/pufferpanel/apufferi/http"
	"github.com/pufferpanel/pufferd/programs"
)

type oauthCache struct {
	oauthToken string
	serverId   string
	scopes     []string
	expireTime int64
}

var cache = make([]*oauthCache, 20)

func OAuth2Handler(scope string, requireServer bool) gin.HandlerFunc {
	return func(gin *gin.Context) {
		authHeader := gin.Request.Header.Get("Authorization")
		var authToken string
		if authHeader == "" {
			authToken = gin.Query("accessToken")
			if authToken == "" {
				pufferdHttp.Respond(gin).Fail().Code(pufferdHttp.NOTAUTHORIZED).Status(400).Message("no access token provided");
				gin.Abort()
				return
			}
		} else {
			authArr := strings.SplitN(authHeader, " ", 2)
			if len(authArr) < 2 || authArr[0] != "Bearer" {
				pufferdHttp.Respond(gin).Code(pufferdHttp.NOTAUTHORIZED).Fail().Status(400).Message("invalid access token format");
				gin.Abort()
				return
			}
			authToken = authArr[1]
		}

		cached := isCachedRequest(authToken)

		if cached != nil {
			gin.Set("server_id", cached.serverId)
			gin.Set("scopes", cached.scopes)
		} else {
			validateToken(authToken, gin)
		}

		rawScopes, _ := gin.Get("scopes")

		if (scope != "") {
			scopes := rawScopes.([]string)
			if (!utils.ContainsValue(scopes, scope)) {
				pufferdHttp.Respond(gin).Fail().Status(403).Code(pufferdHttp.NOTAUTHORIZED).Message("missing scope " + scope).Send()
				gin.Abort()
				return
			}
		}

		if requireServer {
			serverId := gin.Param("id")
			canAccessId, _ := gin.Get("server_id")

			accessId := canAccessId.(string)

			var program programs.Program

			if accessId == "*" {
				program, _ = programs.Get(serverId)
			} else {
				program, _ = programs.Get(accessId)
			}

			if program == nil {
				pufferdHttp.Respond(gin).Fail().Status(404).Code(pufferdHttp.NOSERVER).Message("no server with id " + serverId).Send()
				gin.Abort()
				return
			}

			if accessId != program.Id() && accessId != "*" {
				pufferdHttp.Respond(gin).Fail().Status(403).Code(pufferdHttp.NOTAUTHORIZED).Message("invalid server access").Send()
				gin.Abort()
				return
			}

			gin.Set("server", program)
		}
	}
}

func validateToken(accessToken string, gin *gin.Context) {
	authUrl := config.Get("infoserver")
	token := config.Get("authtoken")
	client := &http.Client{}
	data := url.Values{}
	data.Set("token", accessToken)
	request, _ := http.NewRequest("POST", authUrl, bytes.NewBufferString(data.Encode()))
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil {
		logging.Error("Error talking to auth server", err)
		errMsg := make(map[string]string)
		errMsg["error"] = err.Error()
		gin.JSON(500, errMsg)
		return
	}
	if response.StatusCode != 200 {
		logging.Error("Unexpected response code from auth server", response.StatusCode)
		errMsg := make(map[string]string)
		errMsg["error"] = fmt.Sprintf("Received response %i", response.StatusCode)
		gin.JSON(500, errMsg)
		return
	}
	var respArr map[string]interface{}
	json.NewDecoder(response.Body).Decode(&respArr)
	logging.Debugf("%+v", respArr)
	if respArr["error"] != nil {
		logging.Error("Error parsing response from auth server", err)
		errMsg := make(map[string]string)
		errMsg["error"] = "Failed to parse auth server response"
		gin.JSON(500, errMsg)
		return
	}
	if respArr["active"].(bool) == false {
		gin.AbortWithStatus(401)
		return
	}

	serverId := respArr["server_id"].(string)
	scopes := strings.Split(respArr["scope"].(string), " ")

	cache := &oauthCache{
		oauthToken: accessToken,
		serverId:   serverId,
		scopes:     scopes,
	}
	cacheRequest(cache)

	gin.Set("server_id", serverId)
	gin.Set("scopes", scopes)
}

func isCachedRequest(accessToken string) *oauthCache {
	currentTime := time.Now().Unix()
	for k, v := range cache {
		if v == nil {
			continue
		}
		if v.oauthToken == accessToken {
			if v.expireTime < currentTime {
				return v
			}
			copy(cache[k:], cache[k+1:])
			cache[len(cache)-1] = nil
			cache = cache[:len(cache)-1]
			return nil
		}
	}
	return nil
}

func cacheRequest(request *oauthCache) {
	currentTime := time.Now().Unix()
	request.expireTime = time.Now().Add(time.Minute * 2).Unix()
	for k, v := range cache {
		if v == nil || v.expireTime > currentTime {
			cache[k] = request
			return
		}
	}
}
