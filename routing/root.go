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
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/apufferi/middleware"
	"github.com/pufferpanel/apufferi/response"
	"github.com/pufferpanel/pufferd/oauth2"
	"github.com/pufferpanel/pufferd/routing/server"
	"github.com/pufferpanel/pufferd/shutdown"
)

func ConfigureWeb() *gin.Engine {
	r := gin.New()
	{
		r.Use(gin.Recovery())
		r.Use(gin.LoggerWithWriter(logging.AsWriter(logging.INFO)))
		r.Use(func(c *gin.Context) {
			middleware.ResponseAndRecover(c)
		})
		RegisterRoutes(r)
		server.RegisterRoutes(r)
	}

	oauth2.RefreshToken()

	return r
}

func RegisterRoutes(e *gin.Engine) {
	e.GET("", func(c *gin.Context) {
		response.From(c).Message("pufferd is running")
	})
	e.HEAD("", func(c *gin.Context) {
		c.Status(200)
	})
	//e.GET("/_shutdown", httphandlers.OAuth2Handler("node.stop", false), Shutdown)
}

func Shutdown(c *gin.Context) {
	response.From(c).Message("shutting down")
	go shutdown.CompleteShutdown()
}
