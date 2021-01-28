//go:generate mapstructure-to-hcl2 -type Config,PortForwardingRule
package ankavm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/random"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

const (
	DEFAULT_BOOT_DELAY         = "10s"
	DEFAULT_SOURCE_VM_BEHAVIOR = "suspend"
)

type PortForwardingRule struct {
	PortForwardingGuestPort int    `mapstructure:"port_forwarding_guest_port"`
	PortForwardingHostPort  int    `mapstructure:"port_forwarding_host_port"`
	PortForwardingRuleName  string `mapstructure:"port_forwarding_rule_name"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	InstallerApp string `mapstructure:"installer_app" required:"false"`
	SourceVMName string `mapstructure:"source_vm_name" required:"false"`

	VMName   string `mapstructure:"vm_name" required:"true"`
	DiskSize string `mapstructure:"disk_size" required:"false"`
	RAMSize  string `mapstructure:"ram_size" required:"false"`
	CPUCount string `mapstructure:"cpu_count" required:"false"`

	SourceVMBehavior string `mapstructure:"source_vm_behavior" required:"false"`

	PortForwardingRules []PortForwardingRule `mapstructure:"port_forwarding_rules,omitempty" required:"false"`

	HWUUID     string `mapstructure:"hw_uuid,omitempty" required:"false"`
	BootDelay  string `mapstructure:"boot_delay" required:"false"`
	EnableHtt  bool   `mapstructure:"enable_htt" required:"false"`
	DisableHtt bool   `mapstructure:"disable_htt" required:"false"`
	UseAnkaCP  bool   `mapstructure:"use_anka_cp" required:"false"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, error) {
	var c Config

	var md mapstructure.Metadata
	err := config.Decode(&c, &config.DecodeOpts{
		PluginType:         BuilderId,
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, err
	}
	// Accumulate any errors
	var errs *packer.MultiError

	// Default to the normal anka communicator type
	if c.Comm.Type == "" {
		c.Comm.Type = "anka"
	}

	if c.InstallerApp == "" && c.SourceVMName == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("installer_app or source_vm_name must be specified"))
	}

	// Handle Port Forwarding Rules
	if len(c.PortForwardingRules) > 0 {
		for index, rule := range c.PortForwardingRules {
			if rule.PortForwardingGuestPort == 0 {
				errs = packer.MultiErrorAppend(errs, errors.New("guest port is required"))
			}
			if rule.PortForwardingRuleName == "" {
				c.PortForwardingRules[index].PortForwardingRuleName = random.AlphaNum(10)
			}
		}
	}

	if c.SourceVMBehavior == "" {
		c.SourceVMBehavior = DEFAULT_SOURCE_VM_BEHAVIOR
	}

	invalidSourceVMBehavior := true
	supportedSourceVMBehaviors := []string{"suspend", "stop"}
	for _, behavior := range supportedSourceVMBehaviors {
		if behavior == c.SourceVMBehavior {
			invalidSourceVMBehavior = false
		}
	}
	if invalidSourceVMBehavior {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("%s is not a supported source_vm_behavior", c.SourceVMBehavior))
	}

	if strings.ContainsAny(c.SourceVMName, " \n") {
		errs = packer.MultiErrorAppend(errs, errors.New("source_vm_name name contains spaces"))
	}

	if c.BootDelay == "" {
		c.BootDelay = DEFAULT_BOOT_DELAY
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return &c, nil
}
