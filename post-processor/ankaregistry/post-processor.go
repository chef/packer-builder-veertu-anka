//go:generate mapstructure-to-hcl2 -type Config

package ankaregistry

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/veertuinc/packer-builder-veertu-anka/builder/ankavm"
	"github.com/veertuinc/packer-builder-veertu-anka/client"
)

const BuilderIdRegistry = "packer.post-processor.veertu-anka-registry"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	RegistryName string `mapstructure:"remote"`
	RegistryURL  string `mapstructure:"registry-path"`
	NodeCertPath string `mapstructure:"cert"`
	NodeKeyPath  string `mapstructure:"key"`
	CaRootPath   string `mapstructure:"cacert"`
	IsInsecure   bool   `mapstructure:"insecure"`

	Tag         string `mapstructure:"tag"`
	Description string `mapstructure:"description"`
	RemoteVM    string `mapstructure:"remote-vm"`
	Local       bool   `mapstructure:"local"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         BuilderIdRegistry,
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.Tag == "" {
		return fmt.Errorf("You must specify a valid tag for your Veertu Anka VM (e.g. 'latest')")
	}

	log.Printf("%+v\n", p.config)

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	if artifact.BuilderId() != ankavm.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from anka artifacts.",
			artifact.BuilderId())
		return nil, false, false, err
	}

	ankaClient := client.NewClient()

	registryParams := client.RegistryParams{
		RegistryName: p.config.RegistryName,
		RegistryURL:  p.config.RegistryURL,
		NodeCertPath: p.config.NodeCertPath,
		NodeKeyPath:  p.config.NodeKeyPath,
		CaRootPath:   p.config.CaRootPath,
		IsInsecure:   p.config.IsInsecure,
	}

	pushParams := client.RegistryPushParams{
		Tag:         p.config.Tag,
		Description: p.config.Description,
		RemoteVM:    p.config.RemoteVM,
		Local:       p.config.Local,
		VMName:      artifact.String(),
	}

	// If force is true, revert the template tag (if one exists) on the registry so we can push the VM without issue
	if p.config.PackerForce {
		var id string

		templates, err := ankaClient.RegistryList(registryParams)
		if err != nil {
			return nil, false, false, err
		}

		for i := 0; i < len(templates); i++ {
			if templates[i].Name == artifact.String() {
				id = templates[i].ID
				ui.Say(fmt.Sprintf("Found existing template %s on registry that matches name '%s'", id, artifact.String()))
				break
			}
		}

		if id != "" {
			if err := ankaClient.RegistryRevert(id); err != nil {
				return nil, false, false, err
			}
			ui.Say(fmt.Sprintf("Reverted latest tag for template '%s' on registry", id))
		}
	}

	pushErr := ankaClient.RegistryPush(registryParams, pushParams)

	return artifact, true, false, pushErr
}
