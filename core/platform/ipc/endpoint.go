package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Endpoint struct {
	Address string `json:"address"`
	Token   string `json:"token"`
	PID     int    `json:"pid"`
}

func ReadEndpoint(statePath string) (Endpoint, error) {
	var endpoint Endpoint
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

func PruneStaleState(statePath string) error {
	endpoint, err := ReadEndpoint(statePath)
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

func writeEndpoint(statePath string, endpoint Endpoint) error {
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

func WriteEndpoint(statePath string, endpoint Endpoint) error {
	return writeEndpoint(statePath, endpoint)
}

func WriteEndpointForTesting(statePath string, endpoint Endpoint) error {
	return WriteEndpoint(statePath, endpoint)
}
