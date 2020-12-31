//go:generate mapstructure-to-hcl2 -type Config

package anka

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/veertuinc/packer-builder-veertu-anka/builder/anka"
	"github.com/veertuinc/packer-builder-veertu-anka/client"
)

const BuilderIdImport = "packer.post-processor.veertu-anka-registry"

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

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
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

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	if artifact.BuilderId() != anka.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from anka artifacts.",
			artifact.BuilderId())
		return nil, false, false, err
	}

	ankaClient := &client.Client{}

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
		VMName:      artifact.String(),
	}

	pushErr := ankaClient.RegistryPush(registryParams, pushParams)

	return artifact, true, false, pushErr
}
