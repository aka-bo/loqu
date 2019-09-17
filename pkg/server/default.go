package server

import (
	"net/http"

	"github.com/aka-bo/loqu/pkg/util"
)

//Default is the default request handler
type Default struct {
	serverInfo serverInfo

	stopChan chan bool
}

//Handle the request and write a response
func (d *Default) Handle(w http.ResponseWriter, r *http.Request) {
	logger := util.WithID("Handle", r).WithValues("path", r.URL.Path)
	logger.Info("Handling request")

	select {
	case <-d.stopChan:
		logger.Info("Shutdown signal received. processing will continue normally.")
		d.serverInfo.Stopping = true
	default:
	}

	values := buildResponse(&d.serverInfo, r)
	b, err := marshal(values, true)
	if err != nil {
		logger.Error(err, "Unable to marshal response", "values", values)
		write(w, http.StatusInternalServerError, errorResponse("unable to marshal response"), logger)
	}
	if logger.V(4).Enabled() {
		b2, _ := marshal(values, false)
		logger.Info("Writing response", "body", string(b2))
	}
	write(w, http.StatusOK, b, logger)
}

//Start the HealthCheck
func (d *Default) Start() {
	d.stopChan = make(chan bool, 1)
}

//Stop signals that the shutdown process has begun
func (d *Default) Stop() {
	d.serverInfo.Stopping = true
	d.stopChan <- true
}
