package client

import (
	"encoding/json"
	"fmt"
	"log"
)

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#registry-list
type RegistryListResponse struct {
	Latest string `json:"latest"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

func (c *Client) RegistryList(registryParams RegistryParams) ([]RegistryListResponse, error) {
	output, err := runAnkaRegistryCommand(registryParams, "list")
	if err != nil {
		return nil, err
	}
	if output.Status != "OK" {
		log.Print("Error executing registry list command: ", output.ExceptionType, " ", output.Message)
		return nil, fmt.Errorf(output.Message)
	}

	var response []RegistryListResponse
	err = json.Unmarshal(output.Body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#registry-push
type RegistryPushParams struct {
	VMID        string
	Tag         string
	Description string
	RemoteVM    string
	Local       bool
}

func (c *Client) RegistryPush(registryParams RegistryParams, pushParams RegistryPushParams) error {
	cmdArgs := []string{"push"}
	if pushParams.Tag != "" {
		cmdArgs = append(cmdArgs, "--tag", pushParams.Tag)
	}
	if pushParams.Description != "" {
		cmdArgs = append(cmdArgs, "--description", pushParams.Description)
	}
	if pushParams.RemoteVM != "" {
		cmdArgs = append(cmdArgs, "--remote-vm", pushParams.RemoteVM)
	}
	if pushParams.Local {
		cmdArgs = append(cmdArgs, "--local")
	}
	cmdArgs = append(cmdArgs, pushParams.VMID)

	output, err := runAnkaRegistryCommand(registryParams, cmdArgs...)
	if err != nil {
		return err
	}
	if output.Status != "OK" {
		log.Print("Error executing registry push command: ", output.ExceptionType, " ", output.Message)
		return fmt.Errorf(output.Message)
	}
	return nil
}

// Run command against the registry
type RegistryParams struct {
	RegistryName string
	RegistryURL  string
	NodeCertPath string
	NodeKeyPath  string
	CaRootPath   string
	IsInsecure   bool
}

func runAnkaRegistryCommand(registryParams RegistryParams, args ...string) (machineReadableOutput, error) {
	cmdArgs := []string{"registry"}
	if registryParams.RegistryName != "" {
		cmdArgs = append(cmdArgs, "--remote", registryParams.RegistryName)
	}
	if registryParams.RegistryURL != "" {
		cmdArgs = append(cmdArgs, "--registry-path", registryParams.RegistryURL)
	}
	if registryParams.NodeCertPath != "" {
		cmdArgs = append(cmdArgs, "--cert", registryParams.NodeCertPath)
	}
	if registryParams.NodeKeyPath != "" {
		cmdArgs = append(cmdArgs, "--key", registryParams.NodeKeyPath)
	}
	if registryParams.CaRootPath != "" {
		cmdArgs = append(cmdArgs, "--cacert", registryParams.CaRootPath)
	}
	if registryParams.IsInsecure {
		cmdArgs = append(cmdArgs, "--insecure")
	}

	cmdArgs = append(cmdArgs, args...)

	return runAnkaCommand(cmdArgs...)
}
