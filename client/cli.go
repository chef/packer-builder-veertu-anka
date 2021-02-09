package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	"strings"
)

const (
	statusOK                         = "OK"
	statusERROR                      = "ERROR" //nolint:deadcode,varcheck
	AnkaNameAlreadyExistsErrorCode   = 18
	AnkaVMNotFoundExceptionErrorCode = 3
)

type CLI interface {
	runCommand(args ...string) (machineReadableOutput, error)
	runCommandStreamer(outputStreamer chan string, args ...string) (machineReadableOutput, error)
}

type AnkaCLI struct {
}

func (cli *AnkaCLI) runCommand(args ...string) (machineReadableOutput, error) {
	return cli.runCommandStreamer(nil, args...)
}

func (cli *AnkaCLI) runCommandStreamer(outputStreamer chan string, args ...string) (machineReadableOutput, error) {

	if outputStreamer != nil {
		args = append([]string{"--debug"}, args...)
	}

	cmdArgs := append([]string{"--machine-readable"}, args...)
	log.Printf("Executing anka %s", strings.Join(cmdArgs, " "))
	cmd := exec.Command("anka", cmdArgs...)

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Err on stdoutpipe")
		return machineReadableOutput{}, err
	}

	if outputStreamer == nil {
		cmd.Stderr = cmd.Stdout
	}

	if err = cmd.Start(); err != nil {
		log.Printf("Failed with an error of %v", err)
		return machineReadableOutput{}, err
	}
	outScanner := bufio.NewScanner(outPipe)
	outScanner.Split(customSplit)

	for outScanner.Scan() {
		out := outScanner.Text()
		log.Printf("%s", out)

		if outputStreamer != nil {
			outputStreamer <- out
		}
	}

	scannerErr := outScanner.Err() // Expecting error on final output
	if scannerErr == nil {
		return machineReadableOutput{}, errors.New("missing machine readable output")
	}
	if _, ok := scannerErr.(customErr); !ok {
		return machineReadableOutput{}, err
	}

	finalOutput := scannerErr.Error()
	log.Printf("%s", finalOutput)

	parsed, err := parseOutput([]byte(finalOutput))
	if err != nil {
		return machineReadableOutput{}, err
	}
	if err := cmd.Wait(); err != nil {
		return machineReadableOutput{}, err
	}

	if err = parsed.GetError(); err != nil {
		return machineReadableOutput{}, err
	}

	return parsed, nil
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
