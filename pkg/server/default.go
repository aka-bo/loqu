package server

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/google/uuid"
)

//Default is the default request handler
type Default struct {
	serverInfo serverInfo

	stopChan chan bool
}

//Handle the request and write a response
func (d *Default) Handle(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()
	glog.V(3).Infof("[%s] Handling %s", id, r.URL)

	select {
	case <-d.stopChan:
		glog.Infof("[%s] Shutdown signal received. %s will continue processing normally.", id, r.URL.Path)
		d.serverInfo.Stopping = true
	default:
	}

	values := buildResponse(id, &d.serverInfo, r)
	b, err := marshal(values, true)
	if err != nil {
		glog.Errorf("[%s] unable to marshal response with values=%v", id, values)
		write(w, http.StatusInternalServerError, errorResponse(id, "unable to marshal response"))
	}
	if glog.V(4) {
		b2, _ := marshal(values, false)
		glog.Infof("[%s] Response: %s", id, b2)
	}
	write(w, http.StatusOK, b)
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
