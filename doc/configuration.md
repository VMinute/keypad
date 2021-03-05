# Keypad configuration

The application does not provide a fancy GUI for its configuration, but writing a configuration file is not complicated (mostly if you use an editor that check YAML syntax like Visual Studio Code) and once you configured your setup you would probably not need to change it too often.
Configuration is made by tree main parts:
- **keypads**, defining the input devices
- **targets**, defining the controlled applications
- **keybindings** matching key presses on the input device to actions on the controlled applications
Currently only one type of keypad and two targets are supported, but the application is designed to support multiple input methods and control of different applications.

## Keypads

The **keypads** section contains an array of keypad objects.  
Each keypad object has the following attributes:
| Name           | Type              | Description                                                                                                 |
|----------------|-------------------|-------------------------------------------------------------------------------------------------------------|
| **keypadtype** | string            | type of the keypad, currently the only supported type is *serial*                                           |
| **name**       | string (optional) | Keypad name, if not specified it will use keypadtype. It's useful if you plan to use multiple keypads       |
| **config**     | object            | this is used to specify configuration of a specific keypad, check next section for type-specific parameters |

This is an example of keypad configuration (serial device).

```YAML
keypads:
  - keypadtype: serial
    config:
      port: /dev/serial/by-path/pci-0000:00:14.0-usb-0:4.4.1.1.4:1.0-port0
      baudrate: 9600
```

### Serial Keypad

| Name         | Type   | Description                                                        |
|--------------|--------|--------------------------------------------------------------------|
| **port**     | string | Serial port used to control the device (OS specific)               |
| **baudrate** | number | Baud rate (default is 9600)                                        |
| **parity**   | string | Can be N = None, O = Odd, E = Even (default is N)                  |
| **stopbits** | number | 1 = 1 stop bit, 15 = 1.5 stop bits, 2 = 2 stop bits (default is 1) |
| **size**     | number | Number of bytes per serial byte, default is 8                      |

Port name can be tricky to configure because USB to serial devices are sometimes renamed by the OS on reboot.  
On some Arduino Nano versions/clones the USB to serial device used does not provide an unique ID, so on Linux you have to rely on his placement on the USB bus.  
Usually USB to serial devices are named /dev/ttyUSB* with progressive numbers, but those numbers may change after a reboot. Instead of using those entries you may check the entry that is created under:

```
/dev/serial/by-path/
```

This will provide a fixed name generated from the USB host controller and ports/hubs sequence used to reach your device. This won't be changed as long as you don't plug your keypad in a different USB port.  
At the end your port value will be something like:

```
/dev/serial/by-path/pci-0000:00:14.0-usb-0:4.4.1.1.4:1.0-port0
```

On windows you can use the COM*: device name (ex: *COM5:*) and you can configure a fixed ID for your devices via device manager, as described [here](https://crazyforelectonics.wordpress.com/2016/08/21/changing-com-port-number-of-usb-driver/).

## Targets

Targets are the applications/features that can be controlled by the keypads.  
Currently we support only OBS as a target, plus an "internal" target that allows you to change keybindings (letting a single keypad operate in multiple "modes").  
Each target has some specific configuration parameters and provides a series of commands that could be connected to key presses using the keybindings that are described in the next chapter.

| Name           | Type              | Description                                                                                                                                                                   |
|----------------|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **targettype** | string            | type of the target, currently the only supported type is *obs*                                                                                                                |
| **name**       | string (optional) | Target name, if not specified it will use keypadtype. It's useful if you plan to use control different instances of the same application (ex: OBS instances on different PCs) |
| **config**     | object            | this is used to specify configuration of a specific target, check next section for type-specific parameters                                                                   |

### OBS

This target can be used to control [Open Broadcaster Software](https://obsproject.com/) using the [OBS websocket plugin](https://github.com/Palakis/obs-websocket)

#### Configuration

| Name         | Type   | Description                                                                                                                  |
|--------------|--------|------------------------------------------------------------------------------------------------------------------------------|
| **Host**     | string | Hostname/IP used to connect to the [OBS websocket plugin](https://github.com/Palakis/obs-websocket) (default is *localhost*) |
| **Port**     | number | Port where the [OBS websocket plugin](https://github.com/Palakis/obs-websocket) accepts connections (default is *4444*)      |
| **Password** | string | Password used to authenticate on the [OBS websocket plugin](https://github.com/Palakis/obs-websocket)                        |

This is an example of configuration for an OBS target: 

```YAML
targets:
  - targettype: obs
    config:
      host: localhost
      port: 4444
      password: yourobswebsocketpassword
```

#### Commands

The OBS target supports command to change scene and scene collections and control recording and streaming.
Commands are case-insensitive.  
Parameters as passed as an array.

| Command                     | Parameters    | Description                                                                                                                    |
|-----------------------------|---------------|--------------------------------------------------------------------------------------------------------------------------------|
| **activateScene**           | name (string) | activate a specific scene. If the name is in the format <collection>.<scene> then the scene collection will be activated first |
| **prevScene**               | none          | Moves to the previous scene (in the order they are defined in OBS)                                                             |
| **nextScene**               | none          | Moves to the next scene (in the order they are defined in OBS)                                                                 |
| **activateSceneCollection** | name (string) | activates the specified scene collection                                                                                       |
| **prevSceneCollection**     | none          | Moves to the previous scene collection                                                                                         |
| **nextSceneCollection**     | none          | Moves to the next scene collection                                                                                             |
| **startRecording**          | none          | Starts recording                                                                                                               |
| **stopRecording**           | none          | Stops recording                                                                                                                |
| **toggleRecording**         | none          | Start/Stop recording, depending on current state                                                                               |
| **pauseRecording**          | none          | pause current recording                                                                                                        |
| **resumeRecording**         | none          | resumes current recording                                                                                                      |
| **togglePauseRecording**    | none          | pause/resume recording depending on current state                                                                              |
| **startStreaming**          | none          | Starts streaming                                                                                                               |
| **stopStreaming**           | none          | Stops streaming                                                                                                                |
| **toggleStreaming**         | none          | Start/Stop streaming, depending on current state                                                                               |

### Keyboard

This target can be used to emulate keystrokes, this will let you control application that don't provide an API interface. For example you can map the keystrokes required to move to the next slide in your presentation software.

#### Configuration

currently no configuration is required for the keyboard

#### Commands

| Command      | Parameters                       | Description                                                                                                                                                                                                    |
|--------------|----------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **keypress** | keys (string or array of string) | emulates a specific keypress. First parameter is the main key (see following table for a list) and it can be followed by one or more of the modifiers: ctrl, alt, shift, right-ctrl, right-shift, altgr, super |

Supported keys:

| Key                   | Description                                                                        |
|-----------------------|------------------------------------------------------------------------------------|
| 0..9                  | Numeric key                                                                        |
| A..Z                  | Letter key (to actually input a capital letter you have to add the shift modifier) |
| space                 | space                                                                              |
| backspace             | backspace                                                                          |
| up, down, left, right | arrow keys                                                                         |
| F1..F12               | Function keys                                                                      |
| enter                 | Enter, Return                                                                      |
| esc                   | Escape                                                                             |

### Bindings

This target does not need to be defined, it's always available and can be used to "remap" the keypad, activating a different set of bindings.  
It can be used to remap the keypad dynamically and can be used to support different "modes" in the same configuration (ex: recording and streaming).

#### Commands

| Command      | Parameters    | Description                                                                                         |
|--------------|---------------|-----------------------------------------------------------------------------------------------------|
| **activate** | name (string) | activate a specific set of key bindings.                                                            |
| **prev**     | none          | Moves to the previous set of key bindings (in the order they are defined in the configuration file) |
| **next**     | none          | Moves to the next set of key bindings (in the order they are defined in the configuration file)     |

## Key Bindings

Key bindings are used to connect a key (rapresented by a string) to one or more commands.  
It's possible to define multiple sets of key bindings and, using the bindings command target, remap the keypad dinamically at runtime. This will allow, for example, usage of two different sets of bindings for recording and streaming.

Each set of binding is an array of objects with the following attributes:

| Name         | Type             | Description                                                                                                                                                    |
|--------------|------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **keys**     | array of strings | Keys associated to this binding, all the keys will activate the same commands. Keys can be specified with just their value or in the *keypad_name.key* format. |
| **commands** | array of objects | Commands that will be executed when the keys are pushed. Commands are executed in order and failure executing one of them will stop the entire sequence        |

Each command is defined as:

| Name           | Type             | Description                                                                          |
|----------------|------------------|--------------------------------------------------------------------------------------|
| **command**    | string           | Command name in the format <target>.<command>                                        |
| **parameters** | array (optional) | Additional parameters as an array. Number and type of elements depend on the command |


Using the name string attribute you can specify a unique name for each set (useful only if you want to do remapping).
This is a sample definition of two sets of keybindings (one named recording, the other named streaming).

```YAML
keybindings:
  - name: recording
    bindings:
      - keys:
          - "1"
        commands:
          - command: obs.activateScene
            parameters:
              - "scene1"
      - keys:
          - "2"
        commands:
          - command: obs.activateScene
            parameters:
              - "scene2"
      - keys:
          - serial.A
        commands:
          - command: obs.prevScene
      - keys:
          - serial.B
        commands:
          - command: obs.nextScene
      - keys:
          - serial.C
        commands:
          - command: obs.prevSceneCollection
      - keys:
          - serial.D
        commands:
          - command: obs.nextSceneCollection
      - keys:
          - serial.*
        commands:
          - command: obs.toggleRecording
      - keys:
          - serial.#
        commands:
          - command: obs.togglePauseRecording
      - keys:
          - serial.0
        commands:
          - command: bindings.next
      - keys:
          - serial.7
        commands:
          - command: keyboard.keypress
            parameters:
              - space
      - keys:
          - serial.8
        commands:
          - command: keyboard.keypress
            parameters:
              - B
              - shift
  - name: streaming
    bindings:
          - serial.1
        commands:
          - command: obs.activateScene
            parameters:
              - "scene1"
      - keys:
          - serial.2
        commands:
          - command: obs.activateScene
            parameters:
              - "scene2"
      - keys:
          - serial.A
        commands:
          - command: obs.prevScene
      - keys:
          - serial.B
        commands:
          - command: obs.nextScene
      - keys:
          - serial.C
        commands:
          - command: obs.prevSceneCollection
      - keys:
          - serial.D
        commands:
          - command: obs.nextSceneCollection
      - keys:
          - serial.*
        commands:
          - command: obs.toggleStreaming
      - keys:
          - serial.0
        commands:
          - command: bindings.next
```


