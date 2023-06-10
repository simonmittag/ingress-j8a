package server

import (
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	s.Log.Info("thou shalt pass")

	if s.Log == nil {
		t.Errorf("log should be configured")
	}
}

func TestAuthenticateToKubeExternal(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()

	s.authenticateToKubeExternal()
}

func TestAuthenticateToKubeInternal(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()

	e := s.authenticateToKubeInternal()
	if e == nil {
		t.Errorf("should not authenticate outside cluster")
	}
}

func TestDetectKubeVersion(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()

	s.detectKubeVersion()
}

func TestCheckPermissions(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()

	s.checkPermissions()
}

func TestLogObjects(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()

	s.logObjects()
}
