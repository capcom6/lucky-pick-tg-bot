package fsm

import (
	"encoding/json"
	"fmt"
	"strings"
)

type State struct {
	Name string            `json:"name"`
	Data map[string]string `json:"data"`
}

func (s *State) SetName(name string) {
	s.Name = name
}

func (s *State) IsPrefixed(prefix string) bool {
	return strings.HasPrefix(s.Name, prefix)
}

func (s *State) AddData(key, value string) {
	if s.Data == nil {
		s.Data = make(map[string]string)
	}

	s.Data[key] = value
}

func (s *State) GetData(key string) string {
	return s.Data[key]
}

func (s *State) RemoveData(key string) {
	delete(s.Data, key)
}

func (s *State) ClearData() {
	s.Data = nil
}

func (s *State) Clear() {
	s.Name = ""
	s.ClearData()
}

func (s *State) Marshal() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshal state: %w", err)
	}

	return b, nil
}

func (s *State) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, s); err != nil {
		return fmt.Errorf("unmarshal state: %w", err)
	}

	return nil
}
