package client

import (
	// "syscall"

	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type Runner struct {
	params  RunParams
	cmd     *exec.Cmd
	started time.Time
}

func NewRunner(params RunParams) *Runner {
	args := []string{}

	if params.Debug {
		args = append(args, "--log-level", "debug")
	}

	if params.Stdout == nil {
		params.Stdout = os.Stdout
	}

	if params.Stderr == nil {
		params.Stderr = os.Stderr
	}

	args = append(args, "run")

	if params.Volume != "" {
		args = append(args, "-v", params.Volume)
	} else {
		args = append(args, "-n")
	}

	args = append(args, params.VMName)
	args = append(args, "sh")

	cmd := exec.Command("anka", args...)
	cmd.Stdout = params.Stdout
	cmd.Stderr = params.Stderr

	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	Setpgid: true,
	// 	Pgid:    0,
	// }

	return &Runner{
		params: params,
		cmd:    cmd,
	}
}

func (r *Runner) Start() error {
	log.Printf("Starting command: %s", strings.Join(r.cmd.Args, " "))
	r.started = time.Now()
	stdin, err := r.cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = r.cmd.Start()
	if err != nil {
		return err
	}
	cmdString := strings.Join(r.params.Command, " ")
	log.Print("Executing on sh: ", cmdString)
	_, _ = stdin.Write([]byte(cmdString))
	stdin.Close()
	return err
}

func (r *Runner) Wait() (error, int) {
	err := r.cmd.Wait()
	log.Printf("Command finished in %s with %v", time.Since(r.started), err)
	return err, getExitCode(err)
}

// GetExitCode extracts an exit code from an error where the platform supports it,
// otherwise returns 0 for no error and 1 for an error
func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	if eerr, ok := err.(*exec.ExitError); ok {
		code := eerr.ExitCode()
		if code == 125 {
			code = packer.CmdDisconnect
		}
		return code
	}

	return 1
}
