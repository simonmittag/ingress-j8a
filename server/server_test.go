package server

import (
	"k8s.io/client-go/kubernetes/fake"
	"os"
	"testing"
)

func setupSuite(tb testing.TB) func(tb testing.TB) {
	os.Setenv("INGRESS_J8A_TEST_NOEXIT", "TRUE")

	return func(tb testing.TB) {
		os.Unsetenv("INGRESS_J8A_TEST_NOEXIT")
	}
}

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

func TestBootstrap(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()

	s.Bootstrap()
}
