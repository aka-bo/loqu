package server

import (
	"html/template"
	"net/http"

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
	id := uuid.New()
	glog.V(3).Infof("[%s] Handling %s", id, r.URL.Path)

	c, err := e.upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("[%s] websocket upgrade failed: %v", id, err)
		return
	}

	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if ce, ok := err.(*websocket.CloseError); ok {
				glog.Errorf("[%s] connection closed with %v", id, ce.Code)
				break
			}
			glog.Errorf("[%s] read failed: %v", id, err)
			break
		}
		if glog.V(4) {
			echo := &struct {
				Message     string    `json:"message"`
				MessageType string    `json:"messageType"`
				Info        *response `json:"info"`
			}{
				Message:     string(message),
				MessageType: messageTypeString(mt),
				Info:        buildResponse(id, &e.serverInfo, r),
			}
			b, _ := marshal(echo, false)
			glog.Infof("[%s] echo received: %s", id, b)
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			glog.Errorf("[%s] write failed: %v", id, err)
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

func demoHandler() http.Handler {
	var demoTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.V(3).Infof("Handling %s", r.URL.Path)

		demoTemplate.Execute(w, "ws://"+r.Host+"/echo")
	})
}
