package main

import (
	"path/filepath"

	"github.com/shinerio/skillflow/core/config"
	daemonipc "github.com/shinerio/skillflow/core/platform/ipc"
)

const (
	controlCommandShowUI = "show-ui"
	controlCommandShow   = "show"
	controlCommandHide   = "hide"
	controlCommandQuit   = "quit"
)

type controlEndpoint = daemonipc.Endpoint
type controlRequest = daemonipc.Request
type controlResponse = daemonipc.Response
type loopbackControlServer = daemonipc.Server

func runtimeStateDir() string {
	return filepath.Join(config.AppDataDir(), "runtime")
}

func helperControlPath() string {
	return filepath.Join(runtimeStateDir(), "helper-control.json")
}

func uiControlPath() string {
	return filepath.Join(runtimeStateDir(), "ui-control.json")
}

func daemonServicePath() string {
	return filepath.Join(runtimeStateDir(), "daemon-service.json")
}

func startLoopbackControlServer(statePath string, handler func(command string) error) (*loopbackControlServer, error) {
	return daemonipc.StartLoopbackServer(statePath, handler)
}

func sendLoopbackControlCommand(statePath, command string) error {
	return daemonipc.SendLoopbackCommand(statePath, command)
}

func readControlEndpoint(statePath string) (controlEndpoint, error) {
	return daemonipc.ReadEndpoint(statePath)
}

func pruneStaleLoopbackControlState(statePath string) error {
	return daemonipc.PruneStaleState(statePath)
}

func writeControlEndpoint(statePath string, endpoint controlEndpoint) error {
	return daemonipc.WriteEndpointForTesting(statePath, endpoint)
}
