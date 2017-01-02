package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/goarne/logging"
)

var (
	lastHealthCheck time.Time
	health          chan bool
	timeout         chan bool
)

func StartConsulClient(cs ConsulConfig) {
	timeout = make(chan bool, 1)
	health = make(chan bool)

	go registerService(cs)
	go checkTimeout(cs)
}

func RegisterCheckAlive() {
	health <- true
}

func checkTimeout(cs ConsulConfig) {
	for {
		time.Sleep(time.Second * time.Duration(cs.RegisterInterval))

		if time.Since(lastHealthCheck) > time.Second*time.Duration(cs.RegisterInterval) {
			logging.Trace.Println("Have not received a checkalive within", cs.RegisterInterval, "seconds.")
			timeout <- true
		}
	}
}

func registerService(cs ConsulConfig) {
	defer close(timeout)
	defer close(health)

	for {
		select {
		case <-health:
			logging.Trace.Println("Healthchec ok.")
			lastHealthCheck = time.Now()
			// a read from healthcheck has occurred
		case <-timeout:
			// the read from ch has timed out
			if err := registerToConsul(cs); err != nil {
				logging.Error.Println(err)
			} else {
				logging.Trace.Println("Registered to consul:", cs.Registerurl)
			}
		}
	}
}

//RegisterToConsul method registers service to a Consul instance.
func registerToConsul(cs ConsulConfig) error {
	msortwebService, _ := json.Marshal(cs.Payload.RegistryService)

	client := &http.Client{}
	req, err := http.NewRequest(cs.Method, cs.Registerurl, bytes.NewReader(msortwebService))

	if err != nil {
		return errors.New("Could not create new request." + err.Error())
	}

	resp, err := client.Do(req)

	if err != nil {
		return errors.New("Could not process the request." + err.Error())
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		responseBody, _ := ioutil.ReadAll(resp.Body)

		return errors.New("Received error from server:" + resp.Status + "\n" + string(responseBody))
	}

	return nil
}

//ConsulConfig stores metadata about the service to register in Consul
type ConsulConfig struct {
	Registerurl      string `json:"registerurl"`
	Method           string `json:"method"`
	RegisterInterval int    `json:"registerinterval"`
	Payload          struct {
		RegistryService Service `json:"Service"`
	} `json:"payload"`
}

//Service is the for registering a service to the Consule API
type Service struct {
	Name           string   `json:"Name"`
	Tags           []string `json:"Tags"`
	Address        string   `json:"Address"`
	Port           int      `json:"Port"`
	ServiceAddress string   `json:"ServiceAddress"`
	HTTPCheck      Check    `json:"Check"`
}

type Check struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	HTTP     string `json:"http"`
	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`
}
