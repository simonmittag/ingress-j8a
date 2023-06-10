package server

import "testing"

func TestLogger(t *testing.T) {
	l := NewKLoggerWrapper()
	l.Infof("blah %v", "blah")
}
