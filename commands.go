package structflag

import (
	"errors"
	"fmt"
	"os"
)

type commandDetails struct {
	command     string
	description string
	action      func([]string) error
}

var (
	// ErrCommandNotFound is returned by Commands.Execute
	// if no matching command could be found.
	ErrCommandNotFound = errors.New("Command not found")

	// ErrNotEnoughArguments is returned when a command
	// is called with not enough aruments
	ErrNotEnoughArguments = errors.New("Not enough argumetns")
)

// Commands helps to parse and execute commands from command line arguments
type Commands []commandDetails

// AddWithArgs adds a command with additional string arguments
func (c *Commands) AddWithArgs(command, description string, action func([]string) error) {
	*c = append(*c, commandDetails{command, description, action})
}

// AddWithArg adds a command with a single additional string argument
func (c *Commands) AddWithArg(command, description string, action func(string) error) {
	c.AddWithArgs(command, description, func(args []string) error {
		if len(args) < 1 {
			return ErrNotEnoughArguments
		}
		return action(args[0])
	})
}

// AddWith2Args adds a command with two additional string argument
func (c *Commands) AddWith2Args(command, description string, action func(string, string) error) {
	c.AddWithArgs(command, description, func(args []string) error {
		if len(args) < 2 {
			return ErrNotEnoughArguments
		}
		return action(args[0], args[1])
	})
}

// AddWith3Args adds a command with threee additional string argument
func (c *Commands) AddWith3Args(command, description string, action func(string, string, string) error) {
	c.AddWithArgs(command, description, func(args []string) error {
		if len(args) < 3 {
			return ErrNotEnoughArguments
		}
		return action(args[0], args[1], args[2])
	})
}

// Add adds a command
func (c *Commands) Add(command, description string, action func() error) {
	c.AddWithArgs(command, description, func([]string) error { return action() })
}

// PrintDescription prints a description of all commands to stderr
func (c *Commands) PrintDescription() {
	fmt.Fprint(os.Stderr, "Commands:\n")
	for i, comm := range *c {
		if comm.description == "" {
			fmt.Fprintf(os.Stderr, "%d. %s\n\n", i+1, comm.command)
		} else {
			fmt.Fprintf(os.Stderr, "%d. %s: %s\n\n", i+1, comm.command, comm.description)
		}
	}
	fmt.Fprint(os.Stderr, "Flags:\n")
	PrintDefaults()
}

// Execute executes the command from args[0] or returns
// ErrCommandNotFound if no such command was registered
// of if len(args) == 0
func (c *Commands) Execute(args []string) error {
	if len(args) == 0 {
		return ErrCommandNotFound
	}
	command := args[0]
	for _, comm := range *c {
		if comm.command == command {
			return comm.action(args[1:])
		}
	}
	return ErrCommandNotFound
}
