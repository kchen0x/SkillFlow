package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shinerio/skillflow/core/config"
)

const (
	controlCommandShowUI = "show-ui"
	controlCommandShow   = "show"
	controlCommandHide   = "hide"
	controlCommandQuit   = "quit"
)

type controlEndpoint struct {
	Address string `json:"address"`
	Token   string `json:"token"`
	PID     int    `json:"pid"`
}

type controlRequest struct {
	Token   string `json:"token"`
	Command string `json:"command"`
}

type controlResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type loopbackControlServer struct {
	statePath string
	token     string
	listener  net.Listener
	closeOnce sync.Once
	done      chan struct{}
}

func runtimeStateDir() string {
	return filepath.Join(config.AppDataDir(), "runtime")
}

func helperControlPath() string {
	return filepath.Join(runtimeStateDir(), "helper-control.json")
}

func uiControlPath() string {
	return filepath.Join(runtimeStateDir(), "ui-control.json")
}

func startLoopbackControlServer(statePath string, handler func(command string) error) (*loopbackControlServer, error) {
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return nil, err
	}
	if err := pruneStaleLoopbackControlState(statePath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	token, err := randomControlToken()
	if err != nil {
		_ = listener.Close()
		return nil, err
	}

	server := &loopbackControlServer{
		statePath: statePath,
		token:     token,
		listener:  listener,
		done:      make(chan struct{}),
	}

	if err := writeControlEndpoint(statePath, controlEndpoint{
		Address: listener.Addr().String(),
		Token:   token,
		PID:     os.Getpid(),
	}); err != nil {
		_ = listener.Close()
		return nil, err
	}

	go server.serve(handler)
	return server, nil
}

func (s *loopbackControlServer) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.listener.Close()
		_ = os.Remove(s.statePath)
		<-s.done
	})
	return err
}

func (s *loopbackControlServer) serve(handler func(command string) error) {
	defer close(s.done)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}
		go s.handleConn(conn, handler)
	}
}

func (s *loopbackControlServer) handleConn(conn net.Conn, handler func(command string) error) {
	defer conn.Close()

	var req controlRequest
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil {
		_ = json.NewEncoder(conn).Encode(controlResponse{OK: false, Error: err.Error()})
		return
	}
	if subtleTrim(req.Token) != subtleTrim(s.token) {
		_ = json.NewEncoder(conn).Encode(controlResponse{OK: false, Error: "unauthorized"})
		return
	}
	if err := handler(req.Command); err != nil {
		_ = json.NewEncoder(conn).Encode(controlResponse{OK: false, Error: err.Error()})
		return
	}
	_ = json.NewEncoder(conn).Encode(controlResponse{OK: true})
}

func sendLoopbackControlCommand(statePath, command string) error {
	endpoint, err := readControlEndpoint(statePath)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", endpoint.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(controlRequest{
		Token:   endpoint.Token,
		Command: command,
	}); err != nil {
		return err
	}

	var resp controlResponse
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&resp); err != nil {
		return err
	}
	if !resp.OK {
		if strings.TrimSpace(resp.Error) == "" {
			return fmt.Errorf("control command failed")
		}
		return errors.New(resp.Error)
	}
	return nil
}

func readControlEndpoint(statePath string) (controlEndpoint, error) {
	var endpoint controlEndpoint
	data, err := os.ReadFile(statePath)
	if err != nil {
		return endpoint, err
	}
	if err := json.Unmarshal(data, &endpoint); err != nil {
		return endpoint, err
	}
	if strings.TrimSpace(endpoint.Address) == "" || strings.TrimSpace(endpoint.Token) == "" {
		return endpoint, fmt.Errorf("invalid control endpoint")
	}
	return endpoint, nil
}

func pruneStaleLoopbackControlState(statePath string) error {
	endpoint, err := readControlEndpoint(statePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		removeErr := os.Remove(statePath)
		if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return removeErr
		}
		return nil
	}

	conn, err := net.DialTimeout("tcp", endpoint.Address, 200*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		return nil
	}

	removeErr := os.Remove(statePath)
	if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		return removeErr
	}
	return nil
}

func writeControlEndpoint(statePath string, endpoint controlEndpoint) error {
	data, err := json.Marshal(endpoint)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(statePath), filepath.Base(statePath)+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, statePath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func randomControlToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func subtleTrim(value string) string {
	return strings.TrimSpace(value)
}
