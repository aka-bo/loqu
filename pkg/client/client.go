package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-logr/glogr"
	"github.com/go-logr/logr"
)

// Options is used to configure the client
type Options struct {
	Host string
	Port int

	UseWebSocket    bool
	IntervalSeconds int
}

// Run the client
func Run(o *Options) {
	logger := glogr.New().WithName("Client")
	logger.Info("Run called", "options", o)

	if o.UseWebSocket {
		o.dial(logger)
	} else {
		o.postContinuously(logger)
	}
}

func (o *Options) postContinuously(logger logr.Logger) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	o.post(logger)
	if o.IntervalSeconds <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(o.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.post(logger)
		case <-interrupt:
			logger.Info("interupt")
			return
		}
	}
}

func (o *Options) post(logger logr.Logger) {
	requestBody, err := json.Marshal(map[string]string{
		"loqu": "loqu",
	})

	if err != nil {
		logger.Error(err, "failed to marshal request body")
		return
	}

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	url := fmt.Sprintf("http://%s:%d/post", o.Host, o.Port)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Error(err, "failed to create new request", "url", url)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err, "error sending http request")
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err, "error reading response")
		return
	}

	fmt.Println(string(body))
}
