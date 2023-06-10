package server

import "testing"

func TestLogger(t *testing.T) {
	l := NewKLoggerWrapper()
	l.Infof("blah %v", "blah")
	l.Info("blah")
	l.Errorf("blah %v", "blah")
}
