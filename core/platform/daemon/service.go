package daemon

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"strings"
	"sync"

	"github.com/shinerio/skillflow/core/platform/eventbus"
	daemonipc "github.com/shinerio/skillflow/core/platform/ipc"
)

type ServiceHandler func(ctx context.Context, params json.RawMessage) (any, error)

type Service struct {
	statePath string
	token     string
	listener  net.Listener
	server    *http.Server
	closeOnce sync.Once
	done      chan struct{}
	eventHub  atomic.Pointer[eventbus.Hub]
}

type ServiceRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type ServiceResponse struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

func StartService(statePath string, handlers map[string]ServiceHandler) (*Service, error) {
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return nil, err
	}
	if err := daemonipc.PruneStaleState(statePath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	token, err := randomServiceToken()
	if err != nil {
		_ = listener.Close()
		return nil, err
	}

	svc := &Service{
		statePath: statePath,
		token:     token,
		listener:  listener,
		done:      make(chan struct{}),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/invoke", svc.handleInvoke(handlers))
	mux.HandleFunc("/events", svc.handleEvents())
	svc.server = &http.Server{Handler: mux}

	if err := daemonipc.WriteEndpoint(statePath, daemonipc.Endpoint{
		Address: listener.Addr().String(),
		Token:   token,
		PID:     os.Getpid(),
	}); err != nil {
		_ = listener.Close()
		return nil, err
	}

	go func() {
		defer close(svc.done)
		_ = svc.server.Serve(listener)
	}()
	return svc, nil
}

func (s *Service) SetEventHub(hub *eventbus.Hub) {
	s.eventHub.Store(hub)
}

func (s *Service) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.server.Close()
		_ = os.Remove(s.statePath)
		<-s.done
	})
	return err
}

func (s *Service) handleInvoke(handlers map[string]ServiceHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeServiceCORSHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		if strings.TrimSpace(r.Header.Get("X-SkillFlow-Token")) != strings.TrimSpace(s.token) {
			writeServiceResponse(w, http.StatusUnauthorized, ServiceResponse{OK: false, Error: "unauthorized"})
			return
		}

		var req ServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeServiceResponse(w, http.StatusBadRequest, ServiceResponse{OK: false, Error: err.Error()})
			return
		}
		handler, ok := handlers[strings.TrimSpace(req.Method)]
		if !ok {
			writeServiceResponse(w, http.StatusNotFound, ServiceResponse{OK: false, Error: "method not found"})
			return
		}

		result, err := handler(r.Context(), req.Params)
		if err != nil {
			writeServiceResponse(w, http.StatusOK, ServiceResponse{OK: false, Error: err.Error()})
			return
		}

		var payload json.RawMessage
		if result != nil {
			data, marshalErr := json.Marshal(result)
			if marshalErr != nil {
				writeServiceResponse(w, http.StatusInternalServerError, ServiceResponse{OK: false, Error: marshalErr.Error()})
				return
			}
			payload = data
		}
		writeServiceResponse(w, http.StatusOK, ServiceResponse{OK: true, Result: payload})
	}
}

func (s *Service) handleEvents() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeServiceCORSHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		if strings.TrimSpace(r.Header.Get("X-SkillFlow-Token")) != strings.TrimSpace(s.token) {
			writeServiceResponse(w, http.StatusUnauthorized, ServiceResponse{OK: false, Error: "unauthorized"})
			return
		}

		hub := s.eventHub.Load()
		if hub == nil {
			writeServiceResponse(w, http.StatusServiceUnavailable, ServiceResponse{OK: false, Error: "event stream unavailable"})
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			writeServiceResponse(w, http.StatusInternalServerError, ServiceResponse{OK: false, Error: "streaming unsupported"})
			return
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		ch := hub.Subscribe()
		defer hub.Unsubscribe(ch)

		encoder := json.NewEncoder(w)
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					return
				}
				if err := encoder.Encode(evt); err != nil {
					return
				}
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}
}

func InvokeService(statePath, method string, params any, result any) error {
	endpoint, err := daemonipc.ReadEndpoint(statePath)
	if err != nil {
		return err
	}

	reqBody := ServiceRequest{Method: method}
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return err
		}
		reqBody.Params = data
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://"+endpoint.Address+"/invoke", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SkillFlow-Token", endpoint.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var serviceResp ServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&serviceResp); err != nil {
		return err
	}
	if !serviceResp.OK {
		if strings.TrimSpace(serviceResp.Error) == "" {
			return fmt.Errorf("service call failed")
		}
		return errors.New(serviceResp.Error)
	}
	if result != nil && len(serviceResp.Result) > 0 {
		if err := json.Unmarshal(serviceResp.Result, result); err != nil {
			return err
		}
	}
	return nil
}

func StreamEvents(statePath string, ctx context.Context, handle func(eventbus.Event)) error {
	endpoint, err := daemonipc.ReadEndpoint(statePath)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+endpoint.Address+"/events", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-SkillFlow-Token", endpoint.Token)

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var serviceResp ServiceResponse
		if err := json.NewDecoder(resp.Body).Decode(&serviceResp); err == nil && strings.TrimSpace(serviceResp.Error) != "" {
			return errors.New(serviceResp.Error)
		}
		return fmt.Errorf("event stream status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var evt eventbus.Event
		if err := json.Unmarshal(scanner.Bytes(), &evt); err != nil {
			return err
		}
		handle(evt)
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return ctx.Err()
}

func writeServiceResponse(w http.ResponseWriter, statusCode int, resp ServiceResponse) {
	writeServiceCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeServiceCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-SkillFlow-Token")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func randomServiceToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
