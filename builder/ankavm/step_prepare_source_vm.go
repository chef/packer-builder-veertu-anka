package ankavm

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/veertuinc/packer-builder-veertu-anka/client"
	"github.com/veertuinc/packer-builder-veertu-anka/common"
)

const (
	DEFAULT_DISK_SIZE = "40G"
	DEFAULT_RAM_SIZE  = "4G"
	DEFAULT_CPU_COUNT = "2"
)

type StepPrepareSourceVM struct {
	client   *client.Client
	vmName   string
	vmExists bool
	createVM bool
}

func (s *StepPrepareSourceVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var err error

	s.client = state.Get("client").(*client.Client)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	onError := func(err error) multistep.StepAction {
		return stepError(ui, state, err)
	}

	s.vmName = config.SourceVMName
	s.createVM = false

	installerAppFullName := "anka-packer-base"
	if config.InstallerApp != "" { // If users specifies an InstallerApp and sourceVMName doesn't exist, assume they want to build a new VM template and use the macOS installer version
		s.createVM = true
		ui.Say(fmt.Sprintf("Extracting version from installer app: %q", config.InstallerApp))
		macOSVersionFromInstallerApp, err := obtainMacOSVersionFromInstallerApp(config.InstallerApp) // Grab the version details from the Info.plist inside of the Installer package
		if err != nil {
			return onError(err)
		}
		installerAppFullName = fmt.Sprintf("%s-%s", installerAppFullName, macOSVersionFromInstallerApp) // We need to set the SourceVMName since the user didn't and the logic below creates a VM using it
	}

	if s.vmName == "" {
		s.vmName = installerAppFullName
	}

	// Reuse the base VM template if it matches the one from the installer
	if s.vmExists, err = s.client.Exists(client.ExistsParams{Name: s.vmName}); err != nil {
		return onError(err)
	} else {
		if s.vmExists {
			s.createVM = false
		}
	}

	if s.createVM {
		if err := s.createFromInstallerApp(ui, config); err != nil {
			return onError(err)
		}
	}

	show, err := s.client.Show(s.vmName)
	if err != nil {
		return onError(err)
	}

	if show.IsRunning() {
		if config.StopSourceVM {
			ui.Say(fmt.Sprintf("Stopping VM %s", s.vmName))
			if err := s.client.Stop(client.StopParams{VMName: s.vmName}); err != nil {
				return onError(err)
			}
		} else {
			ui.Say(fmt.Sprintf("Suspending VM %s", s.vmName))
			if err := s.client.Suspend(client.SuspendParams{VMName: s.vmName}); err != nil {
				return onError(err)
			}
		}
	}

	state.Put("source_vm_show", show)

	return multistep.ActionContinue
}

func (s *StepPrepareSourceVM) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	log.Println("Cleaning up prepare source VM step")
	if !s.createVM {
		return
	}

	_, halted := state.GetOk(multistep.StateHalted)
	_, canceled := state.GetOk(multistep.StateCancelled)
	errorObj := state.Get("error")
	switch errorObj.(type) {
	case *common.VMAlreadyExistsError:
		return
	case *common.VMNotFoundException:
		return
	default:
		if halted || canceled {
			ui.Say(fmt.Sprintf("Deleting VM %s", s.vmName))
			if err := s.client.Delete(client.DeleteParams{VMName: s.vmName}); err != nil {
				ui.Error(fmt.Sprint(err))
			}
			return
		}
	}

	if err := s.client.Suspend(client.SuspendParams{VMName: s.vmName}); err != nil {
		ui.Error(fmt.Sprint(err))
		_ = s.client.Delete(client.DeleteParams{VMName: s.vmName})
		panic(err)
	}
}

func (s *StepPrepareSourceVM) createFromInstallerApp(ui packer.Ui, config *Config) error {
	ui.Say(fmt.Sprintf("Creating a new base VM Template (%s) from installer, this will take a while", s.vmName))
	outputStream := make(chan string)
	go func() {
		for msg := range outputStream {
			ui.Say(msg)
		}
	}()
	createParams := client.CreateParams{
		InstallerApp: config.InstallerApp,
		Name:         s.vmName,
		DiskSize:     config.DiskSize,
		CPUCount:     config.CPUCount,
		RAMSize:      config.RAMSize,
	}

	if createParams.DiskSize == "" {
		createParams.DiskSize = DEFAULT_DISK_SIZE
	}

	if createParams.CPUCount == "" {
		createParams.CPUCount = DEFAULT_CPU_COUNT
	}

	if createParams.RAMSize == "" {
		createParams.RAMSize = DEFAULT_RAM_SIZE
	}

	if resp, err := s.client.Create(createParams, outputStream); err != nil {
		return err
	} else {
		ui.Say(fmt.Sprintf("VM %s was created (%s)", s.vmName, resp.UUID))
	}
	close(outputStream)
	return nil
}
