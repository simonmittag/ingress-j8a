package server

import "testing"

func TestNewServer(t *testing.T) {
	s := NewServer()
	s.Log.Info("thou shalt pass")

	if s.Log == nil {
		t.Errorf("log should be configured")
	}
}
