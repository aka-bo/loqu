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

	"github.com/golang/glog"
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
	glog.Infof("Run called with the following: %#v", o)

	if o.UseWebSocket {
		o.dial()
	} else {
		o.postContinuously()
	}
}

func (o *Options) postContinuously() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	o.post()
	if o.IntervalSeconds <= 0 {
		return
	}

	ticker := time.NewTicker(time.Duration(o.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.post()
		case <-interrupt:
			glog.Info("interupt")
			return
		}
	}
}

func (o *Options) post() {
	requestBody, err := json.Marshal(map[string]string{
		"loqu": "loqu",
	})

	if err != nil {
		glog.Fatal(err)
	}

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	url := fmt.Sprintf("http://%s:%d/post", o.Host, o.Port)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		glog.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		glog.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Fatal(err)
	}

	fmt.Println(string(body))
}
