package server

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
	"github.com/google/uuid"
)

//HealthCheck provides a handler for health check requests
type HealthCheck struct {
	serverInfo serverInfo

	stopChan chan bool
}

//Handle healcheck requests
func (h *HealthCheck) Handle(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	glog.V(3).Infof("[%s] Handling %s", id, r.URL.Path)

	select {
	case <-h.stopChan:
		glog.Infof("[%s] Shutdown signal received. %s will now begin returning error codes.", id, r.URL.Path)
		h.serverInfo.Stopping = true
	default:
	}

	b, err := json.Marshal(struct {
		Healthy bool
		Info    *response
	}{
		Healthy: true,
		Info:    buildResponse(id, &h.serverInfo, r),
	})

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		write(w, http.StatusInternalServerError, errorResponse(id, "Unable to marshal json response"))
		return
	}

	status := http.StatusOK

	if h.serverInfo.Stopping {
		status = http.StatusInternalServerError
		glog.Info("returning unhealthy because of shutdown signal")
	}

	write(w, status, b)
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
