package server

import "testing"

func TestLogger(t *testing.T) {
	l := NewKLoggerWrapper()
	l.Infof("blah %v", "blah")
	l.Errorf("blah %v", "blah")
	l.Fatalf("blah %v", "blah")
}
