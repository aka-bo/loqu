package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
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

	sig := <-stop
	glog.Infof("%v signal received. shutting down...", sig)
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
