# Building from source

This application uses [go OBS websocket](https://github.com/christopher-dG/go-obs-websocket) library by Chris De Graaf.  
Unfortunately the current version of the library supports v4.7 of the websocket protocol and has some issues with scenes collections.  
Code is generated automatically and I fixed the issue, basing my work on [this pull request](https://github.com/christopher-dG/go-obs-websocket/pull/8) that hasn't been merged yet for known issue.  
Since those issue don't impact the features I need, I created my own fork.  
To build your own version you need to clone [this branch](https://github.com/VMinute/go-obs-websocket/tree/fix-scene-collections), build it locally and add its path as replace inside [go.mod](go.mod).

I also sent a pull request to the main OBS-websocket to generate a protocol definition file that makes management of scene collections easier. [You can find it here](https://github.com/Palakis/obs-websocket/pull/641)

Once you downloaded the modified obs websocket library, change the path inside [go.mod](go.mod) to point to your local folder and just run:

```
go build
```

to generate your executable.

The application should be executed providing a valid configuration file as command line parameter.

## Code structure

[keypad.go](keypad.go) contains only the main function, all the work is demanded to a keypad-controller object defined in [controller/keypad-controller.go](controller/keypad-controller.go).  
This object will load and check configuration. During this phase keypads, targets and keybindigns are instantiated.  
Then the object will just wait for key events and execute the corresponding commands.
The interface for keypad objects is defined in [keypads/keypad.go](keypads/keypad.go). The serial keypad is implemented in [keypads/serial.go](keypad/serial.go).
Targets interface is defined in [target/commandtarget.go](target/commandtarget.go). Since most of them will require the same basic function to check if a command is valid end execute it in [target/commandsmap.go](target/commandsmap.go) you'll find a useful implementation of a map with command names and check and execute functions.  
OBS commands are implemented in [target/obs.go](target/obs.go).

To add a new keypad type add its definition inside the keypads package and the code required to create it to the *CreateKeypad* func inside [keypads/keypad.go](keypads/keypad.go).

To add a new command target add it's implementation inside the targets package and code required to create an instance inside the *CreateCommand* func of [targets/commandtarget.go](targets/commandtarget.go).

