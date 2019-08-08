/*
 Copyright 2018 Padduck, LLC

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

package template

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/apufferi/response"
	"github.com/pufferpanel/pufferd/httphandlers"
	"github.com/pufferpanel/pufferd/programs"
)

func RegisterRoutes(e *gin.Engine) {
	l := e.Group("templates")
	{
		l.GET("", httphandlers.OAuth2Handler("node.templates", false), GetTemplates)
		l.GET("/:id", httphandlers.OAuth2Handler("node.templates", false), GetTemplate)
		l.GET("/:id/readme", httphandlers.OAuth2Handler("node.templates", false), GetTemplateReadme)
		//l.POST("/:id", httphandlers.OAuth2Handler("node.templates.edit", false), EditTemplate)
	}
}

func GetTemplates(c *gin.Context) {
	response.From(c).Data(programs.GetPlugins())
}

func GetTemplate(c *gin.Context) {
	name, exists := c.GetQuery("id")
	if !exists || name == "" {
		response.From(c).Fail().Status(400).Message("no template name provided")
		return
	}
	data, err := programs.GetPlugin(name)
	if err != nil {
		if os.IsNotExist(err) {
			response.From(c).Fail().Status(404).Message("no template with provided name")
		} else {
			response.From(c).Fail().Status(500).Message("error reading template")
		}
	} else {
		response.From(c).Status(200).Data(data)
	}
}

func GetTemplateReadme(c *gin.Context) {
	name, exists := c.GetQuery("id")
	if !exists || name == "" {
		response.From(c).Fail().Status(400).Message("no template name provided")
		return
	}
	data, err := programs.GetPluginReadme(name)
	if err != nil {
		if os.IsNotExist(err) {
			response.From(c).Fail().Status(404).Message("no template readme with provided name")
		} else {
			response.From(c).Fail().Status(500).Message("error reading template readme")
		}
	} else {
		response.From(c).Status(200).Data(data)
	}
}

func EditTemplate(c *gin.Context) {
	name, exists := c.GetQuery("id")
	if !exists || name == "" {
		response.From(c).Fail().Status(400).Message("no template name provided")
		return
	}
	data, err := programs.GetPlugin(name)
	if err != nil {
		if os.IsNotExist(err) {
			response.From(c).Fail().Status(404).Message("no template with provided name")
		} else {
			response.From(c).Fail().Status(500).Message("error reading template")
		}
	} else {
		response.From(c).Status(200).Data(data)
	}
}
