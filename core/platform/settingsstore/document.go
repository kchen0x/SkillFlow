package settingsstore

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (s *Store) ReadSharedSection(section string, out any) (bool, error) {
	return s.readSection(s.sharedPath, section, out)
}

func (s *Store) ReadLocalSection(section string, out any) (bool, error) {
	return s.readSection(s.localPath, section, out)
}

func (s *Store) WriteSharedSection(section string, value any) error {
	return s.writeSection(s.sharedPath, section, value)
}

func (s *Store) WriteLocalSection(section string, value any) error {
	return s.writeSection(s.localPath, section, value)
}

func (s *Store) readSection(path, section string, out any) (bool, error) {
	section = strings.TrimSpace(section)
	if section == "" {
		return false, fmt.Errorf("section name is required")
	}
	doc, exists, err := readRawDocument(path)
	if err != nil || !exists {
		return false, err
	}
	raw, ok := doc[section]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) writeSection(path, section string, value any) error {
	section = strings.TrimSpace(section)
	if section == "" {
		return fmt.Errorf("section name is required")
	}

	doc, err := readRawDocumentForUpdate(path)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	doc[section] = raw

	if path == s.sharedPath {
		return s.WriteShared(doc)
	}
	return s.WriteLocal(doc)
}

func readRawDocument(path string) (map[string]json.RawMessage, bool, error) {
	var doc map[string]json.RawMessage
	exists, err := readJSON(path, &doc)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return map[string]json.RawMessage{}, false, nil
	}
	if doc == nil {
		doc = map[string]json.RawMessage{}
	}
	return doc, true, nil
}

func readRawDocumentForUpdate(path string) (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]json.RawMessage{}, nil
		}
		return nil, err
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		// Be permissive on write operations so callers can repair malformed local files.
		return map[string]json.RawMessage{}, nil
	}
	if doc == nil {
		doc = map[string]json.RawMessage{}
	}
	return doc, nil
}
