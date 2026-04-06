package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
)

func SendLoopbackCommand(statePath, command string) error {
	endpoint, err := ReadEndpoint(statePath)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", endpoint.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(Request{
		Token:   endpoint.Token,
		Command: command,
	}); err != nil {
		return err
	}

	var resp Response
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
