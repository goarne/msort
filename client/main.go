package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/goarne/config"
	"github.com/goarne/logging"
	"github.com/goarne/msort/core"
)

var (
	appConfig       AppConfig
	msortWebService Service
)

func init() {
	appConfig = AppConfig{}

	cl := config.ConfigLoader{}

	//Adding the concrete unmarshalling of AppConfig structure to the generic Config function.
	cl.Unmarshall = func([]byte, interface{}) error {
		return yaml.Unmarshal(cl.FileContent, &appConfig)
	}

	cl.LoadAppKonfig()

	rotatingTraceWriter := logging.CreateRotatingWriter(appConfig.Tracelogger)
	rotatingErrorWriter := logging.CreateRotatingWriter(appConfig.ErrorLogger)

	tracerLogger := logging.CreateLogWriter(rotatingTraceWriter)
	tracerLogger.Append(os.Stdout)

	errorLogger := logging.CreateLogWriter(rotatingErrorWriter)
	errorLogger.Append(os.Stdout)

	logging.InitLoggers(tracerLogger, tracerLogger, errorLogger, errorLogger)
}

func main() {
	fmt.Println("Msort client.", appConfig)

	if err := LookupService(); err != nil {
		logging.Error.Println(err)
	}
	core.CmdPrm.ConfigFile = appConfig.MsortCommandConfig
	core.CmdPrm.ReadConfig()

	if err := SendArchiveRequest(); err != nil {
		logging.Error.Println(err)
	}
}

func LookupService() error {

	resp, err := http.Get(appConfig.Consul.ServiceUrl)

	if err != nil {
		return errors.New("Could not create new request." + err.Error())
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New("Received error from server:" + resp.Status + "\n" + string(responseBody))
	}

	json.Unmarshal(responseBody, &msortWebService)
	logging.Trace.Println("Found service: ", &msortWebService)
	return nil
}

//RegisterToConsul method registers service to a Consul instance.
func SendArchiveRequest() error {
	cmdPrmBytes, _ := json.Marshal(core.CmdPrm)
	logging.Trace.Println("Sending req:", string(cmdPrmBytes))
	client := &http.Client{}
	req, err := http.NewRequest("POST", msortWebService[0].ServiceAddress+"/find", bytes.NewReader(cmdPrmBytes))

	if err != nil {
		return errors.New("Could not create new request." + err.Error())
	}

	resp, err := client.Do(req)

	if err != nil {
		return errors.New("Could not process the request." + err.Error())
	}

	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)

	logging.Trace.Println("Found: ", string(responseBody))

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New("Received error from server:" + resp.Status + "\n" + string(responseBody))
	}

	return nil
}

type AppConfig struct {
	Consul             ConsulConfig
	MsortCommandConfig string
	Tracelogger        logging.LogConfig
	ErrorLogger        logging.LogConfig
}

//ConsulConfig stores metadata about the service to register in Consul
type ConsulConfig struct {
	ServiceUrl  string `yaml:"serviceurl"`
	ServiceName string `yaml:"servicename"`
}

//Service is the for registering a service to the Consule API

type Service []struct {
	Node                     string   `json:"Node"`
	Address                  string   `json:"Address"`
	ServiceID                string   `json:"ServiceID"`
	ServiceName              string   `json:"ServiceName"`
	ServiceTags              []string `json:"ServiceTags"`
	ServiceAddress           string   `json:"ServiceAddress"`
	ServicePort              int      `json:"ServicePort"`
	ServiceEnableTagOverride bool     `json:"ServiceEnableTagOverride"`
	CreateIndex              int      `json:"CreateIndex"`
	ModifyIndex              int      `json:"ModifyIndex"`
}
