package anka

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/groob/plist"
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

type StepCreateVM struct {
	client client.Client
	vmName string
}

func (s *StepCreateVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	onError := func(err error) multistep.StepAction {
		return stepError(ui, state, err)
	}

	s.client = state.Get("client").(client.Client)
	s.vmName = config.VMName

	macOSVersionFromInstallerApp, err := obtainMacOSVersionFromInstallerApp(config.InstallerApp)
	if err != nil {
		return onError(err)
	}

	s.vmName = fmt.Sprintf("%s-%s", s.vmName, macOSVersionFromInstallerApp)

	state.Put("vm_name", s.vmName)

	if config.AnkaPassword != "" {
		os.Setenv("ANKA_DEFAULT_PASSWD", config.AnkaPassword)
	}

	if config.AnkaUser != "" {
		os.Setenv("ANKA_DEFAULT_USER", config.AnkaUser)
	}

	if config.PackerForce {
		exists, err := s.client.Exists(s.vmName)
		if err != nil {
			return onError(err)
		}

		if exists {
			ui.Say(fmt.Sprintf("Deleting existing virtual machine %s", s.vmName))

			err = s.client.Delete(client.DeleteParams{VMName: s.vmName})
			if err != nil {
				return onError(err)
			}
		}
	}

	err = s.createFromInstallerApp(ui, config)
	if err != nil {
		return onError(err)
	}

	show, err := s.client.Show(s.vmName)
	if err != nil {
		return onError(err)
	}

	if show.IsRunning() {
		if config.StopVM {
			ui.Say(fmt.Sprintf("Stopping VM %s", s.vmName))

			err := s.client.Stop(client.StopParams{VMName: s.vmName})
			if err != nil {
				return onError(err)
			}
		} else {
			ui.Say(fmt.Sprintf("Suspending VM %s", s.vmName))

			err := s.client.Suspend(client.SuspendParams{VMName: s.vmName})
			if err != nil {
				return onError(err)
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	log.Println("Cleaning up create VM step")
	if s.vmName == "" {
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

			err := s.client.Delete(client.DeleteParams{VMName: s.vmName})
			if err != nil {
				ui.Error(fmt.Sprint(err))
			}
			return
		}
	}

	err := s.client.Suspend(client.SuspendParams{
		VMName: s.vmName,
	})
	if err != nil {
		ui.Error(fmt.Sprint(err))

		deleteErr := s.client.Delete(client.DeleteParams{VMName: s.vmName})
		if deleteErr != nil {
			panic(deleteErr)
		}

		panic(err)
	}
}

func obtainMacOSVersionFromInstallerApp(path string) (string, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("installer app does not exist at %q: %w", path, err)
	}
	if err != nil {
		return "", fmt.Errorf("failed to stat installer at %q: %w", path, err)
	}

	plistPath := filepath.Join(path, "Contents", "Info.plist")
	_, err = os.Stat(plistPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("installer app info plist did not exist at %q: %w", plistPath, err)
	}
	if err != nil {
		return "", fmt.Errorf("failed to stat installer app info plist at %q: %w", plistPath, err)
	}
	plistContent, _ := os.Open(plistPath)

	var installAppPlist struct {
		PlatformVersion string `plist:"DTPlatformVersion"`
		ShortVersion    string `plist:"CFBundleShortVersionString"`
	}
	if err = plist.NewXMLDecoder(plistContent).Decode(&installAppPlist); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", installAppPlist.PlatformVersion, installAppPlist.ShortVersion), nil
}

func (s *StepCreateVM) createFromInstallerApp(ui packer.Ui, config *Config) error {
	ui.Say(fmt.Sprintf("Creating a new VM Template (%s) from installer, this will take a while", s.vmName))

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

	resp, err := s.client.Create(createParams, outputStream)
	if err != nil {
		return err
	}

	ui.Say(fmt.Sprintf("VM %s was created (%s)", s.vmName, resp.UUID))

	close(outputStream)

	return nil
}
