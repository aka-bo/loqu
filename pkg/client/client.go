package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-logr/glogr"
	"github.com/go-logr/logr"
	"github.com/aka-bo/loqu/pkg/util"
)

// Options is used to configure the client
type Options struct {
	Host string
	Port int

	UseWebSocket    bool
	IntervalSeconds int
	Data            *string
	ExitMode        bool
}

func (o *Options) dataOrDefault(data fmt.Stringer) []byte {
	if o.Data != nil {
		return []byte(*o.Data)
	}

	return []byte(data.String())
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
	url := fmt.Sprintf("http://%s:%d/post", o.Host, o.Port)
	id := util.NewRequestID()
	logger = logger.WithValues("requestID", id, "url", url)
	logger.Info("post")

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(o.dataOrDefault(time.Now())))
	if err != nil {
		o.handleError(logger, err, "failed to create new request")
		return
	}
	req.Header.Set(util.KeyRequestID, id)

	resp, err := client.Do(req)
	if err != nil {
		o.handleError(logger, err, "error sending http request")
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		o.handleError(logger, err, "error reading response")
		return
	}
	logger.Info("response received", "code", resp.StatusCode)
	fmt.Println(string(body))
}

func (o *Options) handleError(logger logr.Logger, err error, msg string) {
	logger.Error(err, msg)
	if o.ExitMode {
		os.Exit(1)
	}
}
