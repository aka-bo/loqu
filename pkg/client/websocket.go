package client

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

func (o *Options) dial() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	addr := fmt.Sprintf("%s:%d", o.Host, o.Port)
	u := url.URL{Scheme: "ws", Host: addr, Path: "/echo"}
	glog.Infof("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		glog.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				glog.Errorf("read: %v", err)
				return
			}
			glog.Infof("recv: %s", message)
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
			glog.Errorf("write: %v", err)
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
			glog.Info("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				glog.Errorf("write close: %v", err)
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
