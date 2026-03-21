package settingsstore

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

func (s *Store) LoadWindowState() (WindowState, bool, error) {
	var state WindowState
	exists, err := s.ReadLocalSection("window", &state)
	if err != nil || !exists {
		return WindowState{}, false, err
	}
	normalized := NormalizeWindowState(state)
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
	return s.WriteLocalSection("window", normalized)
}
