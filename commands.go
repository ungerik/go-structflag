package structflag

import (
	"errors"
	"fmt"
	"strings"
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

	// ErrCommandNotFound is returned by CommandList.Execute
	// if no matching command could be found.
	ErrCommandNotFound = errors.New("Command not found")

	// ErrNotEnoughArguments is returned when a command
	// is called with not enough aruments
	ErrNotEnoughArguments = errors.New("Not enough argumetns")
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

// PrintUsage prints a description of all commands to stderr
func (c *CommandList) PrintUsage() {
	for _, comm := range *c {
		fmt.Fprintf(Output, "  %s %s %s\n", AppName, comm.command, comm.argDesc)
		if len(comm.commandDesc) == 0 {
			fmt.Fprintln(Output)
		} else {
			for _, desc := range comm.commandDesc {
				fmt.Fprintf(Output, "      %s\n", desc)
			}
		}
	}
}

// Execute executes the command from args[0] or returns
// ErrCommandNotFound if no such command was registered
// of if len(args) == 0
func (c *CommandList) Execute(args []string) error {
	if len(args) == 0 {
		return ErrCommandNotFound
	}
	command := strings.ToLower(args[0])
	for _, comm := range *c {
		if strings.ToLower(comm.command) == command {
			return comm.action(args[1:])
		}
	}
	return ErrCommandNotFound
}
