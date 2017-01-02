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
