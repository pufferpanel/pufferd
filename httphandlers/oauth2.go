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
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/common"
	pufferdHttp "github.com/pufferpanel/apufferi/http"
	"github.com/pufferpanel/pufferd/oauth2"
	"github.com/pufferpanel/pufferd/programs"
	"strings"
)

type oauthCache struct {
	oauthToken string
	scopes     map[string][]string
	expireTime int64
}

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

		if !oauth2.ValidateToken(authToken, gin) {
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
