package main

import (
	"encoding/json"

	"io/ioutil"

	"github.com/goarne/logging"
)

//AppConfig stores the applications configuration which is available runtime.
type AppConfig struct {
	Consul      ConsulConfig
	Server      ServerConfig
	Tracelogger logging.LogConfig
	ErrorLogger logging.LogConfig
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

//ServerConfig stores the HTTP server configuration
type ServerConfig struct {
	Port      int64
	Root      string
	Resources ServerAPI
}

//ServerAPI stores the REST API resources.
type ServerAPI struct {
	FindMediaFiles string
	SortMediaFiles string
	HealthCheck    string
}

//ReadConfig reads configuration from a configfile.
func (a *AppConfig) ReadConfig(configFile string) error {
	fileContent, e := ioutil.ReadFile(configFile)

	if e != nil {
		return e
	}

	json.Unmarshal(fileContent, &a)
	return nil
}
