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
	"github.com/pufferpanel/apufferi/http"
	"github.com/pufferpanel/pufferd/httphandlers"
	"github.com/pufferpanel/pufferd/routing/server"
	"github.com/pufferpanel/pufferd/config"
	"github.com/pufferpanel/pufferd/routing/template"
	"github.com/pufferpanel/pufferd/shutdown"
)

func ConfigureWeb() *gin.Engine{
	r := gin.New()
	{
		r.Use(gin.Recovery())
		if config.GetOrDefault("log.api", "false") == "true" {
			r.Use(httphandlers.ApiLogging())
		}
		r.Use(httphandlers.Recovery())
		RegisterRoutes(r)
		server.RegisterRoutes(r)
		template.RegisterRoutes(r)
	}

	return r
}

func RegisterRoutes(e *gin.Engine) {
	e.GET("", func(c *gin.Context) {
		http.Respond(c).Message("pufferd is running").Send()
	})
	e.GET("/templates", template.GetTemplates)
	e.GET("/_shutdown", httphandlers.OAuth2Handler("node.stop", false), Shutdown)
}

func Shutdown(c *gin.Context) {
	http.Respond(c).Message("shutting down").Send()
	go func() {
		shutdown.CompleteShutdown()
	}()
}