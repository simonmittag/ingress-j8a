package server

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
)

const Version = "v0.1"

type Server struct {
	Version string
	Kube    *kubernetes.Clientset
}

func NewServer() *Server {
	return &Server{
		Version: Version,
		Kube:    nil,
	}
}

func (s *Server) Bootstrap() {
	s.authenticateToKube()
	for {
		s.printPods()
	}
}

func (s *Server) authenticateToKube() {
	config, err := rest.InClusterConfig()
	if err != nil {
		s.panic(err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		s.panic(err)
	}
	s.Kube = clientset
}

func (s *Server) printPods() {
	pods, _ := s.Kube.CoreV1().Pods("").List(
		context.TODO(),
		metav1.ListOptions{})
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	time.Sleep(time.Second * 3)
}

func (s *Server) panic(e error) {
	fmt.Printf("error: %v", e)
}
