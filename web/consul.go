package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

//RegisterToConsul method registers service to a Consul instance.
func RegisterToConsul(pl Service) error {
	msortwebService, _ := json.Marshal(pl)

	client := &http.Client{}
	req, err := http.NewRequest(appConfig.Consul.Method, appConfig.Consul.Registerurl, bytes.NewReader(msortwebService))

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
