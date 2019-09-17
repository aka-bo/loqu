package server

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/aka-bo/loqu/pkg/util"
)

//Echo handler wrapper
type Echo struct {
	serverInfo serverInfo

	stopChan chan bool
	upgrader websocket.Upgrader
}

//Start the Echo... echo... echo
func (e *Echo) Start() {
	e.stopChan = make(chan bool, 1)
	e.upgrader = websocket.Upgrader{}
}

//Stop signals that the shutdown process has begun
func (e *Echo) Stop() {
	e.serverInfo.Stopping = true
	e.stopChan <- true
}

//Handle websocket requests by replying with the received message
func (e *Echo) Handle(w http.ResponseWriter, r *http.Request) {
	logger := util.WithID("Handle", r).WithValues("path", r.URL.Path)
	logger.Info("Handling request")

	c, err := e.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(err, "websocket upgrade failed")
		return
	}

	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if ce, ok := err.(*websocket.CloseError); ok {
				logger.Error(err, "connection closed", "code", ce.Code)
				break
			}
			logger.Error(err, "read failed")
			break
		}
		if logger.V(4).Enabled() {
			logger.Info("message received", "message", string(message), "messageType", messageTypeString(mt))
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			logger.Error(err, "write failed")
			break
		}
	}
}

func messageTypeString(messageType int) string {
	mt := "Unknown"

	switch messageType {
	case websocket.TextMessage:
		mt = "Text"
	case websocket.BinaryMessage:
		mt = "Binary"
	case websocket.CloseMessage:
		mt = "Close"
	case websocket.PingMessage:
		mt = "Ping"
	case websocket.PongMessage:
		mt = "Pong"
	}
	return mt
}
