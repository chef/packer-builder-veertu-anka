package anka

import (
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/veertuinc/packer-builder-veertu-anka/testutils"
	"gotest.tools/assert"
)

func TestRun(t *testing.T) {
	step := StepSetHyperThreading{}
	ui := packer.TestUi(t)
	ctx := context.Background()
	state := new(multistep.BasicStateBag)

	expectedErrors := make(map[string]error)

	client := &testutils.TestClient{
		Errors: expectedErrors,
	}

	state.Put("client", client)
	state.Put("ui", ui)
	state.Put("vm_name", "foo")

	t.Run("test disabled or nil htt values", func(t *testing.T) {
		config := &Config{
			EnableHtt:  false,
			DisableHtt: false,
		}

		state.Put("config", config)

		stepAction := step.Run(ctx, state)
		assert.Equal(t, stepAction, multistep.ActionContinue)
	})

	t.Run("conflicting htt enables", func(t *testing.T) {
		config := &Config{
			EnableHtt:  true,
			DisableHtt: true,
		}

		state.Put("config", config)

		stepAction := step.Run(ctx, state)
		assert.Equal(t, stepAction, multistep.ActionHalt)
	})

	t.Run("test enable htt", func(t *testing.T) {
		expectedErrors["describe foo"] = nil
		expectedErrors["show foo"] = nil
		expectedErrors["stop --force foo"] = nil
		expectedErrors["modify foo set cpu --htt"] = nil

		config := &Config{
			EnableHtt:  true,
			DisableHtt: false,
		}

		state.Put("config", config)

		stepAction := step.Run(ctx, state)

		assert.Equal(t, "describe foo", client.Commands[0])
		assert.Equal(t, "show foo", client.Commands[1])
		assert.Equal(t, "stop --force foo", client.Commands[2])
		assert.Equal(t, "modify foo set cpu --htt", client.Commands[3])

		assert.Equal(t, stepAction, multistep.ActionContinue)
	})

	t.Run("test disable htt with 0 threads", func(t *testing.T) {
		expectedErrors["describe foo"] = nil
		expectedErrors["show foo"] = nil
		expectedErrors["stop --force foo"] = nil
		expectedErrors["modify foo set cpu --no-htt"] = nil

		config := &Config{
			EnableHtt:  false,
			DisableHtt: true,
		}

		state.Put("config", config)

		stepAction := step.Run(ctx, state)

		assert.Equal(t, "describe foo", client.Commands[0])

		assert.Equal(t, stepAction, multistep.ActionContinue)
	})

	t.Run("test disable htt with > 0 threads", func(t *testing.T) {
		expectedErrors["describe foo"] = nil

		expectedErrors["show foo"] = nil
		expectedErrors["stop --force foo"] = nil
		expectedErrors["modify foo set cpu --no-htt"] = nil

		config := &Config{
			EnableHtt:  false,
			DisableHtt: true,
		}

		state.Put("config", config)

		stepAction := step.Run(ctx, state)

		assert.Equal(t, "describe foo", client.Commands[0])
		// need to produce an output so that we can make sure we are analyzing the MachineReadableOutput
		// assert.Equal(t, "show foo", client.Commands[1])
		// assert.Equal(t, "stop --force foo", client.Commands[2])
		// assert.Equal(t, "modify foo set cpu --no-htt", client.Commands[3])

		assert.Equal(t, stepAction, multistep.ActionContinue)
	})
}
