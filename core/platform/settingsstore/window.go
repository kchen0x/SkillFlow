package settingsstore

import (
	"encoding/json"
	"os"
)

type WindowState struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

const (
	MinWindowWidth  = 640
	MinWindowHeight = 480
)

func NormalizeWindowState(state WindowState) WindowState {
	if state.Width < MinWindowWidth {
		state.Width = 0
	}
	if state.Height < MinWindowHeight {
		state.Height = 0
	}
	if state.Width == 0 || state.Height == 0 {
		return WindowState{}
	}
	return state
}

type localWindowStateDocument struct {
	Window *WindowState `json:"window,omitempty"`
}

func (s *Store) LoadWindowState() (WindowState, bool, error) {
	var doc localWindowStateDocument
	exists, err := s.ReadLocal(&doc)
	if err != nil || !exists || doc.Window == nil {
		return WindowState{}, false, err
	}
	normalized := NormalizeWindowState(*doc.Window)
	if normalized.Width == 0 || normalized.Height == 0 {
		return WindowState{}, false, nil
	}
	return normalized, true, nil
}

func (s *Store) SaveWindowState(state WindowState) error {
	normalized := NormalizeWindowState(state)
	if normalized.Width == 0 || normalized.Height == 0 {
		return nil
	}

	localDoc := map[string]json.RawMessage{}
	if data, err := os.ReadFile(s.LocalPath()); err == nil {
		if unmarshalErr := json.Unmarshal(data, &localDoc); unmarshalErr != nil {
			localDoc = map[string]json.RawMessage{}
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	windowData, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	localDoc["window"] = windowData
	return s.WriteLocal(localDoc)
}
