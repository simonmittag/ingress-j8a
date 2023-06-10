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

//TODO cant test this it panics
//func TestAuthenticateToKubeInternal(t *testing.T) {
//	s := NewServer()
//	s.Kube.Client = fake.NewSimpleClientset()
//
//	s.authenticateToKubeInternal()
//}

func TestDetectKubeVersion(t *testing.T) {
	s := NewServer()
	s.Kube.Client = fake.NewSimpleClientset()

	s.detectKubeVersion()
}
