package keypad

import (
	"fmt"
)

// Event is used to report a key event, key must be translated in a valid string
type Event struct {
	Source string
	Key    string
}

// Keypad is th base interface for all the keypads
type Keypad interface {
	Init(name string, configyaml []byte) error // reads configuration and checks if HW is available
	Start(keypresses chan<- Event) error       // starts sending keypad events
	Close()                                    // gracefully terminates the channel
	GetName() string
}

// CreateKeypad creates a keypad instance based on type string
func CreateKeypad(keypadtype string) (Keypad, error) {
	switch keypadtype {
	case "serial":
		return new(serialKeypad), nil
	}
	return nil, fmt.Errorf("%v is not a valid keypad type", keypadtype)
}
