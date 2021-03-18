package anka

import (
	"context"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/veertuinc/packer-builder-veertu-anka/client"
)

type StepSetGeneratedData struct {
	client        client.Client
	vmName        string
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepSetGeneratedData) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	log.Printf("Exposing build contextual variables...")

	osVersion := state.Get("os_version")
	s.client = state.Get("client").(client.Client)
	s.vmName = state.Get("vm_name").(string)

	darwinVersion, err := s.client.RunWithOutput(client.RunParams{
		Command: []string{"run", s.vmName, "uname", "-r"},
	})
	if err != nil {
		return multistep.ActionHalt
	}

	if osVersion == nil {
		osv, err := s.client.RunWithOutput(client.RunParams{
			Command: []string{"run", s.vmName, "sw_vers", "-productVersion"},
		})
		if err != nil {
			return multistep.ActionHalt
		}

		osVersion = string(osv)
	}

	s.GeneratedData.Put("VMName", s.vmName)
	s.GeneratedData.Put("OSVersion", osVersion)
	s.GeneratedData.Put("DarwinVersion", string(darwinVersion))

	return multistep.ActionContinue
}

// Cleanup will run whenever there are any errors.
// No cleanup needs to happen here
func (s *StepSetGeneratedData) Cleanup(_ multistep.StateBag) {
}
