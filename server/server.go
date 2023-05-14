package server

import (
	"context"
	"flag"
	"fmt"
	v1 "k8s.io/api/core/v1"
	nv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strconv"
	"time"
)

const Version = "v0.1"

type Server struct {
	Version          string
	Kube             *kubernetes.Clientset
	KubeVersionMajor int
	KubeVersionMinor int
}

func NewServer() *Server {
	return &Server{
		Version:          Version,
		Kube:             nil,
		KubeVersionMajor: 0,
		KubeVersionMinor: 0,
	}
}

func (s *Server) Bootstrap() {
	s.authenticateToKubeInternal()
	//s.authenticateToKubeExternal()
	for {
		s.printObjects()
	}
}

func (s *Server) authenticateToKubeExternal() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "~/.kube/config", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, _ := kubernetes.NewForConfig(config)
	s.Kube = clientset
}

func (s *Server) authenticateToKubeInternal() {
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
	s.detectKubeVersion(config)
}

func (s *Server) detectKubeVersion(config *rest.Config) {
	dc, _ := discovery.NewDiscoveryClientForConfig(config)
	vi, _ := dc.ServerVersion()
	s.KubeVersionMajor, _ = strconv.Atoi(vi.Major)
	s.KubeVersionMinor, _ = strconv.Atoi(vi.Minor)
}

func (s *Server) printObjects() {
	cm := s.fetchConfigMaps()
	sv := s.fetchServices()
	sl := s.fetchSecrets()
	il := s.fetchIngress()
	fmt.Printf("There are %d config maps in cluster running kube v%v.%v\n", len(cm.Items), s.KubeVersionMajor, s.KubeVersionMinor)
	fmt.Printf("There are %d services in cluster running kube v%v.%v\n", len(sv.Items), s.KubeVersionMajor, s.KubeVersionMinor)
	fmt.Printf("There are %d secrets in cluster running kube v%v.%v\n", len(sl.Items), s.KubeVersionMajor, s.KubeVersionMinor)
	fmt.Printf("There are %d ingress in cluster running kube v%v.%v\n", len(il.Items), s.KubeVersionMajor, s.KubeVersionMinor)
	time.Sleep(time.Second * 3)
}

func (s *Server) fetchServices() *v1.ServiceList {
	sv, _ := s.Kube.CoreV1().Services("").List(
		context.TODO(),
		metav1.ListOptions{})
	return sv
}

func (s *Server) fetchConfigMaps() *v1.ConfigMapList {
	cml, _ := s.Kube.CoreV1().ConfigMaps("").List(
		context.TODO(),
		metav1.ListOptions{})
	return cml
}

func (s *Server) fetchSecrets() *v1.SecretList {
	sl, _ := s.Kube.CoreV1().Secrets("").List(
		context.TODO(),
		metav1.ListOptions{})
	return sl
}

func (s *Server) fetchIngress() *nv1.IngressList {
	il, e := s.Kube.NetworkingV1().Ingresses("").List(
		context.TODO(),
		metav1.ListOptions{})
	fmt.Printf("%v", e)
	return il
}

func (s *Server) panic(e error) {
	fmt.Printf("error: %v", e)
}
