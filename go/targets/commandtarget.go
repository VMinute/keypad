package targets

import (
	"fmt"
)

// CommandTarget defines an object that can execute commands
type CommandTarget interface {
	Init(configyaml []byte) error                                  // reads configuration and checks if target is available
	CheckCommand(command string, parameters []interface{}) error   // validates a command
	ExecuteCommand(command string, parameters []interface{}) error //executes a command
}

// CreateCommand will return CommandTarget depending on targettype
func CreateCommand(targettype string) (CommandTarget, error) {
	switch targettype {
	case "obs":
		return new(obsCommandTarget), nil
	}
	return nil, fmt.Errorf("%v is not a valid command-target type", targettype)
}
