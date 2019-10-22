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
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/itsjamie/gin-cors"
	"github.com/pufferpanel/apufferi/v4"
	"github.com/pufferpanel/apufferi/v4/logging"
	"github.com/pufferpanel/apufferi/v4/response"
	"github.com/pufferpanel/apufferi/v4/scope"
	"github.com/pufferpanel/pufferd/v2"
	"github.com/pufferpanel/pufferd/v2/httphandlers"
	"github.com/pufferpanel/pufferd/v2/messages"
	"github.com/pufferpanel/pufferd/v2/programs"
	"github.com/satori/go.uuid"
	"github.com/spf13/cast"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
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
		l.OPTIONS("/:id/console", response.CreateOptions("CONNECT"))

		l.PUT("/:id", httphandlers.OAuth2Handler(scope.ServersCreate, false), CreateServer)
		l.DELETE("/:id", httphandlers.OAuth2Handler(scope.ServersDelete, true), DeleteServer)
		l.GET("/:id", httphandlers.OAuth2Handler(scope.ServersEditAdmin, true), GetServerAdmin)
		l.POST("/:id", httphandlers.OAuth2Handler(scope.ServersEditAdmin, true), EditServerAdmin)
		l.OPTIONS("/:id", response.CreateOptions("PUT", "DELETE", "GET", "POST"))

		l.GET("/:id/data", httphandlers.OAuth2Handler(scope.ServersEdit, true), GetServer)
		l.POST("/:id/data", httphandlers.OAuth2Handler(scope.ServersEdit, true), EditServer)
		l.OPTIONS("/:id/data", response.CreateOptions("GET", "POST"))

		l.POST("/:id/reload", httphandlers.OAuth2Handler(scope.ServersEditAdmin, true), ReloadServer)
		l.OPTIONS("/:id/reload", response.CreateOptions("POST"))

		l.POST("/:id/start", httphandlers.OAuth2Handler(scope.ServersStart, true), StartServer)
		l.OPTIONS("/:id/start", response.CreateOptions("POST"))

		l.POST("/:id/stop", httphandlers.OAuth2Handler(scope.ServersStop, true), StopServer)
		l.OPTIONS("/:id/stop", response.CreateOptions("POST"))

		l.POST("/:id/kill", httphandlers.OAuth2Handler(scope.ServersStop, true), KillServer)
		l.OPTIONS("/:id/kill", response.CreateOptions("POST"))

		l.POST("/:id/install", httphandlers.OAuth2Handler(scope.ServersInstall, true), InstallServer)
		l.OPTIONS("/:id/install", response.CreateOptions("POST"))

		l.GET("/:id/file/*filename", httphandlers.OAuth2Handler(scope.ServersFilesGet, true), GetFile)
		l.PUT("/:id/file/*filename", httphandlers.OAuth2Handler(scope.ServersFilesPut, true), PutFile)
		l.DELETE("/:id/file/*filename", httphandlers.OAuth2Handler(scope.ServersFilesPut, true), DeleteFile)
		l.OPTIONS("/:id/file/*filename", response.CreateOptions("GET", "PUT", "DELETE"))

		l.POST("/:id/console", httphandlers.OAuth2Handler(scope.ServersConsoleSend, true), PostConsole)
		l.GET("/:id/console", httphandlers.OAuth2Handler(scope.ServersConsole, true), cors.Middleware(cors.Config{
			Origins:     "*",
			Credentials: true,
		}), GetConsole)
		l.OPTIONS("/:id/console", response.CreateOptions("GET", "POST"))

		l.GET("/:id/logs", httphandlers.OAuth2Handler(scope.ServersConsole, true), GetLogs)
		l.OPTIONS("/:id/logs", response.CreateOptions("GET"))

		l.GET("/:id/stats", httphandlers.OAuth2Handler(scope.ServersStat, true), GetStats)
		l.OPTIONS("/:id/stats", response.CreateOptions("GET"))

		l.GET("/:id/status", httphandlers.OAuth2Handler(scope.ServersStat, true), GetStatus)
		l.OPTIONS("/:id/status", response.CreateOptions("GET"))

		l.GET("/:id/socket", httphandlers.OAuth2Handler(scope.ServersConsole, true), cors.Middleware(cors.Config{
			Origins:     "*",
			Credentials: true,
		}), OpenSocket)
		l.OPTIONS("/:id/socket", response.CreateOptions("GET"))
	}

	l.POST("", httphandlers.OAuth2Handler(scope.ServersCreate, false), CreateServer)
	l.OPTIONS("", response.CreateOptions("POST"))
}

// StartServer godoc
// @Summary Starts server
// @Description Starts the given server
// @Accept json
// @Produce json
// @Success 200
// @Failure 400 {object} response.Error
// @Failure 403 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Param id path string true "Server Identifier"
// @securitydefinitions.oauth2.application OAuth2Application
// @scope.server.start
// @Router /server/{id} [post]
func StartServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	_, wait := c.GetQuery("wait")

	if wait {
		err := server.Start()
		if response.HandleError(c, err, http.StatusInternalServerError) {
		} else {
			c.Status(http.StatusNoContent)
		}
	} else {
		go func() {
			_ = server.Start()
		}()
		c.Status(http.StatusAccepted)
	}
}

func StopServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	_, wait := c.GetQuery("wait")

	err := server.Stop()
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	if wait {
		err = server.GetEnvironment().WaitForMainProcess()
		if response.HandleError(c, err, http.StatusInternalServerError) {
		} else {
			c.Status(http.StatusNoContent)
		}
	} else {
		c.Status(204)
	}
}

func KillServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	err := server.Kill()
	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		c.Status(http.StatusNoContent)
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
		response.HandleError(c, pufferd.ErrServerAlreadyExists, http.StatusConflict)
		return
	}

	prg = &programs.Program{}
	err := json.NewDecoder(c.Request.Body).Decode(prg)

	if err != nil {
		logging.Exception("error decoding JSON body", err)
		response.HandleError(c, err, http.StatusBadRequest)
		return
	}

	prg.Identifier = serverId

	if err := programs.Create(prg); err != nil {
		response.HandleError(c, err, http.StatusInternalServerError)
	} else {
		c.JSON(200, &pufferd.ServerIdResponse{Id: serverId})
	}
}

func DeleteServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)
	err := programs.Delete(prg.Id())
	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		c.Status(http.StatusNoContent)
	}
}

func InstallServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	go func(p *programs.Program) {
		_ = p.Install()
	}(prg)

	c.Status(http.StatusAccepted)
}

func EditServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	data := &pufferd.ServerData{}
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if response.HandleError(c, err, http.StatusBadRequest) {
		return
	}

	err = prg.Edit(data.Variables, false)
	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		c.Status(http.StatusNoContent)
	}
}

func EditServerAdmin(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	data := &pufferd.ServerData{}
	err := json.NewDecoder(c.Request.Body).Decode(&data)
	if response.HandleError(c, err, http.StatusBadRequest) {
		return
	}

	err = prg.Edit(data.Variables, true)
	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		c.Status(http.StatusNoContent)
	}
}

func ReloadServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	err := programs.Reload(prg.Id())
	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		c.Status(http.StatusNoContent)
	}
}

func GetServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	data := server.GetData()

	c.JSON(200, &pufferd.ServerData{Variables: data})
}

func GetServerAdmin(c *gin.Context) {
	item, _ := c.MustGet("server").(*apufferi.Server)

	c.JSON(200, &pufferd.ServerDataAdmin{Server: item})
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
		if os.IsNotExist(err) {
			c.AbortWithStatus(404)
		} else if err == pufferd.ErrIllegalFileAccess {
			response.HandleError(c, err, http.StatusBadRequest)
		} else {
			response.HandleError(c, err, http.StatusInternalServerError)
		}
		return
	}

	if data.FileList != nil {
		c.JSON(200, data.FileList)
	} else if data.Contents != nil {
		fileName := filepath.Base(data.Name)

		extraHeaders := map[string]string{
			"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, fileName),
		}

		//discard the built-in response, we cannot use this one at all
		c.DataFromReader(http.StatusOK, data.ContentLength, "application/octet-stream", data.Contents, extraHeaders)
	} else {
		//uhhhhhhhhhhhhh
		response.HandleError(c, errors.New("no file content or file list"), http.StatusInternalServerError)
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
		response.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	var sourceFile io.ReadCloser

	v := c.Request.Header.Get("Content-Type")
	if t, _, _ := mime.ParseMediaType(v); t == "multipart/form-data" {
		sourceFile, _, err = c.Request.FormFile("file")
		if response.HandleError(c, err, http.StatusInternalServerError) {
			return
		}
	} else {
		sourceFile = c.Request.Body
	}

	file, err := server.OpenFile(targetPath)
	defer apufferi.Close(file)
	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		_, err = io.Copy(file, sourceFile)
		if response.HandleError(c, err, http.StatusInternalServerError) {
		} else {
			c.Status(http.StatusNoContent)
		}
	}
}

func DeleteFile(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(*programs.Program)

	targetPath := c.Param("filename")

	err := server.DeleteItem(targetPath)
	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		c.Status(http.StatusNoContent)
	}
}

func PostConsole(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(*programs.Program)

	d, _ := ioutil.ReadAll(c.Request.Body)
	cmd := string(d)
	err := prg.Execute(cmd)
	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		c.Status(http.StatusNoContent)
	}
}

func GetConsole(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(*programs.Program)

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	console, _ := program.GetEnvironment().GetConsole()
	_ = messages.Write(conn, messages.ConsoleMessage{Logs: console})

	program.GetEnvironment().AddListener(conn)
}

func GetStats(c *gin.Context) {
	item, _ := c.Get("server")
	svr := item.(*programs.Program)

	results, err := svr.GetEnvironment().GetStats()
	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		c.JSON(200, results)
	}
}

func GetLogs(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(*programs.Program)

	time := c.DefaultQuery("time", "0")

	castedTime, ok := cast.ToInt64E(time)
	if ok != nil {
		response.HandleError(c, pufferd.ErrInvalidUnixTime, http.StatusBadRequest)
		return
	}

	console, epoch := program.GetEnvironment().GetConsoleFrom(castedTime)
	msg := ""
	for _, k := range console {
		msg += k
	}

	c.JSON(200, &pufferd.ServerLogs{
		Epoch: epoch,
		Logs:  msg,
	})
}

func GetStatus(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(*programs.Program)

	running, err := program.IsRunning()

	if response.HandleError(c, err, http.StatusInternalServerError) {
	} else {
		c.JSON(200, &pufferd.ServerRunning{Running: running})
	}
}

func OpenSocket(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(*programs.Program)

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	console, _ := program.GetEnvironment().GetConsole()
	_ = messages.Write(conn, messages.ConsoleMessage{Logs: console})

	internalMap, _ := c.Get("scopes")
	scopes := internalMap.([]scope.Scope)

	go listenOnSocket(conn, program, scopes)

	program.GetEnvironment().AddListener(conn)
}
