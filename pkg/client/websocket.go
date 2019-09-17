package client

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"

	"github.com/aka-bo/loqu/pkg/util"
)

func (o *Options) dial(logger logr.Logger) {
	id := util.NewRequestID()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	addr := fmt.Sprintf("%s:%d", o.Host, o.Port)
	u := url.URL{Scheme: "ws", Host: addr, Path: "/echo"}
	logger = logger.WithValues("requestID", id, "url", u.String())
	logger.Info("connecting to url")

	headers := http.Header{
		util.KeyRequestID: []string{id},
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		logger.Error(err, "failed to connect to url")
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logger.Error(err, "read error")
				return
			}
			logger.Info("received message", "message", string(message))
		}
	}()

	minInterval := 1
	interval := o.IntervalSeconds
	if interval < minInterval {
		interval = minInterval
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	writeMessage := func(t time.Time) {
		err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		if err != nil {
			logger.Error(err, "write error", err)
			return
		}
	}

	select {
	case <-done:
		return
	default:
		writeMessage(time.Now())
	}

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			writeMessage(t)
		case <-interrupt:
			logger.Info("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logger.Error(err, "error closing the connection")
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
