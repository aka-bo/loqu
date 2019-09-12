package server

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
)

// HealthCheckHandler runs the healthcheck and returns ok if healthy
func healthCheckHandler(stop chan bool) http.Handler {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	var shutdown bool
	started := time.Now()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.V(3).Infof("Handling %s", r.URL.Path)

		select {
		case <-stop:
			glog.Infof("Shutdown signal received. %s will now begin returning error codes.", r.URL.Path)
			shutdown = true
		default:
		}

		b, err := json.Marshal(struct {
			Healthy  bool
			Shutdown bool
			Started  time.Time
			Hostname string
		}{
			Healthy:  true,
			Shutdown: shutdown,
			Started:  started,
			Hostname: host,
		})

		w.Header().Set("Content-Type", "application/json")

		if err != nil {
			write(w, http.StatusInternalServerError, errorResponse("Unable to marshal json response"))
			return
		}

		status := http.StatusOK

		if shutdown {
			status = http.StatusInternalServerError
			glog.Info("returning unhealthy because of shutdown signal")
		}

		write(w, status, b)
	})
}
