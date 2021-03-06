// Package wrap provides wrapper executables for customizing benchmark execution.
package wrap

import (
	"context"
	"flag"
	"os"
	"runtime"

	"github.com/google/subcommands"

	"github.com/mmcloughlin/goperf/pkg/command"
	"github.com/mmcloughlin/goperf/pkg/proc"
	"github.com/mmcloughlin/goperf/pkg/runner"
)

type action interface {
	SetFlags(*flag.FlagSet)
	Apply() error
}

type wrapper struct {
	command.Base
	name     string
	synopsis string
	actions  []action
}

func (cmd *wrapper) Name() string     { return cmd.name }
func (cmd *wrapper) Synopsis() string { return cmd.synopsis }

func (cmd *wrapper) Usage() string {
	// TODO(mbm): set wrapper command usage strings
	return ""
}

func (cmd *wrapper) SetFlags(f *flag.FlagSet) {
	for _, a := range cmd.actions {
		a.SetFlags(f)
	}
}

func (cmd *wrapper) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Sub-process arguments.
	args := f.Args()

	// Lock goroutine to thread before applying actions.
	runtime.LockOSThread()

	for _, a := range cmd.actions {
		if err := a.Apply(); err != nil {
			return cmd.Error(err)
		}
	}

	// Execute the sub-process.
	return cmd.Status(proc.Exec(args))
}

// RunUnder builds a wrapper that runs under the given subcommand, assuming that
// subcommand is registered on this executable.
func RunUnder(cmd subcommands.Command, arg ...string) (runner.Wrapper, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	args := []string{cmd.Name()}
	args = append(args, arg...)
	return runner.RunUnder(self, args...), nil
}
