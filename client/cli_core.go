package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/veertuinc/packer-builder-veertu-anka/common"
)

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#cp
type CopyParams struct {
	Src string
	Dst string
}

func (c *Client) Copy(params CopyParams) error {
	_, err := runAnkaCommand("cp", "-pRLf", params.Src, params.Dst)
	return err
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#clone
type CloneParams struct {
	VMName     string
	SourceUUID string
}

func (c *Client) Clone(params CloneParams) error {
	_, err := runAnkaCommand("clone", params.SourceUUID, params.VMName)
	if err != nil {
		merr, ok := err.(machineReadableError)
		if ok {
			if merr.Code == AnkaNameAlreadyExistsErrorCode {
				return &common.VMAlreadyExistsError{}
			}
		}
		return err
	}

	return nil
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#delete
type DeleteParams struct {
	VMName string
}

func (c *Client) Delete(params DeleteParams) error {
	args := []string{
		"delete",
		"--yes",
	}

	args = append(args, params.VMName)
	_, err := runAnkaCommand(args...)
	return err
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#create
type CreateParams struct {
	Name         string
	InstallerApp string
	OpticalDrive string
	RAMSize      string
	DiskSize     string
	CPUCount     string
}

type CreateResponse struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	CPUCores int    `json:"cpu_cores"`
	RAM      string `json:"ram"`
	ImageID  string `json:"image_id"`
	Status   string `json:"status"`
}

func (c *Client) Create(params CreateParams, outputStreamer chan string) (CreateResponse, error) {
	args := []string{
		"create",
		"--app", params.InstallerApp,
		params.Name,
	}
	output, err := runAnkaCommandStreamer(outputStreamer, args...)
	if err != nil {
		return CreateResponse{}, err
	}

	var response CreateResponse
	err = json.Unmarshal(output.Body, &response)
	if err != nil {
		return response, fmt.Errorf("Failed parsing output: %q (%v)", output.Body, err)
	}

	return response, nil
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#describe
type DescribeResponse struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	UUID    string `json:"uuid"`
	CPU     struct {
		Cores   int `json:"cores"`
		Threads int `json:"threads"`
	} `json:"cpu"`
	RAM string `json:"ram"`
	Usb struct {
		Tablet   int         `json:"tablet"`
		Kbd      int         `json:"kbd"`
		Host     interface{} `json:"host"`
		Location interface{} `json:"location"`
		PciSlot  int         `json:"pci_slot"`
		Mouse    int         `json:"mouse"`
	} `json:"usb"`
	OpticalDrives []interface{} `json:"optical_drives"`
	HardDrives    []struct {
		Controller string `json:"controller"`
		PciSlot    int    `json:"pci_slot"`
		File       string `json:"file"`
	} `json:"hard_drives"`
	NetworkCards []struct {
		Index               int    `json:"index"`
		Mode                string `json:"mode"`
		MacAddress          string `json:"mac_address"`
		PortForwardingRules []struct {
			GuestPort int    `json:"guest_port"`
			RuleName  string `json:"rule_name"`
			Protocol  string `json:"protocol"`
			HostIP    string `json:"host_ip"`
			HostPort  int    `json:"host_port"`
		} `json:"port_forwarding_rules"`
		PciSlot int    `json:"pci_slot"`
		Type    string `json:"type"`
	} `json:"network_cards"`
	Smbios struct {
		Type string `json:"type"`
	} `json:"smbios"`
	Smc struct {
		Type string `json:"type"`
	} `json:"smc"`
	Nvram    bool `json:"nvram"`
	Firmware struct {
		Type string `json:"type"`
	} `json:"firmware"`
	Display struct {
		Headless    int `json:"headless"`
		FrameBuffer struct {
			PciSlot  int    `json:"pci_slot"`
			VncPort  int    `json:"vnc_port"`
			Height   int    `json:"height"`
			Width    int    `json:"width"`
			VncIP    string `json:"vnc_ip"`
			Password string `json:"password"`
		} `json:"frame_buffer"`
	} `json:"display"`
}

func (c *Client) Describe(vmName string) (DescribeResponse, error) {
	output, err := runAnkaCommand("describe", vmName)
	if err != nil {
		return DescribeResponse{}, err
	}

	var response DescribeResponse
	err = json.Unmarshal(output.Body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#modify
func (c *Client) Modify(vmName string, command string, property string, flags ...string) error {
	ankaCommand := []string{"modify", vmName, command, property}
	ankaCommand = append(ankaCommand, flags...)
	output, err := runAnkaCommand(ankaCommand...)
	if err != nil {
		return err
	}
	if output.Status != "OK" {
		log.Print("Error executing modify command: ", output.ExceptionType, " ", output.Message)
		return fmt.Errorf(output.Message)
	}
	return nil
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#run
type RunParams struct {
	VMName         string
	Volume         string
	Command        []string
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	Debug          bool
	User           string
}

func (c *Client) Run(params RunParams) (error, int) {
	runner := NewRunner(params)
	err := runner.Start()
	if err != nil {
		return err, -1
	}

	log.Printf("Waiting for command to run")
	return runner.Wait()
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#show
type ShowResponse struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	CPUCores  int    `json:"cpu_cores"`
	RAM       string `json:"ram"`
	ImageID   string `json:"image_id"`
	Status    string `json:"status"`
	HardDrive uint64 `json:"hard_drive"`
}

func (sr ShowResponse) IsRunning() bool {
	return sr.Status == "running"
}

func (sr ShowResponse) IsStopped() bool {
	return sr.Status == "stopped"
}

func (c *Client) Show(vmName string) (ShowResponse, error) {
	output, err := runAnkaCommand("show", vmName)
	if err != nil {
		merr, ok := err.(machineReadableError)
		if ok {
			if merr.Code == AnkaVMNotFoundExceptionErrorCode {
				return ShowResponse{}, &common.VMNotFoundException{}
			}
		}
		return ShowResponse{}, err
	}

	var response ShowResponse
	err = json.Unmarshal(output.Body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#start
type StartParams struct {
	VMName string
}

func (c *Client) Start(params StartParams) error {
	_, err := runAnkaCommand("start", params.VMName)
	return err
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#stop
type StopParams struct {
	VMName string
	Force  bool
}

func (c *Client) Stop(params StopParams) error {
	args := []string{
		"stop",
	}

	if params.Force {
		args = append(args, "--force")
	}

	args = append(args, params.VMName)
	_, err := runAnkaCommand(args...)
	return err
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#suspend
type SuspendParams struct {
	VMName string
}

func (c *Client) Suspend(params SuspendParams) error {
	_, err := runAnkaCommand("suspend", params.VMName)
	return err
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#version
type VersionResponse struct {
	Status string              `json:"status"`
	Body   VersionResponseBody `json:"body"`
}

type VersionResponseBody struct {
	Product string `json:"product"`
	Version string `json:"version"`
	Build   string `json:"build"`
}

func (c *Client) Version() (VersionResponse, error) {
	var response VersionResponse

	out, err := exec.Command("anka", "--machine-readable", "version").Output()
	if err != nil {
		return response, err
	}

	err = json.Unmarshal([]byte(out), &response)
	return response, err
}
