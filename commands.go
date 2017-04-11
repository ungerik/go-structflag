package structflag

import (
	"errors"
	"strings"

	"github.com/fatih/color"
)

type commandDetails struct {
	command     string
	argDesc     string
	commandDesc []string
	action      func([]string) error
}

var (
	// Commands holds the global list of app commands.
	// A command is an action executed by the application,
	// something that is not saved in a configuration file.
	Commands CommandList

	// ErrNotEnoughArguments is returned when a command
	// is called with not enough aruments
	ErrNotEnoughArguments = errors.New("Not enough argumetns")

	// CommandUsageColor is the color in which the
	// command usage will be printed on the screen.
	CommandUsageColor = color.New(color.FgBlue)
)

// CommandList helps to parse and execute commands from command line arguments
type CommandList []commandDetails

// AddWithArgs adds a command with additional string arguments
func (c *CommandList) AddWithArgs(action func([]string) error, command, argDesc string, commandDesc ...string) {
	*c = append(
		*c,
		commandDetails{
			command:     command,
			argDesc:     argDesc,
			commandDesc: commandDesc,
			action:      action,
		},
	)
}

// AddWithArg adds a command with a single additional string argument
func (c *CommandList) AddWithArg(action func(string) error, command, argDesc string, commandDesc ...string) {
	c.AddWithArgs(func(args []string) error {
		if len(args) < 1 {
			return ErrNotEnoughArguments
		}
		return action(args[0])
	}, command, argDesc, commandDesc...)
}

// AddWith2Args adds a command with two additional string argument
func (c *CommandList) AddWith2Args(action func(string, string) error, command, argDesc string, commandDesc ...string) {
	c.AddWithArgs(func(args []string) error {
		if len(args) < 2 {
			return ErrNotEnoughArguments
		}
		return action(args[0], args[1])
	}, command, argDesc, commandDesc...)
}

// AddWith3Args adds a command with threee additional string argument
func (c *CommandList) AddWith3Args(action func(string, string, string) error, command, argDesc string, commandDesc ...string) {
	c.AddWithArgs(func(args []string) error {
		if len(args) < 3 {
			return ErrNotEnoughArguments
		}
		return action(args[0], args[1], args[2])
	}, command, argDesc, commandDesc...)
}

// Add adds a command
func (c *CommandList) Add(action func() error, command string, description ...string) {
	c.AddWithArgs(func([]string) error { return action() }, command, "", description...)
}

// PrintUsage prints a description of all commands to Output
func (c *CommandList) PrintUsage() {
	for _, comm := range *c {
		CommandUsageColor.Fprintf(Output, "  %s %s %s\n", AppName, comm.command, comm.argDesc)
		if len(comm.commandDesc) == 0 {
			CommandUsageColor.Fprintln(Output)
		} else {
			for _, desc := range comm.commandDesc {
				CommandUsageColor.Fprintf(Output, "      %s\n", desc)
			}
		}
	}
}

// Execute executes the command from args[0] and returns the executed
// command name and the error returned from the command function.
// The error is ErrNotEnoughArguments if args did not have enough
// extra arguments for the command.
// Returns "", nil if no matching command was found, or if len(args) == 0
func (c *CommandList) Execute(args []string) (command string, exeErr error) {
	if len(args) == 0 {
		return "", nil
	}
	commLower := strings.ToLower(args[0])
	for _, details := range *c {
		if strings.ToLower(details.command) == commLower {
			return details.command, details.action(args[1:])
		}
	}
	return "", nil
}
