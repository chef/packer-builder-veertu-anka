package anka

import (
	"context"
	"testing"

	"github.com/veertuinc/packer-builder-veertu-anka/client"
)

func TestRunEnabled(t *testing.T) {
	expectedResults := make(map[string]client.MachineReadableOutput)

	expectedErrors := make(map[string]error)

	cli := &client.FakeCLI{
		Results: expectedResults,
		Errors:  expectedErrors,
	}
	client := client.Client{cli: cli}

	step := StepSetHyperThreading{}

	ctx = context.Context
	state =

	stepAction := step.Run(ctx, state)

	assert.Equal(t, stepAction, multistep.ActionContinue)

	assert.Equal(t, "modify foo set cpu --htt", cli.Commands[0])
	assert.Equal(t, "start foo", cli.Commands[1])
}
