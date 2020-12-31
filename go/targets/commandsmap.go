package targets

import (
	"fmt"
	"strings"
)

// CommandDefinition defines a command with a check and an exec function
type CommandDefinition struct {
	CheckFunc   func(interface{}, []interface{}) error
	ExecuteFunc func(interface{}, []interface{}) error
}

// Map allow easy definition of command with name and check/execute functions
type Map struct {
	commands map[string]CommandDefinition
	target   interface{}
}

// Init connects object to target and map of commands
func (cmdmap *Map) Init(target interface{}, commands map[string]CommandDefinition) {
	cmdmap.target = target
	cmdmap.commands = commands
}

// CheckCommand verifies parameters
func (cmdmap *Map) CheckCommand(command string, parameters []interface{}) error {
	lowercommand := strings.ToLower(command)

	cmd, ok := cmdmap.commands[lowercommand]

	if !ok {
		return fmt.Errorf("Invalid command %s", command)
	}

	return cmd.CheckFunc(cmdmap.target, parameters)
}

// ExecuteCommand invokes the function that implements the command
func (cmdmap *Map) ExecuteCommand(command string, parameters []interface{}) error {
	lowercommand := strings.ToLower(command)

	cmd, ok := cmdmap.commands[lowercommand]

	if !ok {
		return fmt.Errorf("Invalid command %s", command)
	}

	return cmd.ExecuteFunc(cmdmap.target, parameters)
}

// NoParmsCheck can be used to validate commands that have no parameters
func NoParmsCheck(target interface{}, parameters []interface{}) error {
	if len(parameters) != 0 {
		return fmt.Errorf("Command has no parameters")
	}
	return nil
}
