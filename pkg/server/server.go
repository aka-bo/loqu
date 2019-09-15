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
	"github.com/google/uuid"
)

// Options is used to configure the server
type Options struct {
	ShutdownDelaySeconds int
	ListenPort           int
}

type clientInfo struct {
	Address string `json:"address"`
}

type serverInfo struct {
	Hostname string    `json:"hostname"`
	Started  time.Time `json:""`
	Stopping bool      `json:"stopping"`
}

type requestInfo struct {
	Path    string      `json:"path"`
	Query   string      `json:"query,omitempty"`
	Method  string      `json:"method"`
	Headers http.Header `json:"headers"`
}

type response struct {
	ID      uuid.UUID   `json:"id"`
	Client  clientInfo  `json:"client"`
	Server  serverInfo  `json:"server"`
	Request requestInfo `json:"request"`
}

//Handler provides lifecycle hooks for an HttpHandler
type Handler interface {
	Handle(http.ResponseWriter, *http.Request)
	Start()
	Stop()
}

type handlerMap map[string]Handler

func (h handlerMap) register(mux *http.ServeMux) {
	for k, v := range h {
		v.Start()
		mux.HandleFunc(k, http.HandlerFunc(v.Handle))
	}
}

func (h handlerMap) shutdown() {
	for _, v := range h {
		v.Stop()
	}
}

// Run the server, get them logs flowing
func Run(o *Options) {
	glog.Infof("Run called with the following: %#v", o)

	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	serverInfo := serverInfo{
		Hostname: host,
		Started:  time.Now(),
	}

	handlers := handlerMap{
		"/":            &Default{serverInfo: serverInfo},
		"/echo":        &Echo{serverInfo: serverInfo},
		"/healthcheck": &HealthCheck{serverInfo: serverInfo},
	}

	mux := http.NewServeMux()
	handlers.register(mux)
	mux.Handle("/demo", demoHandler())

	addr := fmt.Sprintf(":%d", o.ListenPort)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	server.RegisterOnShutdown(func() {
		glog.Info("Shutdown() called on http.Server")
	})

	go func() {
		glog.Infof("Listing on http://0.0.0.0%s", addr)

		if err := server.ListenAndServe(); err != nil {
			glog.Fatal(err)
		}
	}()

	sig := <-shutdown

	glog.Infof("%v signal received. signaling handlers", sig)
	handlers.shutdown()

	glog.Infof("shutting down in %v seconds", o.ShutdownDelaySeconds)
	<-time.After(time.Duration(o.ShutdownDelaySeconds) * time.Second)
	glog.Infof("proceeding with shutdown")

	glog.Infof("commencing graceful shutdown of web server")
	server.Shutdown(context.Background())

	glog.Flush()
}

func buildResponse(id uuid.UUID, server *serverInfo, r *http.Request) *response {
	return &response{
		ID: id,
		Client: clientInfo{
			Address: r.RemoteAddr,
		},
		Server: *server,
		Request: requestInfo{
			Path:    r.URL.Path,
			Query:   r.URL.RawQuery,
			Method:  r.Method,
			Headers: r.Header,
		},
	}
}

func marshal(v interface{}, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(v, "", "    ")
	}
	return json.Marshal(v)
}

func write(w http.ResponseWriter, status int, v []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(v); err != nil {
		glog.Error("error writing data to the response writer", err.Error())
	}
}

func errorResponse(requestID uuid.UUID, msg string) []byte {
	v := &struct {
		RequestID string
		Message   string
		ErrorCode int
	}{
		RequestID: requestID.String(),
		Message:   msg,
		ErrorCode: http.StatusInternalServerError,
	}

	b, _ := json.Marshal(v)
	return b
}
