package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
)

const (
	statusOK    = "OK"
	statusERROR = "ERROR"
)

type Client struct {
	RegistryURL string
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#registry-list-repos
type DefaultRegistryRepoResponse struct {
	ID     string `json:"id"`
	Host   string `json:"host"`
	Scheme string `json:"scheme"`
	Port   string `json:"port"`
}

func NewClient() *Client {
	registryURL := ""

	output, err := runAnkaCommand("registry", "list-repos", "--default")
	if err != nil {
		log.Printf("Could not determine Registry URL for default registry")
	}

	var response DefaultRegistryRepoResponse
	err = json.Unmarshal(output.Body, &response)
	if err == nil {
		registryURL = fmt.Sprintf("%s://%s:%s", response.Scheme, response.Host, response.Port)
	} else {
		log.Printf("Could not process output to determine Registry URL for default registry")
	}

	return &Client{
		RegistryURL: registryURL,
	}
}

type machineReadableError struct {
	*machineReadableOutput
}

func (ae machineReadableError) Error() string {
	return ae.Message
}

type machineReadableOutput struct {
	Status        string `json:"status"`
	Body          json.RawMessage
	Message       string `json:"message"`
	Code          int    `json:"code"`
	ExceptionType string `json:"exception_type"`
}

func (parsed *machineReadableOutput) GetError() error {
	if parsed.Status != statusOK {
		return machineReadableError{parsed}
	}
	return nil
}

func parseOutput(output []byte) (machineReadableOutput, error) {
	var parsed machineReadableOutput
	if err := json.Unmarshal(output, &parsed); err != nil {
		return parsed, err
	}

	return parsed, nil
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func customSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// A tiny spin off on ScanLines

	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, dropCR(data[0:i]), nil
	}
	if atEOF { // Machine readable data is parsed here
		out := dropCR(data)
		return len(data), out, customErr{data: out}
	}
	return 0, nil, nil
}

type customErr struct {
	data []byte
}

func (e customErr) Error() string {
	return string(e.data)
}
