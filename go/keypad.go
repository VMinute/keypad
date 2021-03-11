package main

import (
	"flag"
	"keypad/controller"
	"log"
    "os"
)

func main() {
	var configname = os.Getenv("HOME") + "/.keypad.yaml"

	flag.Parse()

	if len(flag.Args()) > 0 {
		configname = flag.Args()[0]
	}

	keypadcontroller, err := controller.CreateAndInitController(configname)

	if err != nil {
		log.Fatal(err)
	}

	err = keypadcontroller.StartProcessing()

	if err != nil {
		log.Fatal(err)
	}
}
