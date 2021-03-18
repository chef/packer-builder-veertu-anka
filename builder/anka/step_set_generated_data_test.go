package anka

import (
	"bytes"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	c "github.com/veertuinc/packer-builder-veertu-anka/client"
	mocks "github.com/veertuinc/packer-builder-veertu-anka/mocks"
	"gotest.tools/assert"
)

var (
	osv, darwinVersion c.RunParams
)

func TestSetGeneratedDataRun(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	client := mocks.NewMockClient(mockCtrl)
	util := mocks.NewMockUtil(mockCtrl)

	state := new(multistep.BasicStateBag)
	step := StepSetGeneratedData{
		GeneratedData: &packerbuilderdata.GeneratedData{State: state},
	}
	ui := packer.TestUi(t)
	ctx := context.Background()

	state.Put("ui", ui)
	state.Put("util", util)

	t.Run("expose variables", func(t *testing.T) {
		step.vmName = "foo-11.2-16.4.06"

		state.Put("vm_name", step.vmName)
		state.Put("client", client)

		darwinVersion := c.RunParams{
			Command: []string{"/usr/bin/uname", "-r"},
			VMName:  step.vmName,
			Stdout:  &bytes.Buffer{},
		}

		osv := c.RunParams{
			Command: []string{"/usr/bin/sw_vers", "-productVersion"},
			VMName:  step.vmName,
			Stdout:  &bytes.Buffer{},
		}

		gomock.InOrder(
			client.EXPECT().Run(darwinVersion).Times(1),
			client.EXPECT().Run(osv).Times(1),
		)

		stepAction := step.Run(ctx, state)
		assert.Equal(t, multistep.ActionContinue, stepAction)
	})

	t.Run("expose variables when create vm was used", func(t *testing.T) {
		step.vmName = "foo-11.2-16.4.06"

		state.Put("vm_name", step.vmName)
		state.Put("client", client)
		state.Put("os_version", "11.2")

		darwinVersion := c.RunParams{
			Command: []string{"/usr/bin/uname", "-r"},
			VMName:  step.vmName,
			Stdout:  &bytes.Buffer{},
		}

		client.EXPECT().Run(darwinVersion).Times(1)

		stepAction := step.Run(ctx, state)
		assert.Equal(t, multistep.ActionContinue, stepAction)
	})
}
