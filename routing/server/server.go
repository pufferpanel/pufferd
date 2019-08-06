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

package server

import (
	"encoding/json"
	"fmt"
	"github.com/pufferpanel/apufferi"
	"github.com/pufferpanel/apufferi/response"
	"github.com/pufferpanel/pufferd/errors"
	"io"
	"io/ioutil"
	"mime"
	gohttp "net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/itsjamie/gin-cors"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/httphandlers"
	"github.com/pufferpanel/pufferd/programs"

	"github.com/pufferpanel/pufferd/messages"
	"github.com/satori/go.uuid"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *gohttp.Request) bool {
		return true
	},
}

func RegisterRoutes(e *gin.Engine) {
	l := e.Group("/server")
	{
		l.Handle("CONNECT", "/:id/console", func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Credentials", "false")
		})
		l.PUT("/:id", httphandlers.OAuth2Handler("server.create", false), CreateServer)
		l.DELETE("/:id", httphandlers.OAuth2Handler("server.delete", true), DeleteServer)

		l.GET("/:id", httphandlers.OAuth2Handler("server.edit.admin", true), GetServerAdmin)
		l.POST("/:id", httphandlers.OAuth2Handler("server.edit.admin", true), EditServerAdmin)

		l.GET("/:id/data", httphandlers.OAuth2Handler("server.edit", true), GetServer)
		l.POST("/:id/data", httphandlers.OAuth2Handler("server.edit", true), EditServer)

		l.POST("/:id/reload", httphandlers.OAuth2Handler("server.edit.admin", true), ReloadServer)

		l.GET("/:id/start", httphandlers.OAuth2Handler("server.start", true), StartServer)
		l.GET("/:id/stop", httphandlers.OAuth2Handler("server.stop", true), StopServer)
		l.GET("/:id/kill", httphandlers.OAuth2Handler("server.stop", true), KillServer)

		l.POST("/:id/start", httphandlers.OAuth2Handler("server.start", true), StartServer)
		l.POST("/:id/stop", httphandlers.OAuth2Handler("server.stop", true), StopServer)
		l.POST("/:id/kill", httphandlers.OAuth2Handler("server.stop", true), KillServer)

		l.POST("/:id/install", httphandlers.OAuth2Handler("server.install", true), InstallServer)

		l.GET("/:id/file/*filename", httphandlers.OAuth2Handler("server.files", true), GetFile)
		l.PUT("/:id/file/*filename", httphandlers.OAuth2Handler("server.files.put", true), PutFile)
		l.DELETE("/:id/file/*filename", httphandlers.OAuth2Handler("server.files.delete", true), DeleteFile)

		l.POST("/:id/console", httphandlers.OAuth2Handler("server.console.send", true), PostConsole)
		l.GET("/:id/console", httphandlers.OAuth2Handler("server.console", true), cors.Middleware(cors.Config{
			Origins:     "*",
			Credentials: true,
		}), GetConsole)
		l.GET("/:id/logs", httphandlers.OAuth2Handler("server.console", true), GetLogs)

		l.GET("/:id/stats", httphandlers.OAuth2Handler("server.stats", true), GetStats)
		l.GET("/:id/status", httphandlers.OAuth2Handler("server.stats", true), GetStatus)

		l.GET("/:id/socket", httphandlers.OAuth2Handler("server.console", true), cors.Middleware(cors.Config{
			Origins:     "*",
			Credentials: true,
		}), OpenSocket)
	}
	l.POST("", httphandlers.OAuth2Handler("server.create", false), CreateServer)
	e.GET("/network", httphandlers.OAuth2Handler("server.network", false), NetworkServer)
}

func StartServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	err := server.Start()
	if err != nil {
		response.Respond(c).Status(500).Data(err).Message("error starting server").Send()
	} else {
		response.Respond(c).Send()
	}
}

func StopServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	_, wait := c.GetQuery("wait")

	err := server.Stop()
	if err != nil {
		errorConnection(c, err)
		return
	}

	if wait {
		err = server.GetEnvironment().WaitForMainProcess()
		if err != nil {
			errorConnection(c, err)
			return
		}
	}
	response.Respond(c).Send()
}

func KillServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	err := server.Kill()
	if err != nil {
		errorConnection(c, err)
		return
	}

	response.Respond(c).Send()
}

func CreateServer(c *gin.Context) {
	serverId := c.Param("id")
	if serverId == "" {
		id := uuid.NewV4()
		serverId = id.String()
	}
	prg, _ := programs.Get(serverId)

	if prg != nil {
		response.Respond(c).Status(409).Message("server already exists").Send()
		return
	}

	prg = &programs.Program{}
	err := json.NewDecoder(c.Request.Body).Decode(prg)

	if err != nil {
		logging.Exception("error decoding JSON body", err)
		response.Respond(c).Status(400).Message("error parsing json").Data(err).Send()
		return
	}

	prg.Identifier = serverId

	if !programs.Create(prg) {
		errorConnection(c, nil)
	} else {
		data := make(map[string]interface{})
		data["id"] = serverId
		response.Respond(c).Data(data).Send()
	}
}

func DeleteServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(programs.Program)
	err := programs.Delete(prg.Id())
	if err != nil {
		response.Respond(c).Status(500).Data(err).Message("error deleting server").Send()
	} else {
		response.Respond(c).Send()
	}
}

func InstallServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(programs.Program)

	go func(p programs.Program) {
		_ = p.Install()
	}(prg)
	response.Respond(c).Send()
}

func EditServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(programs.Program)

	data := make(map[string]interface{}, 0)
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		response.Respond(c).Status(500).Data(err).Message("error editing server").Send()
	}

	err = prg.Edit(data, false)

	if err != nil {
		response.Respond(c).Status(500).Data(err).Message("error editing server").Send()
	} else {
		response.Respond(c).Send()
	}
}

func EditServerAdmin(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(programs.Program)

	data := &admin{}
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		response.Respond(c).Status(500).Data(err).Message("error editing server").Send()
	}

	err = prg.Edit(data.Data, true)

	if err != nil {
		response.Respond(c).Status(500).Data(err).Message("error editing server").Send()
	} else {
		response.Respond(c).Send()
	}
}

func ReloadServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(programs.Program)

	err := programs.Reload(prg.Id())
	if err != nil {
		response.Respond(c).Status(500).Data(err).Message("error reloading server").Send()
	} else {
		response.Respond(c).Send()
	}
}

func GetServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	data := server.GetData()
	result := make(map[string]interface{}, 0)
	result["data"] = data

	response.Respond(c).Data(data).Send()
}

func GetServerAdmin(c *gin.Context) {
	item, _ := c.Get("server")

	response.Respond(c).Data(item).Send()
}

func GetFile(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	targetPath := c.Param("filename")
	logging.Debug("Getting following file: %s", targetPath)

	data, err := server.GetItem(targetPath)
	defer apufferi.Close(data.Contents)

	if err != nil {
		answer := response.Respond(c).Fail()
		if os.IsNotExist(err) {
			answer.Status(404)
		} else if err == errors.ErrIllegalFileAccess {
			answer.Status(500).Error(err)
		} else {
			answer.Status(500).Error(err)
		}
		answer.Send()
		return
	}

	if data.FileList != nil {
		response.Respond(c).Data(data.FileList).Send()
	} else if data.Contents != nil {
		fileName := filepath.Base(data.Name)

		extraHeaders := map[string]string{
			"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, fileName),
		}

		c.DataFromReader(gohttp.StatusOK, data.ContentLength, "application/octet-stream", data.Contents, extraHeaders)
	} else {
		//uhhhhhhhhhhhhh
		response.Respond(c).Fail().Send()
	}
}

func PutFile(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	targetPath := c.Param("filename")

	if targetPath == "" {
		c.Status(404)
		return
	}

	var err error

	_, mkFolder := c.GetQuery("folder")
	if mkFolder {
		err = server.CreateFolder(targetPath)
		if err != nil {
			errorConnection(c, err)
		} else {
			response.Respond(c).Send()
		}
		return
	}

	var sourceFile io.ReadCloser

	v := c.Request.Header.Get("Content-Type")
	if t, _, _ := mime.ParseMediaType(v); t == "multipart/form-data" {
		sourceFile, _, err = c.Request.FormFile("file")
		if err != nil {
			errorConnection(c, err)
			return
		}
	} else {
		sourceFile = c.Request.Body
	}

	file, err := server.OpenFile(targetPath)
	defer apufferi.Close(file)
	if err != nil {
		errorConnection(c, err)
	} else {
		_, err = io.Copy(file, sourceFile)
		if err != nil {
			errorConnection(c, err)
		} else {
			response.Respond(c).Send()
		}
	}
}

func DeleteFile(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	targetPath := c.Param("filename")

	err := server.DeleteItem(targetPath)
	if err != nil {
		errorConnection(c, err)
	} else {
		response.Respond(c).Send()
	}
}

func PostConsole(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(programs.Program)

	d, _ := ioutil.ReadAll(c.Request.Body)
	cmd := string(d)
	err := prg.Execute(cmd)
	if err != nil {
		errorConnection(c, err)
	} else {
		response.Respond(c).Send()
	}
}

func GetConsole(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(programs.Program)

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Exception("error creating websocket", err)
		errorConnection(c, err)
		return
	}

	console, _ := program.GetEnvironment().GetConsole()
	_ = messages.Write(conn, messages.ConsoleMessage{Logs: console})

	program.GetEnvironment().AddListener(conn)
}

func GetStats(c *gin.Context) {
	item, _ := c.Get("server")
	svr := item.(programs.Program)

	results, err := svr.GetEnvironment().GetStats()
	if err != nil {
		result := make(map[string]interface{})
		result["memory"] = 0
		result["cpu"] = 0
		response.Respond(c).Data(result).Status(200).Send()
	} else {
		response.Respond(c).Data(results).Send()
	}
}

func NetworkServer(c *gin.Context) {
	s := c.DefaultQuery("ids", "")
	if s == "" {
		response.Respond(c).Status(400).Message("no server ids provided").Send()
		return
	}
	ids := strings.Split(s, ",")
	result := make(map[string]string)
	for _, v := range ids {
		program, _ := programs.Get(v)
		if program == nil {
			continue
		}
		result[program.Id()] = program.GetNetwork()
	}
	response.Respond(c).Data(result).Send()
}

func GetLogs(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(programs.Program)

	time := c.DefaultQuery("time", "0")

	castedTime, ok := strconv.ParseInt(time, 10, 64)

	if ok != nil {
		//c.AbortWithError(400, errors.New("Time provided is not a valid UNIX time"))
		response.Respond(c).Message("time provided is not a valid UNIX time").Send()
		return
	}

	console, epoch := program.GetEnvironment().GetConsoleFrom(castedTime)
	msg := ""
	for _, k := range console {
		msg += k
	}
	result := make(map[string]interface{})
	result["epoch"] = epoch
	result["logs"] = msg
	response.Respond(c).Data(result).Send()
}

func GetStatus(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(programs.Program)

	running, err := program.IsRunning()
	result := make(map[string]interface{})

	if err != nil {
		result["error"] = err.Error()
		response.Respond(c).Data(result).Status(500).Send()
	} else {
		result["running"] = running
		response.Respond(c).Data(result).Send()
	}
}

func OpenSocket(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(programs.Program)

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Exception("error creating websocket", err)
		errorConnection(c, err)
		return
	}

	console, _ := program.GetEnvironment().GetConsole()
	_ = messages.Write(conn, messages.ConsoleMessage{Logs: console})

	internalMap, _ := c.Get("serverScopes")
	scopes := internalMap.(map[string][]string)

	go listenOnSocket(conn, program, scopes[program.Id()])

	program.GetEnvironment().AddListener(conn)
}

func errorConnection(c *gin.Context, err error) {
	logging.Exception("error on API call", err)
	response.Respond(c).Status(500).Data(err).Message("error handling request").Send()
}

type admin struct {
	Data map[string]interface{} `json:"data"`
}
