package ipc

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Server struct {
	statePath string
	token     string
	listener  net.Listener
	closeOnce sync.Once
	done      chan struct{}
}

func StartLoopbackServer(statePath string, handler func(command string) error) (*Server, error) {
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return nil, err
	}
	if err := PruneStaleState(statePath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	token, err := randomToken()
	if err != nil {
		_ = listener.Close()
		return nil, err
	}

	server := &Server{
		statePath: statePath,
		token:     token,
		listener:  listener,
		done:      make(chan struct{}),
	}

	if err := writeEndpoint(statePath, Endpoint{
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

func (s *Server) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.listener.Close()
		_ = os.Remove(s.statePath)
		<-s.done
	})
	return err
}

func (s *Server) serve(handler func(command string) error) {
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

func (s *Server) handleConn(conn net.Conn, handler func(command string) error) {
	defer conn.Close()

	var req Request
	if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil {
		_ = json.NewEncoder(conn).Encode(Response{OK: false, Error: err.Error()})
		return
	}
	if strings.TrimSpace(req.Token) != strings.TrimSpace(s.token) {
		_ = json.NewEncoder(conn).Encode(Response{OK: false, Error: "unauthorized"})
		return
	}
	if err := handler(req.Command); err != nil {
		_ = json.NewEncoder(conn).Encode(Response{OK: false, Error: err.Error()})
		return
	}
	_ = json.NewEncoder(conn).Encode(Response{OK: true})
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
