package targets

import (
	"fmt"
	"log"
	"strings"
	"time"

	obsws "github.com/christopher-dG/go-obs-websocket"
	"gopkg.in/yaml.v3"
)

type obsCommand struct {
	Command    string
	Parameters []interface{}
}

type obsCommandTarget struct {
	client           obsws.Client
	quitflag         bool
	sceneCollections []string
	activeCollection string
	scenes           []string
	activeScene      string
	commands         chan obsCommand
	errors           chan error
	commandsMap      *Map
	streaming        bool
	recording        bool
	recordingPaused  bool
}

type obsCommandTargetConfig struct {
	Host     string
	Port     int16
	Password string
}

var obsCommands = map[string]CommandDefinition{
	"activatescene": {
		CheckFunc:   activateSceneCheck,
		ExecuteFunc: activateSceneExec},
	"prevscene": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: prevSceneExec},
	"nextscene": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: nextSceneExec},
	"activatescenecollection": {
		CheckFunc:   activateSceneCheck,
		ExecuteFunc: activateSceneCollectionExec},
	"prevscenecollection": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: prevSceneCollectionExec},
	"nextscenecollection": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: nextSceneCollectionExec},
	"startrecording": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: startRecordingExec},
	"stoprecording": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: stopRecordingExec},
	"togglerecording": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: toggleRecordingExec},
	"pauserecording": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: pauseRecordingExec},
	"resumerecording": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: resumeRecordingExec},
	"togglepauserecording": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: togglePauseRecordingExec},
	"startstreaming": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: startStreamingExec},
	"stopstreaming": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: stopStreamingExec},
	"togglestreaming": {
		CheckFunc:   NoParmsCheck,
		ExecuteFunc: toggleStreamingExec},
}

func activateSceneCheck(target interface{}, parameters []interface{}) error {

	if len(parameters) != 1 {
		return fmt.Errorf("Invalid parameters count for activateScene command")
	}

	if _, ok := parameters[0].(string); !ok {
		return fmt.Errorf("Invalid parameter type for activateScene command")
	}

	return nil
}

func activateSceneExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)

	scenename := parameters[0].(string)

	if obs.getSceneIndex(scenename) == -1 {
		if !strings.Contains(scenename, ".") {
			return fmt.Errorf("Invalid scene name %s", scenename)
		}

		nameparts := strings.SplitN(scenename, ".", 2)

		collectionname := nameparts[0]
		scenename = nameparts[1]

		err := obs.activateSceneCollection(collectionname)

		if err != nil {
			return err
		}
	}

	return obs.activateScene(scenename)
}

func prevSceneExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	index := obs.getSceneIndex(obs.activeScene)
	index = index - 1
	if index < 0 {
		index = len(obs.scenes) - 1
	}
	return obs.activateScene(obs.scenes[index])
}

func nextSceneExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	index := obs.getSceneIndex(obs.activeScene)
	index = index + 1
	if index >= len(obs.scenes) {
		index = 0
	}
	return obs.activateScene(obs.scenes[index])
}

func activateSceneCollectionExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	scenecollectionname := parameters[0].(string)

	return obs.activateSceneCollection(scenecollectionname)
}

func prevSceneCollectionExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	index := obs.getSceneCollectionIndex(obs.activeCollection)
	index = index - 1
	if index < 0 {
		index = len(obs.sceneCollections) - 1
	}
	return obs.activateSceneCollection(obs.sceneCollections[index])
}

func nextSceneCollectionExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	index := obs.getSceneCollectionIndex(obs.activeCollection)
	index = index + 1
	if index >= len(obs.sceneCollections) {
		index = 0
	}
	return obs.activateSceneCollection(obs.sceneCollections[index])
}

func startRecordingExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	return obs.startRecording()
}

func stopRecordingExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	return obs.stopRecording()
}

func toggleRecordingExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	if obs.recording {
		return obs.stopRecording()
	}
	return obs.startRecording()
}

func pauseRecordingExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	return obs.pauseRecording()
}

func resumeRecordingExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	return obs.resumeRecording()
}

func togglePauseRecordingExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	if obs.recordingPaused {
		return obs.resumeRecording()
	}
	return obs.pauseRecording()
}

func startStreamingExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	return obs.startStreaming()
}

func stopStreamingExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	return obs.stopStreaming()
}

func toggleStreamingExec(target interface{}, parameters []interface{}) error {
	obs := target.(*obsCommandTarget)
	if obs.streaming {
		return obs.stopStreaming()
	}
	return obs.startStreaming()
}

func (obs *obsCommandTarget) CheckCommand(command string, parameters []interface{}) error {
	return obs.commandsMap.CheckCommand(command, parameters)
}

func (obs *obsCommandTarget) ExecuteCommand(command string, parameters []interface{}) error {
	if !obs.client.Connected() {
		return fmt.Errorf("OBS winsock connection is not active")
	}

	var cmd = obsCommand{
		Command:    command,
		Parameters: parameters,
	}

	obs.commands <- cmd
	return <-obs.errors
}

func (obs *obsCommandTarget) Init(configyaml []byte) error {

	cfg := obsCommandTargetConfig{
		Port:     4444,
		Host:     "localhost",
		Password: "",
	}

	err := yaml.Unmarshal(configyaml, &cfg)

	if err != nil {
		log.Printf("error %v parsing obs target configuration", err)
		return err
	}

	obs.commandsMap = new(Map)

	obs.commandsMap.Init(obs, obsCommands)

	obs.client.Host = cfg.Host
	obs.client.Port = int(cfg.Port)
	if cfg.Password != "" {
		obs.client.Password = cfg.Password
	}

	obs.quitflag = false
	obs.commands = make(chan obsCommand)
	obs.errors = make(chan error)

	obs.recording = false
	obs.streaming = false
	obs.recordingPaused = false

	go obs.manageWebSockCommunication()

	return nil
}

func (obs *obsCommandTarget) onSwitchScenes(e obsws.Event) {
	se := e.(obsws.SwitchScenesEvent)

	obs.activeScene = se.SceneName
}

func (obs *obsCommandTarget) onScenesChanged(e obsws.Event) {
	se := e.(obsws.ScenesChangedEvent)

	obs.scenes = make([]string, len(se.Scenes))

	for index, s := range se.Scenes {
		obs.scenes[index] = s.Name
	}
}

func (obs *obsCommandTarget) onScenesCollectionChanged(e obsws.Event) {
	se := e.(obsws.SceneCollectionChangedEvent)

	obs.activeCollection = se.SceneCollection
	obs.refreshScenes()
}

func (obs *obsCommandTarget) onSceneCollectionListChanged(_ obsws.Event) {
	obs.refreshSceneCollections()
}

func (obs *obsCommandTarget) onRecordingStarting(_ obsws.Event) {
	obs.recording = true
	obs.recordingPaused = false
}

func (obs *obsCommandTarget) onRecordingStopping(_ obsws.Event) {
	obs.recording = false
	obs.recordingPaused = false
}

func (obs *obsCommandTarget) onRecordingPaused(_ obsws.Event) {
	obs.recordingPaused = true
}

func (obs *obsCommandTarget) onRecordingResumed(_ obsws.Event) {
	obs.recordingPaused = false
}

func (obs *obsCommandTarget) onStreamingStarting(_ obsws.Event) {
	obs.streaming = true
}

func (obs *obsCommandTarget) onStreamingStopping(_ obsws.Event) {
	obs.streaming = false
}

func (obs *obsCommandTarget) createEventHandlers() {
	obs.client.AddEventHandler("SwitchScenes", obs.onSwitchScenes)
	obs.client.AddEventHandler("ScenesChanged", obs.onScenesChanged)
	obs.client.AddEventHandler("SceneCollectionChanged", obs.onScenesCollectionChanged)
	obs.client.AddEventHandler("SceneCollectionListChanged", obs.onSceneCollectionListChanged)
	obs.client.AddEventHandler("RecordingStarting", obs.onRecordingStarting)
	obs.client.AddEventHandler("RecordingStopping", obs.onRecordingStopping)
	obs.client.AddEventHandler("RecordingPaused", obs.onRecordingPaused)
	obs.client.AddEventHandler("RecordingResumed", obs.onRecordingResumed)
	obs.client.AddEventHandler("StreamStarting", obs.onStreamingStarting)
	obs.client.AddEventHandler("StreamStopping", obs.onStreamingStopping)
}

func (obs *obsCommandTarget) refreshScenes() error {
	obs.activeScene = ""

	slreq := obsws.NewGetSceneListRequest()
	slresp, err := slreq.SendReceive(obs.client)

	if err != nil {
		return err
	}

	obs.scenes = make([]string, len(slresp.Scenes))

	obs.activeScene = slresp.CurrentScene

	for index, scene := range slresp.Scenes {
		obs.scenes[index] = scene.Name
	}

	return nil
}

func (obs *obsCommandTarget) refreshSceneCollections() error {

	obs.activeCollection = ""
	obs.activeScene = ""

	lscreq := obsws.NewListSceneCollectionsRequest()

	lscresp, err := lscreq.SendReceive(obs.client)

	if err != nil {
		return err
	}

	obs.sceneCollections = make([]string, len(lscresp.SceneCollections))

	for index, collection := range lscresp.SceneCollections {
		obs.sceneCollections[index] = collection.Name
	}

	gcsreq := obsws.NewGetCurrentSceneCollectionRequest()

	gcsresp, err := gcsreq.SendReceive(obs.client)

	if err != nil {
		return err
	}

	obs.activeCollection = gcsresp.ScName

	return obs.refreshScenes()
}

func (obs *obsCommandTarget) refreshOBSState() error {

	ssreq := obsws.NewGetStreamingStatusRequest()

	ssresp, err := ssreq.SendReceive(obs.client)

	if err != nil {
		return err
	}

	obs.streaming = ssresp.Streaming
	obs.recording = ssresp.Recording
	return nil
}

func (obs *obsCommandTarget) processCommand(command obsCommand) error {
	return obs.commandsMap.ExecuteCommand(command.Command, command.Parameters)
}

func (obs *obsCommandTarget) pingObs() bool {
	gvreq := obsws.NewGetVersionRequest()

	_, err := gvreq.SendReceive(obs.client)

	if err != nil {
		return false
	}
	return true
}

func (obs *obsCommandTarget) manageWebSockCommunication() {
	for !obs.quitflag {
		if obs.client.Connected() {
			obs.client.Disconnect()
		}

		err := obs.client.Connect()

		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		err = obs.refreshSceneCollections()
		if err != nil {
			log.Print(err)
			continue
		}

		err = obs.refreshOBSState()

		if err != nil {
			log.Print(err)
			continue
		}

		obs.createEventHandlers()

		loop := true

		for loop {
			select {
			case commandMsg := <-obs.commands:
				obs.errors <- obs.processCommand(commandMsg)
			case <-time.After(time.Second * 5):
				if !obs.pingObs() {
					loop = false
				}
			}
		}

		if obs.client.Connected() {
			obs.client.Disconnect()
		}
	}

	if obs.client.Connected() {
		obs.client.Disconnect()
	}
}

func (obs *obsCommandTarget) getSceneIndex(sceneName string) int {
	for index, s := range obs.scenes {
		if s == sceneName {
			return index
		}
	}
	return -1
}

func (obs *obsCommandTarget) getSceneCollectionIndex(sceneCollectionName string) int {
	for index, s := range obs.sceneCollections {
		if s == sceneCollectionName {
			return index
		}
	}
	return -1
}

func (obs *obsCommandTarget) activateScene(scenename string) error {
	scsreq := obsws.NewSetCurrentSceneRequest(scenename)

	_, err := scsreq.SendReceive(obs.client)

	if err != nil {
		return err
	}

	obs.activeScene = scenename
	return nil
}

func (obs *obsCommandTarget) activateSceneCollection(scenecollectionname string) error {
	if obs.getSceneCollectionIndex(scenecollectionname) == -1 {
		return fmt.Errorf("Invalid collection scene name")
	}

	scscreq := obsws.NewSetCurrentSceneCollectionRequest(scenecollectionname)

	_, err := scscreq.SendReceive(obs.client)

	if err != nil {
		return err
	}

	obs.activeCollection = scenecollectionname
	return err
}

func (obs *obsCommandTarget) startRecording() error {

	if obs.recording {
		return nil
	}

	srreq := obsws.NewStartRecordingRequest()

	_, err := srreq.SendReceive(obs.client)

	return err
}

func (obs *obsCommandTarget) stopRecording() error {

	if !obs.recording {
		return nil
	}

	srreq := obsws.NewStopRecordingRequest()

	_, err := srreq.SendReceive(obs.client)

	return err
}

func (obs *obsCommandTarget) pauseRecording() error {

	if !obs.recording || obs.recordingPaused {
		return nil
	}

	prreq := obsws.NewPauseRecordingRequest()

	_, err := prreq.SendReceive(obs.client)

	return err
}

func (obs *obsCommandTarget) resumeRecording() error {

	if !obs.recording || !obs.recordingPaused {
		return nil
	}

	rrreq := obsws.NewResumeRecordingRequest()

	_, err := rrreq.SendReceive(obs.client)

	return err
}

func (obs *obsCommandTarget) startStreaming() error {

	if obs.streaming {
		return nil
	}

	ssreq := obsws.NewStartStreamingRequest(nil, "", nil, nil, "", "", false, "", "")

	_, err := ssreq.SendReceive(obs.client)

	return err
}

func (obs *obsCommandTarget) stopStreaming() error {

	if !obs.streaming {
		return nil
	}

	ssreq := obsws.NewStopStreamingRequest()

	_, err := ssreq.SendReceive(obs.client)

	return err
}
