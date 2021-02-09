package client

import "strings"

type FakeCLI struct {
	Commands []string
	Results  map[string]machineReadableOutput
	Errors   map[string]error
}

func (c *FakeCLI) runCommand(args ...string) (machineReadableOutput, error) {
	fullCmd := strings.Join(args, " ")
	c.Commands = append(c.Commands, fullCmd)
	return c.Results[fullCmd], c.Errors[fullCmd]
}

func (c *FakeCLI) runCommandStreamer(outputStreamer chan string, args ...string) (machineReadableOutput, error) {
	return c.Results[args[0]], c.Errors[args[0]]
}
