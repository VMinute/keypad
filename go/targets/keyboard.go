package targets

import (
	"fmt"
	"reflect"

	"github.com/micmonay/keybd_event"
)

type keybdCommandTarget struct {
	commandsMap *Map
	kb          keybd_event.KeyBonding
}

var keybdCommands = map[string]CommandDefinition{
	"keypress": {
		CheckFunc:   keySequenceCheck,
		ExecuteFunc: keySequenceExec},
}

var keyMap = map[string]int{
	"0":         keybd_event.VK_0,
	"1":         keybd_event.VK_1,
	"2":         keybd_event.VK_2,
	"3":         keybd_event.VK_3,
	"4":         keybd_event.VK_4,
	"5":         keybd_event.VK_5,
	"6":         keybd_event.VK_6,
	"7":         keybd_event.VK_7,
	"8":         keybd_event.VK_8,
	"9":         keybd_event.VK_9,
	"A":         keybd_event.VK_A,
	"B":         keybd_event.VK_B,
	"C":         keybd_event.VK_C,
	"D":         keybd_event.VK_D,
	"E":         keybd_event.VK_E,
	"F":         keybd_event.VK_F,
	"G":         keybd_event.VK_G,
	"H":         keybd_event.VK_H,
	"I":         keybd_event.VK_I,
	"J":         keybd_event.VK_J,
	"K":         keybd_event.VK_K,
	"L":         keybd_event.VK_L,
	"M":         keybd_event.VK_M,
	"N":         keybd_event.VK_N,
	"O":         keybd_event.VK_O,
	"P":         keybd_event.VK_P,
	"Q":         keybd_event.VK_Q,
	"R":         keybd_event.VK_R,
	"S":         keybd_event.VK_S,
	"T":         keybd_event.VK_T,
	"U":         keybd_event.VK_U,
	"V":         keybd_event.VK_V,
	"W":         keybd_event.VK_W,
	"X":         keybd_event.VK_X,
	"Y":         keybd_event.VK_Y,
	"Z":         keybd_event.VK_Z,
	"space":     keybd_event.VK_SPACE,
	"backspace": keybd_event.VK_BACKSPACE,
	"up":        keybd_event.VK_UP,
	"down":      keybd_event.VK_DOWN,
	"left":      keybd_event.VK_LEFT,
	"right":     keybd_event.VK_RIGHT,
	"enter":     keybd_event.VK_ENTER,
	"esc":       keybd_event.VK_ESC,
	"F1":        keybd_event.VK_F1,
	"F2":        keybd_event.VK_F2,
	"F3":        keybd_event.VK_F3,
	"F4":        keybd_event.VK_F4,
	"F5":        keybd_event.VK_F5,
	"F6":        keybd_event.VK_F6,
	"F7":        keybd_event.VK_F7,
	"F8":        keybd_event.VK_F8,
	"F9":        keybd_event.VK_F9,
	"F10":       keybd_event.VK_F10,
	"F11":       keybd_event.VK_F11,
	"F12":       keybd_event.VK_F12,
}

var modifiersMap = map[string]string{
	"ctrl":        "HasCTRL",
	"alt":         "HasALT",
	"shift":       "HasSHIFT",
	"right-ctrl":  "HasRCTRL",
	"right-shift": "HasRSHIFT",
	"altgr":       "HasALTGR",
	"super":       "HasSuper",
}

func keySequenceCheck(target interface{}, parameters []interface{}) error {

	if len(parameters) < 1 {
		return fmt.Errorf("Invalid parameters count for keypress command")
	}

	if _, ok := parameters[0].(string); !ok {
		return fmt.Errorf("Invalid parameter type for keypress command")
	}

	if _, ok := keyMap[parameters[0].(string)]; !ok {
		return fmt.Errorf("Invalid key value for keypress command")
	}

	for i := 1; i < len(parameters); i++ {
		if _, ok := parameters[i].(string); !ok {
			return fmt.Errorf("Invalid parameter type for keypress command")
		}

		if _, ok := modifiersMap[parameters[i].(string)]; !ok {
			return fmt.Errorf("Invalid key modifier keypress command")
		}
	}

	return nil
}

func keySequenceExec(target interface{}, parameters []interface{}) error {
	keybd := target.(*keybdCommandTarget)

	key := parameters[0].(string)
	keycode := keyMap[key]

	keybd.kb.Clear()

	parms := make([]reflect.Value, 1)
	parms[0] = reflect.ValueOf(true)

	kbvalue := reflect.ValueOf(&keybd.kb)

	for i := 1; i < len(parameters); i++ {
		methodname := modifiersMap[parameters[i].(string)]
		method := kbvalue.MethodByName(methodname)
		method.Call(parms)
	}

	keybd.kb.SetKeys(keycode)
	return keybd.kb.Launching()
}

func (keybd *keybdCommandTarget) Init(configyaml []byte) error {
	err := error(nil)

	keybd.commandsMap = new(Map)
	keybd.commandsMap.Init(keybd, keybdCommands)

	keybd.kb, err = keybd_event.NewKeyBonding()
	return err
}

func (keybd *keybdCommandTarget) CheckCommand(command string, parameters []interface{}) error {
	return keybd.commandsMap.CheckCommand(command, parameters)
}

func (keybd *keybdCommandTarget) ExecuteCommand(command string, parameters []interface{}) error {
	return keybd.commandsMap.ExecuteCommand(command, parameters)
}
