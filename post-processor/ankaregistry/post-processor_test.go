package ankaregistry

import (
	"context"
	"fmt"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/veertuinc/packer-builder-veertu-anka/builder/anka"
	c "github.com/veertuinc/packer-builder-veertu-anka/client"
	mocks "github.com/veertuinc/packer-builder-veertu-anka/mocks"
)

func TestAnkaRegistryPostProcessor(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	client := mocks.NewMockClient(mockCtrl)

	ui := packer.TestUi(t)

	config := Config{
		RegistryName: "go-mock-anka-registry",
		RegistryURL:  "mockurl:mockport",
		RemoteVM:     "foo",
		Tag:          "registry-push",
		Description:  "mock for testing anka registry push",
	}

	pp := PostProcessor{
		config: config,
	}

	artifact := &anka.Artifact{}

	pp.PostProcess(context.Background(), ui, artifact)

	registryParams := c.RegistryParams{
		RegistryName: config.RegistryName,
		RegistryURL:  config.RegistryURL,
	}

	pushParams := c.RegistryPushParams{
		Tag:         config.Tag,
		Description: config.Description,
		RemoteVM:    config.RemoteVM,
		Local:       false,
	}

	mockui := packer.MockUi{}
	mockui.Say(fmt.Sprintf("Pushing template to Anka Registry as %s with tag %s", config.RemoteVM, config.Tag))

	client.EXPECT().RegistryPush(registryParams, pushParams).Return(nil).Times(1)
	assert.Equal(t, mockui.SayMessages[0].Message, "Pushing template to Anka Registry as foo with tag registry-push")
}
