package client

// import (
// 	"strings"
// )

// type FakeCLI struct {
// 	Commands []string
// 	Results  map[string]client.MachineReadableOutput
// 	Errors   map[string]error
// }

// func (c *FakeCLI) RunCommand(args ...string) (client.MachineReadableOutput, error) {
// 	fullCmd := strings.Join(args, " ")
// 	c.Commands = append(c.Commands, fullCmd)
// 	return c.Results[fullCmd], c.Errors[fullCmd]
// }

// func (c *FakeCLI) RunCommandStreamer(outputStreamer chan string, args ...string) (client.MachineReadableOutput, error) {
// 	return c.Results[args[0]], c.Errors[args[0]]
// }
