package ankavm

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/groob/plist"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/random"
	"github.com/veertuinc/packer-builder-veertu-anka/client"
	"github.com/veertuinc/packer-builder-veertu-anka/common"
)

type StepCreateVM struct {
	client *client.Client
	vmName string
}

func (s *StepCreateVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	source_vm := state.Get("source_vm_show").(client.ShowResponse)
	s.client = state.Get("client").(*client.Client)

	onError := func(err error) multistep.StepAction {
		return stepError(ui, state, err)
	}

	clonedVMName := config.VMName
	if clonedVMName == "" { // If user doesn't give a vm_name, generate one
		clonedVMName = fmt.Sprintf("anka-packer-%s", random.AlphaNum(10))
	}
	s.vmName = clonedVMName
	state.Put("vm_name", clonedVMName)

	// If the user forces the build (packer build --force), delete the existing VM that would fail the build
	exists, err := s.client.Exists(client.ExistsParams{Name: clonedVMName})
	if exists && config.PackerForce {
		ui.Say(fmt.Sprintf("Deleting existing virtual machine %s", clonedVMName))
		if err = s.client.Delete(client.DeleteParams{VMName: clonedVMName}); err != nil {
			return onError(err)
		}
	}
	if err != nil {
		return onError(err)
	}

	ui.Say(fmt.Sprintf("Cloning source VM %s into a new virtual machine: %s", source_vm.Name, clonedVMName))
	if err = s.client.Clone(client.CloneParams{VMName: clonedVMName, SourceUUID: source_vm.UUID}); err != nil {
		return onError(err)
	}

	showResponse, err := s.client.Show(clonedVMName)
	if err != nil {
		return onError(err)
	}

	if err := s.modifyVMResources(showResponse, config, ui); err != nil {
		return onError(err)
	}

	describeResponse, err := s.client.Describe(clonedVMName)
	if err != nil {
		return onError(err)
	}
	state.Put("instance_id", describeResponse.UUID) // Expose the VM UUID as the "ID" contextual build variable

	if err := s.modifyVMProperties(describeResponse, showResponse, config, ui); err != nil {
		return onError(err)
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

func (s *StepCreateVM) modifyVMProperties(describeResponse client.DescribeResponse, showResponse client.ShowResponse, config *Config, ui packer.Ui) error {
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
			if err := s.client.Stop(stopParams); err != nil {
				return err
			}
			err := s.client.Modify(showResponse.Name, "add", "port-forwarding", "--host-port", strconv.Itoa(wantedPortForwardingRule.PortForwardingHostPort), "--guest-port", strconv.Itoa(wantedPortForwardingRule.PortForwardingGuestPort), wantedPortForwardingRule.PortForwardingRuleName)
			if !config.PackerForce { // If force is enabled, just skip
				if err != nil {
					return err
				}
			}
		}
	}

	if config.HWUUID != "" {
		if err := s.client.Stop(stopParams); err != nil {
			return err
		}
		ui.Say(fmt.Sprintf("Modifying VM custom-variable hw.UUID to %s", config.HWUUID))
		err := s.client.Modify(showResponse.Name, "set", "custom-variable", "hw.UUID", config.HWUUID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StepCreateVM) modifyVMResources(showResponse client.ShowResponse, config *Config, ui packer.Ui) error {
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
			if err := s.client.Stop(stopParams); err != nil {
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
			if err := s.client.Stop(stopParams); err != nil { // Prevent 'VM is already running' error
				return err
			}
		}
		if diskSizeBytes < showResponse.HardDrive {
			return fmt.Errorf("Shrinking VM disks is not allowed! Source VM Disk Size: %v", convertDiskSizeFromBytes(showResponse.HardDrive))
		}
	}

	if config.RAMSize != "" && config.RAMSize != showResponse.RAM {
		if err := s.client.Stop(stopParams); err != nil {
			return err
		}
		ui.Say(fmt.Sprintf("Modifying VM %s RAM to %s", showResponse.Name, config.RAMSize))
		err := s.client.Modify(showResponse.Name, "set", "ram", config.RAMSize)
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
			if err := s.client.Stop(stopParams); err != nil {
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
	err = plist.NewXMLDecoder(plistContent).Decode(&installAppPlist)
	if err != nil {
		return "", fmt.Errorf("failed to decode app info plist content: %w", err)
	}

	return fmt.Sprintf("%s-%s", installAppPlist.PlatformVersion, installAppPlist.ShortVersion), nil
}
