package server

import (
	"flag"
	"k8s.io/client-go/kubernetes/fake"
	"os"
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

func TestBootstrap(t *testing.T) {
	//uses a flag that needs resetting
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	s := NewServer(TestNoExit)
	if !s.hasOption(TestNoExit) {
		t.Errorf("needs to have test option")
	}
	s.Kube.Client = fake.NewSimpleClientset()

	s.Bootstrap()
}
