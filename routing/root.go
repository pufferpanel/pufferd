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

package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/v4/logging"
	"github.com/pufferpanel/apufferi/v4/middleware"
	"github.com/pufferpanel/apufferi/v4/response"
	"github.com/pufferpanel/pufferd/v2"
	_ "github.com/pufferpanel/pufferd/v2/docs"
	"github.com/pufferpanel/pufferd/v2/oauth2"
	"github.com/pufferpanel/pufferd/v2/routing/server"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"strings"
)

// @title Pufferd API
// @version 2.0
// @description PufferPanel daemon service
// @contact.name PufferPanel
// @contact.url https://pufferpanel.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func ConfigureWeb() *gin.Engine {
	r := gin.New()
	{
		r.Use(gin.Recovery())
		r.Use(gin.LoggerWithWriter(logging.AsWriter(logging.INFO)))
		r.Use(func(c *gin.Context) {
			if c.GetHeader("Connection") == "Upgrade" {
				return
			}
			if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
				return
			}
			middleware.ResponseAndRecover(c)
		})
		RegisterRoutes(r)
		server.RegisterRoutes(r)
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	oauth2.RefreshToken()

	return r
}

func RegisterRoutes(e *gin.Engine) {
	e.GET("", func(c *gin.Context) {
		c.JSON(200, &pufferd.PufferdRunning{Message: "pufferd is running"})
	})
	e.HEAD("", func(c *gin.Context) {
		c.Status(200)
	})
	e.Handle("OPTIONS", "", response.CreateOptions("GET", "HEAD"))
}
