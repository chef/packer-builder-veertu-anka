package testutils

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/veertuinc/packer-builder-veertu-anka/client"
	"github.com/veertuinc/packer-builder-veertu-anka/common"
)

type TestClient struct {
	Commands []string
	Errors   map[string]error
}

func (c *TestClient) Version() (client.VersionResponse, error) {
	var response client.VersionResponse

	out, err := exec.Command("anka", "--machine-readable", "version").Output()
	if err != nil {
		return response, err
	}

	err = json.Unmarshal([]byte(out), &response)
	return response, err
}

func (c *TestClient) Suspend(params client.SuspendParams) error {
	args := []string{"suspend", params.VMName}
	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return c.Errors[fullCmd]
}

func (c *TestClient) Start(params client.StartParams) error {
	args := []string{"start", params.VMName}
	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return c.Errors[fullCmd]
}

func (c *TestClient) Run(params client.RunParams) (error, int) {
	runner := client.NewRunner(params)
	if err := runner.Start(); err != nil {
		return err, 1
	}

	log.Printf("Waiting for command to run")
	return runner.Wait()
}

func (c *TestClient) Create(params client.CreateParams, outputStreamer chan string) (client.CreateResponse, error) {
	var response client.CreateResponse

	args := []string{
		"create",
		"--app", params.InstallerApp,
		"--ram-size", params.RAMSize,
		"--cpu-count", params.CPUCount,
		"--disk-size", params.DiskSize,
		params.Name,
	}

	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return response, c.Errors[fullCmd]
}

func (c *TestClient) Describe(vmName string) (client.DescribeResponse, error) {
	var response client.DescribeResponse
	// this is where we may need to define how runCommand and runAnkaCommand should be executed
	// or we could add back the client response and set checks on the output there

	args := []string{"describe", vmName}
	fullCmd := strings.Join(args, " ")
	fmt.Println(fullCmd)
	c.Commands = append(c.Commands, fullCmd)
	return response, c.Errors[fullCmd]
}

func (c *TestClient) Show(vmName string) (client.ShowResponse, error) {
	var response client.ShowResponse

	args := []string{"show", vmName}
	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return response, c.Errors[fullCmd]
}

func (c *TestClient) Copy(params client.CopyParams) error {
	args := []string{"cp", "-af", params.Src, params.Dst}
	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return c.Errors[fullCmd]
}

func (c *TestClient) Clone(params client.CloneParams) error {
	args := []string{"clone", params.SourceUUID, params.VMName}
	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return c.Errors[fullCmd]
}

func (c *TestClient) Stop(params client.StopParams) error {
	args := []string{"stop"}

	if params.Force {
		args = append(args, "--force")
	}

	args = append(args, params.VMName)

	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return c.Errors[fullCmd]
}

func (c *TestClient) Delete(params client.DeleteParams) error {
	args := []string{"modify", params.VMName}
	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return c.Errors[fullCmd]
}

func (c *TestClient) Exists(vmName string) (bool, error) {
	_, err := c.Show(vmName)
	if err == nil {
		return true, nil
	}
	switch err.(type) {
	case *json.UnmarshalTypeError:
	case *common.VMNotFoundException:
		return false, nil
	}
	return false, err
}

func (c *TestClient) Modify(vmName string, command string, property string, flags ...string) error {
	args := []string{"modify", vmName, command, property}
	args = append(args, flags...)
	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return c.Errors[fullCmd]
}
