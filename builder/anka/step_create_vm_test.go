package anka

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	c "github.com/veertuinc/packer-builder-veertu-anka/client"
	mocks "github.com/veertuinc/packer-builder-veertu-anka/mocks"
	"gotest.tools/assert"
)

func TestCreateVMRun(t *testing.T) {
	var createResponse c.CreateResponse
	err := json.Unmarshal(json.RawMessage(`{"UUID": "abcd-efgh-1234-5678"}`), &createResponse)
	if err != nil {
		t.Fail()
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	client := mocks.NewMockClient(mockCtrl)

	step := StepCreateVM{}
	ui := packer.TestUi(t)
	ctx := context.Background()
	state := new(multistep.BasicStateBag)

	state.Put("ui", ui)

	t.Run("create vm", func(t *testing.T) {
		config := &Config{
			DiskSize:     "500G",
			CPUCount:     "32G",
			RAMSize:      "16G",
			InstallerApp: "/Applications/Install macOS Big Sur.app/",
			VMName:       "foo",
			PackerConfig: common.PackerConfig{
				PackerBuilderType: "veertu-anka-vm-create",
			},
		}

		state.Put("config", config)
		state.Put("client", client)

		step.vmName = "foo-11.2-16.4.06"

		state.Put("vm_name", step.vmName)

		createParams := c.CreateParams{
			InstallerApp: config.InstallerApp,
			Name:         step.vmName,
			DiskSize:     config.DiskSize,
			CPUCount:     config.CPUCount,
			RAMSize:      config.RAMSize,
		}

		client.EXPECT().Create(createParams, gomock.Any()).Return(createResponse, nil).Times(1)

		mockui := packer.MockUi{}
		mockui.Say(fmt.Sprintf("Creating a new VM Template (%s) from installer, this will take a while", step.vmName))
		mockui.Say(fmt.Sprintf("VM %s was created (%s)", step.vmName, createResponse.UUID))

		stepAction := step.Run(ctx, state)
		assert.Equal(t, mockui.SayMessages[0].Message, "Creating a new VM Template (foo-11.2-16.4.06) from installer, this will take a while")
		assert.Equal(t, mockui.SayMessages[1].Message, "VM foo-11.2-16.4.06 was created (abcd-efgh-1234-5678)")
		assert.Equal(t, multistep.ActionContinue, stepAction)
	})

	t.Run("create vm with packer force", func(t *testing.T) {
		config := &Config{
			DiskSize:     "500G",
			CPUCount:     "32G",
			RAMSize:      "16G",
			InstallerApp: "/Applications/Install macOS Big Sur.app/",
			VMName:       "foo",
			PackerConfig: common.PackerConfig{
				PackerForce:       true,
				PackerBuilderType: "veertu-anka-vm-create",
			},
		}

		state.Put("config", config)
		state.Put("client", client)

		step.vmName = "foo-11.2-16.4.06"

		state.Put("vm_name", step.vmName)

		createParams := c.CreateParams{
			InstallerApp: config.InstallerApp,
			Name:         step.vmName,
			DiskSize:     config.DiskSize,
			CPUCount:     config.CPUCount,
			RAMSize:      config.RAMSize,
		}

		gomock.InOrder(
			client.EXPECT().Exists(step.vmName).Return(true, nil).Times(1),
			client.EXPECT().Delete(c.DeleteParams{VMName: step.vmName}).Return(nil).Times(1),
			client.EXPECT().Create(createParams, gomock.Any()).Return(createResponse, nil).Times(1),
		)

		mockui := packer.MockUi{}
		mockui.Say(fmt.Sprintf("Deleting existing virtual machine %s", step.vmName))
		mockui.Say(fmt.Sprintf("Creating a new VM Template (%s) from installer, this will take a while", step.vmName))
		mockui.Say(fmt.Sprintf("VM %s was created (%s)", step.vmName, createResponse.UUID))

		stepAction := step.Run(ctx, state)
		assert.Equal(t, mockui.SayMessages[0].Message, "Deleting existing virtual machine foo-11.2-16.4.06")
		assert.Equal(t, mockui.SayMessages[1].Message, "Creating a new VM Template (foo-11.2-16.4.06) from installer, this will take a while")
		assert.Equal(t, mockui.SayMessages[2].Message, "VM foo-11.2-16.4.06 was created (abcd-efgh-1234-5678)")
		assert.Equal(t, multistep.ActionContinue, stepAction)
	})
}
