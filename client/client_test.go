package client

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
)

func TestSuspend(t *testing.T) {
	expectedResults := make(map[string]machineReadableOutput)
	expectedResults["suspend foo"] = machineReadableOutput{}
	expectedResults["suspend bar"] = machineReadableOutput{}

	expectedErrors := make(map[string]error)
	expectedErrors["suspend foo"] = nil
	expectedErrors["suspend bar"] = nil

	cli := &FakeCLI{
		Results: expectedResults,
		Errors:  expectedErrors,
	}
	client := Client{cli: cli}

	errOne := client.Suspend(SuspendParams{VMName: "foo"})
	errTwo := client.Suspend(SuspendParams{VMName: "bar"})

	assert.Equal(t, "suspend foo", cli.Commands[0])
	assert.Equal(t, "suspend bar", cli.Commands[1])
	assert.Nil(t, errOne)
	assert.Nil(t, errTwo)
}

func TestStart(t *testing.T) {
	expectedResults := make(map[string]machineReadableOutput)
	expectedResults["start foo"] = machineReadableOutput{}
	expectedResults["start bar"] = machineReadableOutput{}

	expectedErrors := make(map[string]error)
	expectedErrors["start foo"] = nil
	expectedErrors["start bar"] = errors.New("I suck!")

	cli := &FakeCLI{
		Results: expectedResults,
		Errors:  expectedErrors,
	}
	client := Client{cli: cli}

	errOne := client.Start(StartParams{VMName: "foo"})
	errTwo := client.Start(StartParams{VMName: "bar"})

	assert.Equal(t, "start foo", cli.Commands[0])
	assert.Equal(t, "start bar", cli.Commands[1])
	assert.Nil(t, errOne)
	assert.Nil(t, errTwo)
}
