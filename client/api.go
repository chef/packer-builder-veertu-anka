package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/packer/packer-plugin-sdk/net"
)

func registryRESTRequest(method string, url string, body io.Reader) (machineReadableOutput, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return machineReadableOutput{}, err
	}

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

		return parseOutput(body)
	}

	return machineReadableOutput{}, fmt.Errorf("unsupported http response code: %d", resp.StatusCode)
}
