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
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/v3"
	"github.com/pufferpanel/apufferi/v3/logging"
	"github.com/pufferpanel/apufferi/v3/response"
	"github.com/pufferpanel/apufferi/v3/scope"
	"github.com/pufferpanel/pufferd/v2/config"
	"github.com/pufferpanel/pufferd/v2/programs"
	"io"
	"os"
	"path"
	"strings"
)

type oauthCache struct {
	oauthToken string
	scopes     map[string][]string
	expireTime int64
}

func OAuth2Handler(requiredScope scope.Scope, requireServer bool) gin.HandlerFunc {
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
				response.Respond(gin).Fail().Status(400).Message("no access token provided").Send()
				gin.Abort()
				return
			}
		} else {
			authArr := strings.SplitN(authHeader, " ", 2)
			if len(authArr) < 2 || authArr[0] != "Bearer" {
				response.Respond(gin).Fail().Status(400).Message("invalid access token format").Send()
				gin.Abort()
				return
			}
			authToken = authArr[1]
		}

		f, err := os.OpenFile(path.Join(config.Get().Data.BasePath, "public.pem"), os.O_RDONLY, 660)
		defer apufferi.Close(f)
		if err != nil {
			logging.Exception("Error handling oauth2 validation", err)
			response.Respond(gin).Fail().Status(500).Message("error validating access token").Send()
			return
		}

		var buf bytes.Buffer

		_, _ = io.Copy(&buf, f)

		block, _ := pem.Decode(buf.Bytes())
		if block == nil {
			logging.Exception("Error handling oauth2 validation", errors.New("public key is not in PEM format"))
			response.Respond(gin).Fail().Status(500).Message("error validating access token").Send()
			return
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			logging.Exception("Error handling oauth2 validation", err)
			response.Respond(gin).Fail().Status(500).Message("error validating access token").Send()
			return
		}

		pubKey, ok := pub.(*ecdsa.PublicKey)
		if !ok {
			logging.Exception("Error handling oauth2 validation", err)
			response.Respond(gin).Fail().Status(500).Message("error validating access token").Send()
			return
		}

		token, err := apufferi.ParseToken(pubKey, authToken)
		if err != nil {
			if err != jwt.ErrSignatureInvalid {
				logging.Exception("Error handling oauth2 validation", err)
			}
			response.Respond(gin).Fail().Status(403).Message("invalid access").Send()
			return
		}

		serverId := gin.GetString("serverId")
		scopes := make([]scope.Scope, 0)
		if token.Claims.PanelClaims.Scopes[serverId] != nil {
			scopes = append(scopes, token.Claims.PanelClaims.Scopes[serverId]...)
		}
		if token.Claims.PanelClaims.Scopes[""] != nil {
			scopes = append(scopes, token.Claims.PanelClaims.Scopes[""]...)
		}

		if !apufferi.Contains(scopes, requiredScope) {
			response.Respond(gin).Fail().Status(403).Message(fmt.Sprintf("missing scope %s", requiredScope)).Send()
			return
		}

		if requireServer {
			program, _ := programs.Get(serverId)
			if program == nil {
				response.Respond(gin).Fail().Status(404).Message("no server with id " + serverId).Send()
				return
			}

			gin.Set("server", program)
		}

		gin.Set("scopes", scopes)

		failure = false
	}
}
