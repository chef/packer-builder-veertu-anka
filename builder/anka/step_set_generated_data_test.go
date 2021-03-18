package anka

import (
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

		gomock.InOrder(
			client.EXPECT().RunWithOutput(
				c.RunParams{
					Command: []string{"run", step.vmName, "uname", "-r"},
				},
			).Times(1),
			client.EXPECT().RunWithOutput(
				c.RunParams{
					Command: []string{"run", step.vmName, "sw_vers", "-productVersion"},
				},
			).Times(1),
		)

		stepAction := step.Run(ctx, state)

		assert.Equal(t, multistep.ActionContinue, stepAction)
	})

	t.Run("expose variables when create vm was used", func(t *testing.T) {
		step.vmName = "foo-11.2-16.4.06"

		state.Put("vm_name", step.vmName)
		state.Put("client", client)
		state.Put("os_version", "11.2")

		client.EXPECT().RunWithOutput(
			c.RunParams{
				Command: []string{"run", step.vmName, "uname", "-r"},
			},
		).Times(1)

		stepAction := step.Run(ctx, state)

		assert.Equal(t, multistep.ActionContinue, stepAction)
	})
}
