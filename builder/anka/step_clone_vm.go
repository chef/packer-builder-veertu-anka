package anka

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/veertuinc/packer-builder-veertu-anka/client"
	"github.com/veertuinc/packer-builder-veertu-anka/common"
)

type StepCloneVM struct {
	client client.Client
	vmName string
}

func (s *StepCloneVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	s.client = state.Get("client").(client.Client)
	s.vmName = config.VMName

	state.Put("vm_name", s.vmName)

	onError := func(err error) multistep.StepAction {
		return stepError(ui, state, err)
	}

	// If the user forces the build (packer build --force), delete the existing VM that would fail the build
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

	exists, err := s.client.Exists(config.SourceVMName)
	if err != nil {
		return onError(err)
	}
	if !exists {
		return onError(fmt.Errorf("source vm %s does not exist. create it before cloning", config.SourceVMName))
	}

	sourceShow, err := s.client.Show(config.SourceVMName)
	if err != nil {
		return onError(err)
	}

	ui.Say(fmt.Sprintf("Cloning source VM %s into a new virtual machine: %s", sourceShow.Name, s.vmName))

	err = s.client.Clone(client.CloneParams{VMName: s.vmName, SourceUUID: sourceShow.UUID})
	if err != nil {
		return onError(err)
	}

	clonedShow, err := s.client.Show(s.vmName)
	if err != nil {
		return onError(err)
	}

	err = s.modifyVMResources(clonedShow, config, ui)
	if err != nil {
		return onError(err)
	}

	err = s.modifyVMProperties(clonedShow, config, ui)
	if err != nil {
		return onError(err)
	}

	if clonedShow.IsRunning() {
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

func (s *StepCloneVM) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	log.Println("Cleaning up clone VM step")
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

	err := s.client.Suspend(client.SuspendParams{VMName: s.vmName})
	if err != nil {
		ui.Error(fmt.Sprint(err))

		_ = s.client.Delete(client.DeleteParams{VMName: s.vmName})

		panic(err)
	}
}

func (s *StepCloneVM) modifyVMResources(showResponse client.ShowResponse, config *Config, ui packer.Ui) error {
	stopParams := client.StopParams{
		VMName: showResponse.Name,
		Force:  true,
	}

	if config.DiskSize != "" {
		err, diskSizeBytes := convertDiskSizeToBytes(config.DiskSize)
		if err != nil {
			return err
		}

		if diskSizeBytes > showResponse.HardDrive {
			err := s.client.Stop(stopParams)
			if err != nil {
				return err
			}

			ui.Say(fmt.Sprintf("Modifying VM %s disk size to %s", showResponse.Name, config.DiskSize))

			err = s.client.Modify(showResponse.Name, "set", "hard-drive", "-s", config.DiskSize)
			if err != nil {
				return err
			}

			// Resize the inner VM disk too with diskutil
			err, _ = s.client.Run(client.RunParams{
				VMName:  showResponse.Name,
				Command: []string{"diskutil", "apfs", "resizeContainer", "disk1", "0"},
			})
			if err != nil {
				return err
			}

			// Prevent 'VM is already running' error
			err = s.client.Stop(stopParams)
			if err != nil {
				return err
			}
		}

		if diskSizeBytes < showResponse.HardDrive {
			return fmt.Errorf("Shrinking VM disks is not allowed! Source VM Disk Size (bytes): %v", showResponse.HardDrive)
		}
	}

	if config.RAMSize != "" && config.RAMSize != showResponse.RAM {
		err := s.client.Stop(stopParams)
		if err != nil {
			return err
		}

		ui.Say(fmt.Sprintf("Modifying VM %s RAM to %s", showResponse.Name, config.RAMSize))

		err = s.client.Modify(showResponse.Name, "set", "ram", config.RAMSize)
		if err != nil {
			return err
		}
	}

	if config.CPUCount != "" {
		stringCPUCount, err := strconv.ParseInt(config.CPUCount, 10, 32)
		if err != nil {
			return err
		}

		if int(stringCPUCount) != showResponse.CPUCores {
			err := s.client.Stop(stopParams)
			if err != nil {
				return err
			}

			ui.Say(fmt.Sprintf("Modifying VM %s CPU core count to %v", showResponse.Name, stringCPUCount))

			err = s.client.Modify(showResponse.Name, "set", "cpu", "-c", strconv.Itoa(int(stringCPUCount)))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *StepCloneVM) modifyVMProperties(showResponse client.ShowResponse, config *Config, ui packer.Ui) error {
	describeResponse, err := s.client.Describe(showResponse.Name)
	if err != nil {
		return err
	}

	stopParams := client.StopParams{
		VMName: showResponse.Name,
		Force:  true,
	}

	if len(config.PortForwardingRules) > 0 {
		// Check if the rule already exists
		existingForwardedPorts := make(map[int]struct{})
		for _, existingNetworkCard := range describeResponse.NetworkCards {
			for _, existingPortForwardingRule := range existingNetworkCard.PortForwardingRules {
				existingForwardedPorts[existingPortForwardingRule.HostPort] = struct{}{}
			}
		}

		for _, wantedPortForwardingRule := range config.PortForwardingRules {
			ui.Say(fmt.Sprintf("Ensuring %s port-forwarding (Guest Port: %s, Host Port: %s, Rule Name: %s)", showResponse.Name, strconv.Itoa(wantedPortForwardingRule.PortForwardingGuestPort), strconv.Itoa(wantedPortForwardingRule.PortForwardingHostPort), wantedPortForwardingRule.PortForwardingRuleName))
			// Check if host port is set already and warn the user
			if _, ok := existingForwardedPorts[wantedPortForwardingRule.PortForwardingHostPort]; ok {
				ui.Error(fmt.Sprintf("Found an existing host port rule (%s)! Skipping without setting...", strconv.Itoa(wantedPortForwardingRule.PortForwardingHostPort)))
				continue
			}

			err := s.client.Stop(stopParams)
			if err != nil {
				return err
			}

			err = s.client.Modify(showResponse.Name, "add", "port-forwarding", "--host-port", strconv.Itoa(wantedPortForwardingRule.PortForwardingHostPort), "--guest-port", strconv.Itoa(wantedPortForwardingRule.PortForwardingGuestPort), wantedPortForwardingRule.PortForwardingRuleName)
			// If force is enabled, just skip
			if !config.PackerConfig.PackerForce {
				if err != nil {
					return err
				}
			}
		}
	}

	if config.HWUUID != "" {
		err := s.client.Stop(stopParams)
		if err != nil {
			return err
		}

		ui.Say(fmt.Sprintf("Modifying VM custom-variable hw.UUID to %s", config.HWUUID))

		err = s.client.Modify(showResponse.Name, "set", "custom-variable", "hw.UUID", config.HWUUID)
		if err != nil {
			return err
		}
	}

	return nil
}
