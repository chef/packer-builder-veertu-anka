package client

import (
	"bufio"
	"errors"
	"log"
	"os/exec"
	"strings"

	"github.com/veertuinc/packer-builder-veertu-anka/common"
)

const (
	AnkaNameAlreadyExistsErrorCode   = 18
	AnkaVMNotFoundExceptionErrorCode = 3
)

// A helper that can return a bool for whether or not a VM exists locally
type ExistsParams struct {
	Name string
	Tag  string
}

func (c *Client) Exists(params ExistsParams) (bool, error) {
	resp, err := c.Show(params.Name)
	if err == nil {
		if params.Tag != "" && resp.Tag != params.Tag {
			return false, nil
		}
		return true, nil
	}
	switch err.(type) {
	// case *json.UnmarshalTypeError:
	case *common.VMNotFoundException:
		return false, nil
	}
	return false, err
}

func runAnkaCommand(args ...string) (machineReadableOutput, error) {
	return runAnkaCommandStreamer(nil, args...)
}

func runAnkaCommandStreamer(outputStreamer chan string, args ...string) (machineReadableOutput, error) {
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
	_ = cmd.Wait()

	if err = parsed.GetError(); err != nil {
		return machineReadableOutput{}, err
	}

	return parsed, nil
}
