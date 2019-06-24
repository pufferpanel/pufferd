package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/pufferpanel/apufferi"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/messages"
	"github.com/pufferpanel/pufferd/programs"
	"reflect"
	"strings"
)

func listenOnSocket(conn *websocket.Conn, server programs.Program, scopes []string) {
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			logging.Exception("error on reading from websocket", err)
			return
		}
		if msgType != websocket.TextMessage {
			continue
		}
		mapping := make(map[string]interface{})

		err = json.Unmarshal(data, &mapping)
		if err != nil {
			logging.Exception("error on decoding websocket message", err)
			continue
		}

		messageType := mapping["type"]
		if message, ok := messageType.(string); ok {
			switch strings.ToLower(message) {
			case "stat":
				{
					if apufferi.ContainsValue(scopes, "server.stats") {
						results, err := server.GetEnvironment().GetStats()
						msg := messages.StatMessage{}
						if err != nil {
							msg.Cpu = 0
							msg.Memory = 0
						} else {
							msg.Cpu, _ = results["cpu"].(float64)
							msg.Memory, _ = results["memory"].(float64)
						}
						_ = conn.WriteJSON(&messages.Transmission{Message: msg, Type: msg.Key()})
					}
				}
			case "start":
				{
					if apufferi.ContainsValue(scopes, "server.start") {
						_ = server.Start()
					}
					break
				}
			case "stop":
				{
					if apufferi.ContainsValue(scopes, "server.stop") {
						_ = server.Stop()
					}
				}
			case "install":
				{
					if apufferi.ContainsValue(scopes, "server.install") {
						_ = server.Install()
					}
				}
			case "kill":
				{
					if apufferi.ContainsValue(scopes, "server.kill") {
						_ = server.Kill()
					}
				}
			case "reload":
				{
					if apufferi.ContainsValue(scopes, "server.reload") {
						_ = programs.Reload(server.Id())
					}
				}
			case "ping":
				{
					_ = conn.WriteJSON(map[string]string{"ping": "pong"})
				}
			case "console":
				{
					cmd, ok := mapping["command"].(string)
					if ok {
						if run, _ := server.IsRunning(); run {
							_ = server.GetEnvironment().ExecuteInMainProcess(cmd)
						}
					}
				}
			case "file":
				{
					action, ok := mapping["action"].(string)
					if !ok {
						break
					}
					path, ok := mapping["path"].(string)
					if !ok {
						break
					}

					switch strings.ToLower(action) {
					case "get":
						{
							file, list, err := server.GetFile(path)
							if err != nil {
								_ = conn.WriteJSON(map[string]string{"error": err.Error()})
							}

							if list != nil {
								_ = conn.WriteJSON(list)
							} else if file != nil {

							}
						}
					}
				}
			}
		} else {
			logging.Error("message type is not a string, but was %s", reflect.TypeOf(messageType))
		}
	}
}
