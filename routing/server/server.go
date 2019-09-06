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
		l.PUT("/:id", httphandlers.OAuth2Handler("servers.create", false), CreateServer)
		l.DELETE("/:id", httphandlers.OAuth2Handler("servers.delete", true), DeleteServer)

		l.GET("/:id", httphandlers.OAuth2Handler("servers.edit.admin", true), GetServerAdmin)
		l.POST("/:id", httphandlers.OAuth2Handler("servers.edit.admin", true), EditServerAdmin)

		l.GET("/:id/data", httphandlers.OAuth2Handler("servers.edit", true), GetServer)
		l.POST("/:id/data", httphandlers.OAuth2Handler("servers.edit", true), EditServer)

		l.POST("/:id/reload", httphandlers.OAuth2Handler("servers.edit.admin", true), ReloadServer)

		l.GET("/:id/start", httphandlers.OAuth2Handler("servers.start", true), StartServer)
		l.GET("/:id/stop", httphandlers.OAuth2Handler("servers.stop", true), StopServer)
		l.GET("/:id/kill", httphandlers.OAuth2Handler("servers.stop", true), KillServer)

		l.POST("/:id/start", httphandlers.OAuth2Handler("servers.start", true), StartServer)
		l.POST("/:id/stop", httphandlers.OAuth2Handler("servers.stop", true), StopServer)
		l.POST("/:id/kill", httphandlers.OAuth2Handler("servers.stop", true), KillServer)

		l.POST("/:id/install", httphandlers.OAuth2Handler("servers.install", true), InstallServer)

		l.GET("/:id/file/*filename", httphandlers.OAuth2Handler("servers.files", true), GetFile)
		l.PUT("/:id/file/*filename", httphandlers.OAuth2Handler("servers.files.put", true), PutFile)
		l.DELETE("/:id/file/*filename", httphandlers.OAuth2Handler("servers.files.delete", true), DeleteFile)

		l.POST("/:id/console", httphandlers.OAuth2Handler("servers.console.send", true), PostConsole)
		l.GET("/:id/console", httphandlers.OAuth2Handler("servers.console", true), cors.Middleware(cors.Config{
			Origins:     "*",
			Credentials: true,
		}), GetConsole)
		l.GET("/:id/logs", httphandlers.OAuth2Handler("servers.console", true), GetLogs)

		l.GET("/:id/stats", httphandlers.OAuth2Handler("servers.stats", true), GetStats)
		l.GET("/:id/status", httphandlers.OAuth2Handler("servers.stats", true), GetStatus)

		l.GET("/:id/socket", httphandlers.OAuth2Handler("servers.console", true), cors.Middleware(cors.Config{
			Origins:     "*",
			Credentials: true,
		}), OpenSocket)
	}
	l.POST("", httphandlers.OAuth2Handler("servers.create", false), CreateServer)
	e.GET("/network", httphandlers.OAuth2Handler("servers.network", false), NetworkServer)
}

func StartServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	err := server.Start()
	if err != nil {
		response.From(c).Status(500).Data(err).Message("error starting server")
	} else {
		response.From(c)
	}
}

func StopServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

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
}

func KillServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	err := server.Kill()
	if err != nil {
		errorConnection(c, err)
		return
	}
}

func CreateServer(c *gin.Context) {
	serverId := c.Param("id")
	if serverId == "" {
		id := uuid.NewV4()
		serverId = id.String()
	}
	prg, _ := programs.Get(serverId)

	if prg != nil {
		response.From(c).Status(409).Message("server already exists")
		return
	}

	prg = &programs.Program{}
	err := json.NewDecoder(c.Request.Body).Decode(prg)

	if err != nil {
		logging.Exception("error decoding JSON body", err)
		response.From(c).Status(400).Message("error parsing json").Data(err)
		return
	}

	prg.Identifier = serverId

	if !programs.Create(prg) {
		errorConnection(c, nil)
	} else {
		data := make(map[string]interface{})
		data["id"] = serverId
		response.From(c).Data(data)
	}
}

func DeleteServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)
	err := programs.Delete(prg.Id())
	if err != nil {
		response.From(c).Status(500).Data(err).Message("error deleting server")
	}
}

func InstallServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	go func(p *programs.Program) {
		_ = p.Install()
	}(prg)
}

func EditServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	data := make(map[string]interface{}, 0)
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		response.From(c).Status(500).Data(err).Message("error editing server")
		return
	}

	err = prg.Edit(data, false)

	if err != nil {
		response.From(c).Status(500).Data(err).Message("error editing server")
	}
}

func EditServerAdmin(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	data := &admin{}
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		response.From(c).Status(500).Data(err).Message("error editing server")
		return
	}

	err = prg.Edit(data.Data, true)

	if err != nil {
		response.From(c).Status(500).Data(err).Message("error editing server")
	}
}

func ReloadServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	err := programs.Reload(prg.Id())
	if err != nil {
		response.From(c).Status(500).Data(err).Message("error reloading server")
	}
}

func GetServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	data := server.GetData()
	result := make(map[string]interface{}, 0)
	result["data"] = data

	response.From(c).Data(data)
}

func GetServerAdmin(c *gin.Context) {
	item, _ := c.Get("server")

	response.From(c).Data(item)
}

func GetFile(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	targetPath := c.Param("filename")
	logging.Debug("Getting following file: %s", targetPath)

	data, err := server.GetItem(targetPath)
	defer func() {
		if data != nil {
			apufferi.Close(data.Contents)
		}
	}()

	if err != nil {
		answer := response.From(c).Fail()
		if os.IsNotExist(err) {
			answer.Status(404)
		} else if err == errors.ErrIllegalFileAccess {
			answer.Status(500).Error(err)
		} else {
			answer.Status(500).Error(err)
		}
		return
	}

	if data.FileList != nil {
		response.From(c).Data(data.FileList)
	} else if data.Contents != nil {
		fileName := filepath.Base(data.Name)

		extraHeaders := map[string]string{
			"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, fileName),
		}

		//discard the built-in response, we cannot use this one at all
		response.From(c).Discard()
		c.DataFromReader(gohttp.StatusOK, data.ContentLength, "application/octet-stream", data.Contents, extraHeaders)
	} else {
		//uhhhhhhhhhhhhh
		response.From(c).Fail()
	}
}

func PutFile(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

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
			response.From(c)
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
		}
	}
}

func DeleteFile(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	targetPath := c.Param("filename")

	err := server.DeleteItem(targetPath)
	if err != nil {
		errorConnection(c, err)
	}
}

func PostConsole(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	d, _ := ioutil.ReadAll(c.Request.Body)
	cmd := string(d)
	err := prg.Execute(cmd)
	if err != nil {
		errorConnection(c, err)
	}
}

func GetConsole(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(*programs.Program)

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Exception("error creating websocket", err)
		errorConnection(c, err)
		return
	}

	response.From(c).Discard()

	console, _ := program.GetEnvironment().GetConsole()
	_ = messages.Write(conn, messages.ConsoleMessage{Logs: console})

	program.GetEnvironment().AddListener(conn)
}

func GetStats(c *gin.Context) {
	item, _ := c.Get("server")
	svr := item.(*programs.Program)

	results, err := svr.GetEnvironment().GetStats()
	if err != nil {
		result := make(map[string]interface{})
		result["memory"] = 0
		result["cpu"] = 0
		response.From(c).Data(result)
	} else {
		response.From(c).Data(results)
	}
}

func NetworkServer(c *gin.Context) {
	s := c.DefaultQuery("ids", "")
	if s == "" {
		response.From(c).Status(400).Message("no server ids provided")
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
	response.From(c).Data(result)
}

func GetLogs(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(*programs.Program)

	time := c.DefaultQuery("time", "0")

	castedTime, ok := strconv.ParseInt(time, 10, 64)

	if ok != nil {
		//c.AbortWithError(400, errors.New("Time provided is not a valid UNIX time"))
		response.From(c).Fail().Status(400).Message("time provided is not a valid UNIX time")
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
	response.From(c).Data(result)
}

func GetStatus(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(*programs.Program)

	running, err := program.IsRunning()
	result := make(map[string]interface{})

	if err != nil {
		result["error"] = err.Error()
		response.From(c).Data(result).Status(500)
	} else {
		result["running"] = running
		response.From(c).Data(result)
	}
}

func OpenSocket(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(*programs.Program)

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
	response.From(c).Status(500).Data(err).Message("error handling request")
}

type admin struct {
	Data map[string]interface{} `json:"data"`
}
