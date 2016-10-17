//The package handles a webserver which serves MSORT API

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	//"time"

	"github.com/goarne/logging"
	"github.com/goarne/msort/core"
	"github.com/goarne/web"
)

var (
	appConfig AppConfig
)

type filesFound struct {
	NumberOfFilesFound int
	Files              []*core.ArchiveFile
}

func init() {
	appConfig = lastAppKonfig()
	StartConsulClient(appConfig.Consul)

	rotatingTraceWriter := logging.CreateRotatingWriter(appConfig.Tracelogger)
	rotatingErrorWriter := logging.CreateRotatingWriter(appConfig.ErrorLogger)

	tracerLogger := logging.CreateLogWriter(rotatingTraceWriter)
	tracerLogger.Append(os.Stdout)

	errorLogger := logging.CreateLogWriter(rotatingErrorWriter)
	errorLogger.Append(os.Stdout)

	logging.InitLoggers(tracerLogger, tracerLogger, errorLogger, errorLogger)
}

func main() {

	router := createWebRouter()

	if err := http.ListenAndServe(":"+strconv.FormatInt(appConfig.Server.Port, 10), router); err != nil {
		logging.Error.Println(err)
	}

	logging.Trace.Println("Server stopped.")
}

func createWebRouter() *web.WebRouter {
	rootPath := web.NewRoute().Path(appConfig.Server.Root).Method(web.HttpGet).HandlerFunc(httpGetSample)
	healthCheckPath := web.NewRoute().Path(appConfig.Server.Root + appConfig.Server.Resources.HealthCheck).Method(web.HttpGet).HandlerFunc(httpGetHealthCheck)
	sortFileRoute := web.NewRoute().Path(appConfig.Server.Root + appConfig.Server.Resources.SortMediaFiles).Method(web.HttpPost).HandlerFunc(httpPostSortFiles)
	findFileRoute := web.NewRoute().Path(appConfig.Server.Root + appConfig.Server.Resources.FindMediaFiles).Method(web.HttpPost).HandlerFunc(httpPostFindFiles)

	router := web.NewWebRouter()
	router.AddRoute(rootPath)
	router.AddRoute(healthCheckPath)
	router.AddRoute(sortFileRoute)
	router.AddRoute(findFileRoute)

	return router
}

func httpGetHealthCheck(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("Ok!"))
	RegisterCheckAlive()
}

func httpGetSample(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("Sample JSON payload\n"))
	var msortPrms core.CmdLineParams
	encoder := json.NewEncoder(resp)
	encoder.Encode(&msortPrms)
}

func httpPostFindFiles(resp http.ResponseWriter, req *http.Request) {
	logging.Trace.Println("Received request:", req)

	err := parseCommand(req)

	if err != nil {
		logging.Trace.Println(err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	core.StartAsync(1)

	fileList := make(chan *core.ArchiveFile)

	go core.FindFiles(fileList)

	foundFiles := filesFound{0, make([]*core.ArchiveFile, 0)}

	for file := range fileList {
		foundFiles.Files = append(foundFiles.Files, file)
	}

	defer core.WaitAsync()
	foundFiles.NumberOfFilesFound = core.FileCount

	encoder := json.NewEncoder(resp)
	encoder.Encode(&foundFiles)
}

func httpPostSortFiles(resp http.ResponseWriter, req *http.Request) {
	logging.Trace.Println("Received request:", req)

	err := parseCommand(req)

	if err != nil {
		logging.Trace.Println(err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	core.StartAsync(2)

	filesToArchive := make(chan *core.ArchiveFile)

	go core.FindFiles(filesToArchive)

	go core.ArchiveFiles(filesToArchive)

	core.WaitAsync()
	kvittering := "Found " + strconv.Itoa(core.FileCount) + " file(s)."
	logging.Trace.Println(kvittering)
	resp.Write([]byte(kvittering))
}

//Laster inn applikasjonskonfigurasjon fra en JSON konfigurasjonsfil.
func lastAppKonfig() AppConfig {

	configFile := flag.String("config-file", "", "Relative path to application configfile (json)")
	flag.Parse()

	if strings.Compare("", *configFile) == 0 {
		fmt.Println("Missing config-file.")
		os.Exit(1)
	}

	appConfig := AppConfig{}

	if err := appConfig.ReadConfig(*configFile); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Read config:", appConfig)
	return appConfig
}

func parseCommand(req *http.Request) error {
	cmdPrm := core.CmdLineParams{}

	if req.Body == nil {
		return errors.New("Missing body in request.")
	}

	err := json.NewDecoder(req.Body).Decode(&cmdPrm)

	if err != nil {
		return err
	}

	core.CmdPrm.FilePattern = cmdPrm.FilePattern
	core.CmdPrm.Overwrite = cmdPrm.Overwrite
	core.CmdPrm.ShallArchive = cmdPrm.ShallArchive
	core.CmdPrm.Source = cmdPrm.Source
	core.CmdPrm.Target = cmdPrm.Target
	core.CmdPrm.TargetPattern = cmdPrm.TargetPattern
	core.CmdPrm.Verbose = cmdPrm.Verbose

	return nil

}
