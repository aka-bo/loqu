package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/glogr"
	"github.com/go-logr/logr"
	"github.com/golang/glog"

	"github.com/aka-bo/loqu/pkg/util"
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
	Body    string      `json:"body,omitempty"`
	Headers http.Header `json:"headers"`
}

type response struct {
	ID      string      `json:"id"`
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
		mux.Handle(k, requestIDHandler(v.Handle))
	}
}

func requestIDHandler(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(util.RequestContext(r)))
	})
}

func (h handlerMap) shutdown() {
	for _, v := range h {
		v.Stop()
	}
}

// Run the server, get them logs flowing
func Run(o *Options) {
	logger := glogr.New().WithName("Server")
	logger.Info("Run called", "options", o)

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
		"/echo":        &Echo{serverInfo: serverInfo, shutdownGracePeriodSeconds: o.ShutdownDelaySeconds - 1},
		"/healthcheck": &HealthCheck{serverInfo: serverInfo},
	}

	mux := http.NewServeMux()
	handlers.register(mux)
	// mux.Handle("/demo", demoHandler())

	addr := fmt.Sprintf(":%d", o.ListenPort)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	server.RegisterOnShutdown(func() {
		logger.Info("Shutdown() called on http.Server")
	})

	go func() {
		logger.Info("Starting server", "addr", addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "server exited with error")
		}
	}()

	sig := <-shutdown

	logger.Info("signal received. signaling handlers", "signal", sig.String())
	handlers.shutdown()

	logger.Info("shutting down with delay", "delay", o.ShutdownDelaySeconds)
	<-time.After(time.Duration(o.ShutdownDelaySeconds) * time.Second)
	logger.Info("proceeding with shutdown")

	logger.Info("commencing graceful shutdown of web server")
	server.Shutdown(context.Background())

	glog.Flush()
}

func buildResponse(server *serverInfo, r *http.Request) *response {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	return &response{
		ID: util.GetRequestID(r),
		Client: clientInfo{
			Address: r.RemoteAddr,
		},
		Server: *server,
		Request: requestInfo{
			Body:    string(body),
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

func write(w http.ResponseWriter, status int, v []byte, logger logr.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(v); err != nil {
		logger.Error(err, "error writing data to the response writer")
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
