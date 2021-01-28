//go:generate mapstructure-to-hcl2 -type Config

package ankastop

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/veertuinc/packer-builder-veertu-anka/builder/ankavm"
	"github.com/veertuinc/packer-builder-veertu-anka/client"
)

const BuilderIdImport = "packer.post-processor.veertu-anka-stop"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

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

	err := ankaClient.Stop(client.StopParams{
		VMName: artifact.String(),
		Force:  p.config.PackerConfig.PackerForce,
	})

	return artifact, true, false, err
}
