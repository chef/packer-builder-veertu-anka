package client

import (
	"fmt"
	"log"
)

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#registry-list-repos
type RegistryRepo struct {
	ID     string
	Host   string
	Scheme string
	Port   string
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#registry-push
type RegistryPushParams struct {
	VMName      string
	Tag         string
	Description string
	RemoteVM    string
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
	cmdArgs = append(cmdArgs, pushParams.VMName)

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
