package anka

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	c "github.com/veertuinc/packer-builder-veertu-anka/client"
	"github.com/veertuinc/packer-builder-veertu-anka/testutils"
	"gotest.tools/assert"
)

func TestRun(t *testing.T) {
	step := StepSetHyperThreading{}
	ui := packer.TestUi(t)
	ctx := context.Background()
	state := new(multistep.BasicStateBag)

	expectedResults := make(map[string]c.MachineReadableOutput)
	expectedErrors := make(map[string]error)

	cli := &testutils.FakeCLI{
		Results: expectedResults,
		Errors:  expectedErrors,
	}

	client := &c.Client{Cli: cli}

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
		expectedResults["describe foo"] = c.MachineReadableOutput{
			Body: json.RawMessage(`{}`),
		}
		expectedResults["show foo"] = c.MachineReadableOutput{
			Body: json.RawMessage(`{}`),
		}
		expectedResults["stop --force foo"] = c.MachineReadableOutput{}
		expectedResults["modify foo set cpu --htt"] = c.MachineReadableOutput{
			Status: "OK",
		}

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

		assert.Equal(t, "describe foo", cli.Commands[0])
		assert.Equal(t, "show foo", cli.Commands[1])
		assert.Equal(t, "stop --force foo", cli.Commands[2])
		assert.Equal(t, "modify foo set cpu --htt", cli.Commands[3])

		assert.Equal(t, stepAction, multistep.ActionContinue)
	})
}
