package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/packer-plugin-sdk/net"
)

func registryRESTRequest(method string, url string, body io.Reader) (machineReadableOutput, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return machineReadableOutput{}, err
	}

	log.Printf("[API REQUEST] [%s] %s", method, url)

	httpClient := net.HttpClientWithEnvironmentProxy()
	resp, err := httpClient.Do(request)
	if err != nil {
		return machineReadableOutput{}, err
	}

	if resp.StatusCode == 200 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return machineReadableOutput{}, err
		}

		log.Printf("[API RESPONSE] %s", string(body))

		return parseOutput(body)
	}

	return machineReadableOutput{}, fmt.Errorf("unsupported http response code: %d", resp.StatusCode)
}
