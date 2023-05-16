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
	Version string
	Kube    *Kube
}

type Kube struct {
	Client       *kubernetes.Clientset
	Config       *rest.Config
	VersionMajor int
	VersionMinor int
}

func NewServer() *Server {
	return &Server{
		Version: Version,
		Kube: &Kube{
			Client:       nil,
			Config:       nil,
			VersionMajor: 0,
			VersionMinor: 0,
		},
	}
}

func (s *Server) Bootstrap() {
	e := s.authenticate()
	if e != nil {
		s.panic(e)
	}

	s.detectKubeVersion()

	for {
		s.printObjects()
	}
}

func (s *Server) authenticate() error {
	//TODO for now always authenticate external first to make development easier.
	//put this behind a flag eventually and do internal as default
	if e := s.authenticateToKubeExternal(); e != nil {
		e1 := s.authenticateToKubeInternal()
		if e1 != nil {
			return e
		} else {
			return nil
		}
	}
	return nil
}

func (s *Server) authenticateToKubeExternal() error {
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
		return err
	} else {
		s.Kube.Config = config
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err == nil {
		s.Kube.Client = clientset
		return nil
	} else {
		return err
	}
}

func (s *Server) authenticateToKubeInternal() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	} else {
		s.Kube.Config = config
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err == nil {
		s.Kube.Client = clientset
		return nil
	} else {
		return err
	}
}

func (s *Server) detectKubeVersion() {
	dc, _ := discovery.NewDiscoveryClientForConfig(s.Kube.Config)
	vi, _ := dc.ServerVersion()
	s.Kube.VersionMajor, _ = strconv.Atoi(vi.Major)
	s.Kube.VersionMinor, _ = strconv.Atoi(vi.Minor)
}

func (s *Server) printObjects() {
	cm := s.fetchConfigMaps()
	sv := s.fetchServices()
	sl := s.fetchSecrets()
	il := s.fetchIngress()
	fmt.Printf("There are %d config maps in cluster running kube v%v.%v\n", len(cm.Items), s.Kube.VersionMajor, s.Kube.VersionMinor)
	fmt.Printf("There are %d services in cluster running kube v%v.%v\n", len(sv.Items), s.Kube.VersionMajor, s.Kube.VersionMinor)
	fmt.Printf("There are %d secrets in cluster running kube v%v.%v\n", len(sl.Items), s.Kube.VersionMajor, s.Kube.VersionMinor)
	fmt.Printf("There are %d ingress in cluster running kube v%v.%v\n", len(il.Items), s.Kube.VersionMajor, s.Kube.VersionMinor)
	time.Sleep(time.Second * 7)
}

func (s *Server) fetchServices() *v1.ServiceList {
	sv, _ := s.Kube.Client.CoreV1().Services("").List(
		context.TODO(),
		metav1.ListOptions{})
	return sv
}

func (s *Server) fetchConfigMaps() *v1.ConfigMapList {
	cml, _ := s.Kube.Client.CoreV1().ConfigMaps("").List(
		context.TODO(),
		metav1.ListOptions{})
	return cml
}

func (s *Server) fetchSecrets() *v1.SecretList {
	sl, _ := s.Kube.Client.CoreV1().Secrets("").List(
		context.TODO(),
		metav1.ListOptions{})
	return sl
}

func (s *Server) fetchIngress() *nv1.IngressList {
	il, _ := s.Kube.Client.NetworkingV1().Ingresses("").List(
		context.TODO(),
		metav1.ListOptions{})
	return il
}

func (s *Server) panic(e error) {
	fmt.Printf("error: %v", e)
}
