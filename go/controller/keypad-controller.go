package controller

import (
	"fmt"
	"io/ioutil"
	keypad "keypad/keypads"
	"keypad/targets"
	"log"
	"strings"

	"gopkg.in/yaml.v3"
)

type keybindingCommandItem struct {
	Command    string
	Parameters []interface{}
}

type keybindingItem struct {
	Keys     []string
	Commands []keybindingCommandItem
}

type keybindingDefinition struct {
	Name     string
	Bindings []keybindingItem
}

type keypadItem struct {
	Name       string
	KeypadType string
	Config     interface{}
}

type keypadConfiguration struct {
	Keypads     []keypadItem
	Targets     []commandtargetItem
	KeyBindings []keybindingDefinition
}

type commandtargetItem struct {
	Name       string
	TargetType string
	Config     interface{}
}

type keybindingRuntimeItem struct {
	Target     targets.CommandTarget
	Command    string
	Parameters []interface{}
}

// KeypadsController links keypad events and commands
type KeypadsController interface {
	StartProcessing() error
	Init(configyaml []byte) error                                  // reads configuration and checks if target is available
	CheckCommand(command string, parameters []interface{}) error   // validates a command
	ExecuteCommand(command string, parameters []interface{}) error //executes a command
}

type keypadsControllerData struct {
	keypads        map[string]keypad.Keypad                      // keypads that can trigger key events
	targets        map[string]targets.CommandTarget              // objects that can execute commands
	keybindings    map[string]map[string][]keybindingRuntimeItem // bindings between keys and commands
	bindingsOrder  []string                                      // used to cycle to next/prev binding
	activeBindings string                                        // currently active bindings
	keyevents      <-chan keypad.Event                           // channel used to receive key events
	commandsMap    *targets.Map                                  // used to behave like a target for internal commands
}

// CreateAndInitController reads configuration file and initializes all the objects
// inside the controller
func CreateAndInitController(configfile string) (KeypadsController, error) {

	yamlfile, err := ioutil.ReadFile(configfile)

	if err != nil {
		log.Printf("Error %v reading configuration from %s", err, configfile)
		return nil, err
	}

	var config keypadConfiguration

	err = yaml.Unmarshal(yamlfile, &config)

	if err != nil {
		log.Printf("Error %v parsing configuration from %s", err, configfile)
		return nil, err
	}

	controller := new(keypadsControllerData)
	controller.keypads = make(map[string]keypad.Keypad)
	controller.targets = make(map[string]targets.CommandTarget)
	controller.keybindings = make(map[string]map[string][]keybindingRuntimeItem)
	controller.activeBindings = ""
	controller.commandsMap = new(targets.Map)

	controller.commandsMap.Init(controller, commandsMap)

	for index, keypadcfg := range config.Keypads {
		keypad, err := keypad.CreateKeypad(keypadcfg.KeypadType)

		if err != nil {
			log.Printf("Error %v parsing configuration from %s", err, configfile)
			return nil, err
		}

		configyaml, err := yaml.Marshal(keypadcfg.Config)

		name := keypadcfg.Name

		if name == "" {
			name = keypadcfg.KeypadType
		}

		err = keypad.Init(name, configyaml)

		if err != nil {
			log.Printf("Error %v initializing keypad %d", err, index)
			return nil, err
		}

		controller.keypads[name] = keypad
	}

	for index, targetcfg := range config.Targets {
		target, err := targets.CreateCommand(targetcfg.TargetType)

		if err != nil {
			log.Printf("Error %v parsing configuration from %s", err, configfile)
			return nil, err
		}

		configyaml, err := yaml.Marshal(targetcfg.Config)

		err = target.Init(configyaml)

		if err != nil {
			log.Printf("Error %v initializing command target %d", err, index)
			return nil, err
		}

		name := targetcfg.Name

		if name == "" {
			name = targetcfg.TargetType
		}

		controller.targets[name] = target
	}

	controller.targets["bindings"] = controller
	controller.bindingsOrder = make([]string, len(config.KeyBindings))

	for index, keybindingdefinition := range config.KeyBindings {
		name := "default"

		if keybindingdefinition.Name != "" {
			name = keybindingdefinition.Name
		}

		bindingsmap := make(map[string][]keybindingRuntimeItem)

		for _, keybinding := range keybindingdefinition.Bindings {
			runtimecommands := make([]keybindingRuntimeItem, len(keybinding.Commands))

			for index, command := range keybinding.Commands {
				cmdparts := strings.SplitN(command.Command, ".", 2)

				target := controller.targets[cmdparts[0]]

				if target == nil {
					return nil, fmt.Errorf("Invalid command target %s", cmdparts[0])
				}

				err := target.CheckCommand(cmdparts[1], command.Parameters)

				if err != nil {
					return nil, err
				}

				runtimecommands[index].Target = target
				runtimecommands[index].Command = cmdparts[1]
				runtimecommands[index].Parameters = command.Parameters
			}

			for _, key := range keybinding.Keys {
				bindingsmap[key] = runtimecommands
			}
		}

		controller.bindingsOrder[index] = name
		controller.keybindings[name] = bindingsmap

		if controller.activeBindings == "" {
			controller.activeBindings = name
		}
	}

	if len(controller.keypads) == 0 || len(controller.keybindings) == 0 || len(controller.targets) == 0 {
		return nil, fmt.Errorf("You must configure at least one keypad, one target and one set of key bindings")
	}

	return controller, nil
}

func (kc *keypadsControllerData) StartProcessing() error {

	keyeventschannel := make(chan keypad.Event)

	defer close(keyeventschannel)

	kc.keyevents = keyeventschannel

	for _, kp := range kc.keypads {
		defer kp.Close()

		err := kp.Start(keyeventschannel)

		if err != nil {
			return err
		}
	}

	for true {
		select {
		case keypress := <-kc.keyevents:
			go kc.processKeypress(keypress.Source, keypress.Key)
		}
	}

	return nil
}

func (kc *keypadsControllerData) processKeypress(source string, keypress string) {

	items, ok := kc.keybindings[kc.activeBindings][source+"."+keypress]

	if !ok {
		items, ok = kc.keybindings[kc.activeBindings][keypress]
	}

	if !ok {
		log.Printf("Key %s.%s has no valid bindings", source, keypress)
		return
	}

	for _, item := range items {
		err := item.Target.ExecuteCommand(item.Command, item.Parameters)

		if err != nil {
			log.Printf("Error %v processing key bindings for %s.%s", err, source, keypress)
			return
		}
	}

	log.Printf("Key bindings for %s.%s correctly processed", source, keypress)
}

var commandsMap = map[string]targets.CommandDefinition{
	"activate": {
		CheckFunc:   activateBindingsCheck,
		ExecuteFunc: activateBindingsExec},
	"next": {
		CheckFunc:   targets.NoParmsCheck,
		ExecuteFunc: nextBindingsExec},
	"previous": {
		CheckFunc:   targets.NoParmsCheck,
		ExecuteFunc: prevBindingsExec},
}

func activateBindingsCheck(target interface{}, parameters []interface{}) error {
	kc := target.(*keypadsControllerData)

	if len(parameters) != 1 {
		return fmt.Errorf("Invalid number of parameters")
	}

	bindings := parameters[0].(string)

	if _, ok := kc.keybindings[bindings]; !ok {
		return fmt.Errorf("Invalid bindings name %s", bindings)
	}
	return nil
}

func activateBindingsExec(target interface{}, parameters []interface{}) error {
	kc := target.(*keypadsControllerData)

	if len(parameters) != 1 {
		return fmt.Errorf("Invalid number of parameters")
	}

	bindings := parameters[0].(string)

	kc.activateBindings(bindings)
	return nil
}

func nextBindingsExec(target interface{}, parameters []interface{}) error {
	kc := target.(*keypadsControllerData)

	index := kc.getBindingsPos(kc.activeBindings)

	index = index + 1

	if index >= len(kc.bindingsOrder) {
		index = 0
	}

	kc.activateBindings(kc.bindingsOrder[index])
	return nil
}

func prevBindingsExec(target interface{}, parameters []interface{}) error {
	kc := target.(*keypadsControllerData)

	index := kc.getBindingsPos(kc.activeBindings)

	index = index - 1

	if index < 0 {
		index = len(kc.bindingsOrder) - 1
	}

	kc.activateBindings(kc.bindingsOrder[index])
	return nil
}

func (kc *keypadsControllerData) Init(configyaml []byte) error {
	// Init does not need to be implemented
	return nil
}

func (kc *keypadsControllerData) CheckCommand(command string, parameters []interface{}) error {
	return kc.commandsMap.CheckCommand(command, parameters)
}

func (kc *keypadsControllerData) ExecuteCommand(command string, parameters []interface{}) error {
	return kc.commandsMap.ExecuteCommand(command, parameters)
}

func (kc *keypadsControllerData) getBindingsPos(bindings string) int {
	for index, b := range kc.bindingsOrder {
		if b == bindings {
			return index
		}
	}
	return -1
}

func (kc *keypadsControllerData) activateBindings(bindings string) {
	if kc.activeBindings != bindings {
		kc.activeBindings = bindings
		log.Printf("Binding %s activated", bindings)
	}
}
