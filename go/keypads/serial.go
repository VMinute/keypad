package keypad

import (
	"fmt"
	"github.com/tarm/serial"
	"gopkg.in/yaml.v3"
	"log"
)

type serialKeypad struct {
	name string
	Port *serial.Port
}

type serialKeypadConfiguration struct {
	Port     string
	BaudRate int
	Parity   string
	StopBits byte
	Size     byte
}

func (s *serialKeypad) Init(name string, configyaml []byte) error {

	s.name = name

	cfg := serialKeypadConfiguration{
		Port:     "",
		BaudRate: 9600,
		Parity:   "N",
		StopBits: 1,
		Size:     8,
	}

	err := yaml.Unmarshal(configyaml, &cfg)

	if err != nil {
		log.Printf("error %v parsing serial driver configuration", err)
		return err
	}

	if cfg.Port == "" {
		return fmt.Errorf("no serial port name has been configured")
	}

	portconf := serial.Config{
		Name:     cfg.Port,
		Baud:     cfg.BaudRate,
		Parity:   serial.Parity([]byte(cfg.Parity)[0]),
		StopBits: serial.StopBits(cfg.StopBits),
		Size:     cfg.Size,
	}

	s.Port, err = serial.OpenPort(&portconf)

	if err != nil {
		log.Printf("error %v opening port", err)
		return err
	}

	return nil
}

func (s *serialKeypad) processKeys(keyevents chan<- Event) error {
	b := make([]byte, 1)

	var err error
	var _ int

	for _, err = s.Port.Read(b); err == nil; _, err = s.Port.Read(b) {
		keyevents <- Event{Source: s.name, Key: string(b[0])}
	}
	return err
}

func (s *serialKeypad) Start(keyevents chan<- Event) error {
	go s.processKeys(keyevents)
	return nil
}

func (s *serialKeypad) Close() {
	s.Port.Close()
}

func (s *serialKeypad) GetName() string {
	return s.name
}
