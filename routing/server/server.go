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
	"io"
	"io/ioutil"
	"mime"
	gohttp "net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/itsjamie/gin-cors"
	"github.com/pkg/errors"
	"github.com/pufferpanel/apufferi/http"
	ppErrors "github.com/pufferpanel/pufferd/errors"
	"github.com/pufferpanel/pufferd/httphandlers"
	"github.com/pufferpanel/pufferd/logging"
	"github.com/pufferpanel/pufferd/programs"
	"github.com/pufferpanel/pufferd/utils"
	"strconv"
	"strings"
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
		l.POST("/:id", httphandlers.OAuth2Handler("server.edit", true), EditServer)

		l.GET("/:id/start", httphandlers.OAuth2Handler("server.start", true), StartServer)
		l.GET("/:id/stop", httphandlers.OAuth2Handler("server.stop", true), StopServer)
		l.GET("/:id/kill", httphandlers.OAuth2Handler("server.stop", true), KillServer)

		l.POST("/:id/start", httphandlers.OAuth2Handler("server.start", true), StartServer)
		l.POST("/:id/stop", httphandlers.OAuth2Handler("server.stop", true), StopServer)
		l.POST("/:id/kill", httphandlers.OAuth2Handler("server.stop", true), KillServer)

		l.POST("/:id/install", httphandlers.OAuth2Handler("server.install", true), InstallServer)

		l.GET("/:id/file/*filename", httphandlers.OAuth2Handler("server.file.get", true), GetFile)
		l.PUT("/:id/file/*filename", httphandlers.OAuth2Handler("server.file.put", true), PutFile)
		//l.DELETE("/:id/file/*filename", httphandlers.OAuth2Handler("server.file.delete", true), DeleteFile)
		l.DELETE("/:id/file/*filename", httphandlers.OAuth2Handler("server.file.put", true), DeleteFile)

		l.POST("/:id/console", httphandlers.OAuth2Handler("server.console.send", true), PostConsole)
		l.GET("/:id/console", httphandlers.OAuth2Handler("server.console", true), cors.Middleware(cors.Config{
			Origins:     "*",
			Credentials: true,
		}), GetConsole)
		l.GET("/:id/logs", httphandlers.OAuth2Handler("server.log", true), GetLogs)

		l.GET("/:id/stats", httphandlers.OAuth2Handler("server.stats", true), GetStats)

		l.POST("/:id/reload", httphandlers.OAuth2Handler("server.reload", true), ReloadServer)
	}
	e.GET("/network", httphandlers.OAuth2Handler("server.network", false), NetworkServer)
}

func StartServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	server.Start()
	http.Respond(c).Send()
}

func StopServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	wait := c.Param("wait")
	if wait == "" || (wait != "true" && wait != "false") {
		wait = "true"
	}

	err := server.Stop()
	if err != nil {
		errorConnection(c, err)
		return
	}

	if wait == "true" {
		err = server.GetEnvironment().WaitForMainProcess()
		if err != nil {
			errorConnection(c, err)
			return
		}
	}
	http.Respond(c).Send()
}

func KillServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	err := server.Kill()
	if err != nil {
		errorConnection(c, err)
		return
	}

	http.Respond(c).Send()
}

func CreateServer(c *gin.Context) {
	serverId := c.Param("id")
	prg, _ := programs.Get(serverId)

	if prg != nil {
		http.Respond(c).Status(409).Message("server already exists").Send()
		return
	}

	data := make(map[string]interface{}, 0)
	err := json.NewDecoder(c.Request.Body).Decode(&data)

	if err != nil {
		logging.Error("Error decoding JSON body", err)
		http.Respond(c).Status(400).Message("error parsing json").Data(err).Code(http.MALFORMEDJSON).Send()
		return
	}

	serverType := data["type"].(string)

	if !programs.Create(serverId, serverType, data) {
		errorConnection(c, nil)
	} else {
		http.Respond(c).Send()
	}
}

func DeleteServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(programs.Program)
	err := programs.Delete(prg.Id())
	if err != nil {
		http.Respond(c).Status(500).Data(err).Message("error deleting server").Send()
	} else {
		http.Respond(c).Send()
	}
}

func InstallServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(programs.Program)

	http.Respond(c).Send()
	go func() {
		prg.Install()
	}()
}

func EditServer(c *gin.Context) {
	item, _ := c.Get("server")
	prg := item.(programs.Program)

	data := make(map[string]interface{}, 0)
	json.NewDecoder(c.Request.Body).Decode(&data)

	prg.Edit(data)
	http.Respond(c).Send()
}

func GetFile(c *gin.Context) {

	item, _ := c.Get("server")
	server := item.(programs.Program)

	targetPath := c.Param("filename")

	targetFile := utils.JoinPath(server.GetEnvironment().GetRootDirectory(), targetPath)

	if !utils.EnsureAccess(targetFile, server.GetEnvironment().GetRootDirectory()) {
		http.Respond(c).Status(403).Message("invalid file path").Status(http.NOTAUTHORIZED).Send()
		return
	}

	info, err := os.Stat(targetFile)

	if os.IsNotExist(err) {
		errorConnection(c, err)
		return
	}

	if info.IsDir() {
		type FileDesc struct {
			Name      string `json:"name"`
			Modified  int64  `json:"modifyTime"`
			Size      int64  `json:"size,omitempty"`
			File      bool   `json:"isFile"`
			Extension string `json:"extension,omitempty"`
		}

		files, _ := ioutil.ReadDir(targetFile)
		fileNames := make([]interface{}, 0)
		if targetPath != "" && targetPath != "." && targetPath != "/" {
			newFile := &FileDesc{
				Name: "..",
				File: false,
			}
			fileNames = append(fileNames, newFile)
		}
		for _, file := range files {
			newFile := &FileDesc{
				Name: file.Name(),
				File: !file.IsDir(),
			}

			if newFile.File {
				newFile.Size = file.Size()
				newFile.Modified = file.ModTime().Unix()
				newFile.Extension = filepath.Ext(file.Name())
			}

			fileNames = append(fileNames, newFile)
		}
		http.Respond(c).Data(fileNames).Send()
	} else {
		_, err := os.Open(targetFile)
		if err != nil {
			if err == os.ErrNotExist {
				http.Respond(c).Status(404).Code(http.NOFILE).Send()
			} else {
				errorConnection(c, err)
			}
		}
		c.File(targetFile)
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

	targetFile := utils.JoinPath(server.GetEnvironment().GetRootDirectory(), targetPath)

	if !utils.EnsureAccess(targetFile, server.GetEnvironment().GetRootDirectory()) {
		http.Respond(c).Status(403).Message("invalid file path").Status(http.NOTAUTHORIZED).Send()
		return
	}

	_, mkFolder := c.GetQuery("folder")
	if mkFolder {
		err := os.Mkdir(targetFile, 0644)
		if err != nil {
			errorConnection(c, err)
		} else {
			http.Respond(c).Send()
		}
		return
	}
	file, err := os.Create(targetFile)
	defer file.Close()

	if err != nil {
		errorConnection(c, err)
		logging.Error("Error writing file", err)
		return
	}

	var sourceFile io.ReadCloser

	v := c.Request.Header.Get("Content-Type")
	if t, _, _ := mime.ParseMediaType(v); t == "multipart/form-data" {
		sourceFile, _, err = c.Request.FormFile("file")
	} else {
		sourceFile = c.Request.Body
	}

	if err != nil {
		errorConnection(c, err)
		logging.Error("Error writing file", err)
	} else {
		_, err = io.Copy(file, sourceFile)
		http.Respond(c).Send()
	}
}

func DeleteFile(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	targetPath := c.Param("filename")

	targetFile := utils.JoinPath(server.GetEnvironment().GetRootDirectory(), targetPath)

	if !utils.EnsureAccess(targetFile, server.GetEnvironment().GetRootDirectory()) {
		http.Respond(c).Status(403).Message("invalid file path").Status(http.NOTAUTHORIZED).Send()
		return
	}

	err := os.Remove(targetFile)
	if err != nil {
		errorConnection(c, err)
		logging.Error("Failed to delete file", err)
	} else {
		http.Respond(c).Send()
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
		http.Respond(c).Send()
	}
}

func GetConsole(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(programs.Program)

	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.Error("Error creating websocket", err)
		errorConnection(c, err)
		return
	}
	console, _ := program.GetEnvironment().GetConsole()
	for _, v := range console {
		conn.WriteMessage(websocket.TextMessage, []byte(v))
	}
	program.GetEnvironment().AddListener(conn)
}

func GetStats(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	results, err := server.GetEnvironment().GetStats()
	if err != nil {
		result := make(map[string]interface{})
		result["error"] = err.Error()
		_, isOffline := err.(ppErrors.ServerOffline)
		if isOffline {
			http.Respond(c).Data(result).Status(200).Send()
		} else {
			http.Respond(c).Data(result).Status(500).Send()
		}
	} else {
		http.Respond(c).Data(results).Send()
	}
}

func ReloadServer(c *gin.Context) {
	item, _ := c.Get("server")
	server := item.(programs.Program)

	err := programs.Reload(server.Id())
	if err != nil {
		errorConnection(c, err)
		return
	}
	http.Respond(c).Send()
}

func NetworkServer(c *gin.Context) {

	servers := c.DefaultQuery("ids", "")
	if servers == "" {
		http.Respond(c).Status(400).Code(http.NOSERVERID).Message("no server ids provided").Send()
		return
	}
	serverIds := strings.Split(servers, ",")
	result := make(map[string]string)
	for _, v := range serverIds {
		program, _ := programs.Get(v)
		if program == nil {
			continue
		}
		result[program.Id()] = program.GetNetwork()
	}
	http.Respond(c).Data(result).Send()
}

func GetLogs(c *gin.Context) {
	item, _ := c.Get("server")
	program := item.(programs.Program)

	time := c.DefaultQuery("time", "0")

	castedTime, ok := strconv.ParseInt(time, 10, 64)

	if ok != nil {
		c.AbortWithError(400, errors.New("Time provided is not a valid UNIX time"))
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
	http.Respond(c).Data(result).Send()
}

func errorConnection(c *gin.Context, err error) {
	http.Respond(c).Status(500).Code(http.UNKNOWN).Data(err).Message("error handling request").Send()
}
