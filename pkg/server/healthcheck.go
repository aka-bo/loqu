package server

import (
	"encoding/json"
	"net/http"

	"github.com/aka-bo/loqu/pkg/util"
)

//HealthCheck provides a handler for health check requests
type HealthCheck struct {
	serverInfo serverInfo

	stopChan chan bool
}

//Handle healcheck requests
func (h *HealthCheck) Handle(w http.ResponseWriter, r *http.Request) {
	logger := util.WithID("Handle", r).WithValues("path", r.URL.Path)
	logger.Info("Handling request")

	select {
	case <-h.stopChan:
		logger.Info("Shutdown signal received. Handler will now begin returning error codes")
		h.serverInfo.Stopping = true
	default:
	}

	b, err := json.Marshal(struct {
		Healthy bool
		Info    *response
	}{
		Healthy: true,
		Info:    buildResponse(&h.serverInfo, r),
	})

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		write(w, http.StatusInternalServerError, errorResponse("Unable to marshal json response"), logger)
		return
	}

	status := http.StatusOK

	if h.serverInfo.Stopping {
		status = http.StatusInternalServerError
		logger.Info("Returning unhealthy because of shutdown signal")
	}

	write(w, status, b, logger)
}

//Start the HealthCheck
func (h *HealthCheck) Start() {
	h.stopChan = make(chan bool, 1)
}

//Stop signals that the shutdown process has begun
func (h *HealthCheck) Stop() {
	h.serverInfo.Stopping = true
	h.stopChan <- true
}
