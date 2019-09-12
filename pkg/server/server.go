package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
)

// Options is used to configure the server
type Options struct {
	ShutdownDelaySeconds int
	ListenPort           int
}

// Run the server, get them logs flowing
func Run(o *Options) {
	glog.Infof("Run called with the following: %#v", o)

	stop := make(chan bool, 1)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		glog.V(3).Infof("Handling %s", req.URL.Path)

		values := &struct {
			Path   string
			Query  string
			Method string
		}{
			Path:   req.URL.Path,
			Query:  req.URL.RawQuery,
			Method: req.Method,
		}
		b, err := json.Marshal(values)
		if err != nil {
			glog.Errorf("unable to marshal response with values=%v", values)
			write(w, http.StatusInternalServerError, errorResponse("unable to marshal response"))
		}
		write(w, http.StatusOK, b)
	})
	mux.Handle("/healthcheck", healthCheckHandler(stop))

	addr := fmt.Sprintf(":%d", o.ListenPort)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		glog.Infof("Listing on http://0.0.0.0%s", addr)

		if err := server.ListenAndServe(); err != nil {
			glog.Fatal(err)
		}
	}()

	sig := <-shutdown

	glog.Infof("%v signal received. signaling healthcheck to fail", sig)
	stop <- true

	glog.Infof("shutting down in %v seconds", o.ShutdownDelaySeconds)
	<-time.After(time.Duration(o.ShutdownDelaySeconds) * time.Second)
	glog.Infof("proceeding with shutdown")

	glog.Infof("commencing graceful shutdown of web server")
	server.Shutdown(context.Background())

	glog.Flush()
}

func write(w http.ResponseWriter, status int, v []byte) {
	w.WriteHeader(status)
	if _, err := w.Write(v); err != nil {
		glog.Error("error writing data to the response writer", err.Error())
	}
}

func errorResponse(msg string) []byte {
	v := &struct {
		Message   string
		ErrorCode int
	}{
		Message:   msg,
		ErrorCode: http.StatusInternalServerError,
	}

	b, _ := json.Marshal(v)
	return b
}
