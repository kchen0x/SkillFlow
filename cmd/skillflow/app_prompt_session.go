package main

import (
	"fmt"
	"sync"

	promptcatalogapp "github.com/shinerio/skillflow/core/promptcatalog/app"
)

type PromptImportPrepareResult struct {
	SessionID string                          `json:"sessionId"`
	Creates   []promptcatalogapp.ImportPrompt `json:"creates"`
	Conflicts []promptcatalogapp.ImportPrompt `json:"conflicts"`
}

type promptImportSession struct {
	FilePath string
	Preview  *promptcatalogapp.ImportPreview
}

type promptImportSessionStore struct {
	mu       sync.Mutex
	nextID   uint64
	sessions map[string]*promptImportSession
}

func newPromptImportSessionStore() *promptImportSessionStore {
	return &promptImportSessionStore{
		sessions: make(map[string]*promptImportSession),
	}
}

func (s *promptImportSessionStore) Create(filePath string, preview *promptcatalogapp.ImportPreview) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	id := fmt.Sprintf("prompt-import-%d", s.nextID)
	s.sessions[id] = &promptImportSession{
		FilePath: filePath,
		Preview:  clonePromptImportPreview(preview),
	}
	return id
}

func (s *promptImportSessionStore) Get(id string) (*promptImportSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	return clonePromptImportSession(session), true
}

func (s *promptImportSessionStore) Take(id string) (*promptImportSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	delete(s.sessions, id)
	return clonePromptImportSession(session), true
}

func (s *promptImportSessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

func clonePromptImportSession(session *promptImportSession) *promptImportSession {
	if session == nil {
		return nil
	}
	return &promptImportSession{
		FilePath: session.FilePath,
		Preview:  clonePromptImportPreview(session.Preview),
	}
}

func clonePromptImportPreview(preview *promptcatalogapp.ImportPreview) *promptcatalogapp.ImportPreview {
	if preview == nil {
		return nil
	}
	return &promptcatalogapp.ImportPreview{
		Creates:   append([]promptcatalogapp.ImportPrompt(nil), preview.Creates...),
		Conflicts: append([]promptcatalogapp.ImportPrompt(nil), preview.Conflicts...),
	}
}
