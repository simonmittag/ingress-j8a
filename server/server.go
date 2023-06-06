package server

import (
	"context"
	"errors"
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

var KubeVersionMinimum = Kube{
	VersionMajor: 1,
	VersionMinor: 22,
}

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
	e = s.checkKubeVersion()
	if e != nil {
		s.panic(e)
	}
	e = s.checkPermissions()
	if e != nil {
		s.panic(e)
	}

	for {
		s.logObjects()
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

func (s *Server) checkKubeVersion() error {
	s.detectKubeVersion()
	if s.Kube.VersionMajor < KubeVersionMinimum.VersionMajor ||
		s.Kube.VersionMinor < KubeVersionMinimum.VersionMinor {
		msg := fmt.Sprintf("Detected unsupported Kubernetes version %v.%v", s.Kube.VersionMajor, s.Kube.VersionMinor)
		fmt.Println(msg)
		return errors.New(msg)
	} else {
		fmt.Printf("\nDetected Kubernetes version %v.%v", s.Kube.VersionMajor, s.Kube.VersionMinor)
		return nil
	}
}

func (s *Server) checkPermissions() error {
	_, e1 := s.fetchConfigMaps()
	if e1 != nil {
		fmt.Printf("\ninsufficient privileges to access configMaps")
		return e1
	}

	_, e2 := s.fetchServices()
	if e2 != nil {
		fmt.Printf("\ninsufficient privileges to access services")
		return e2
	}

	_, e3 := s.fetchSecrets()
	if e3 != nil {
		fmt.Printf("insufficient privileges to secrets")
		return e3
	}

	_, e4 := s.fetchIngress()
	if e4 != nil {
		fmt.Printf("\ninsufficient privileges to access ingress")
		return e4
	} else {
		fmt.Printf("\nsuccessfully checked privileges to access cluster configuration")
		return nil
	}
}

func (s *Server) logObjects() {
	cm, _ := s.fetchConfigMaps()
	sv, _ := s.fetchServices()
	sl, _ := s.fetchSecrets()
	il, _ := s.fetchIngress()
	fmt.Printf("\ndetected %d config maps in cluster", len(cm.Items))
	fmt.Printf("\ndetected %d services in cluster", len(sv.Items))
	fmt.Printf("\ndetected %d secrets in cluster", len(sl.Items))
	fmt.Printf("\ndetected %d ingress in cluster", len(il.Items))
	//klog.Info("accessed config objects")
	time.Sleep(time.Second * 7)
}

func (s *Server) fetchServices() (*v1.ServiceList, error) {
	return s.Kube.Client.CoreV1().Services("").List(
		context.TODO(),
		metav1.ListOptions{})
}

func (s *Server) fetchConfigMaps() (*v1.ConfigMapList, error) {
	return s.Kube.Client.CoreV1().ConfigMaps("").List(
		context.TODO(),
		metav1.ListOptions{})
}

func (s *Server) fetchSecrets() (*v1.SecretList, error) {
	return s.Kube.Client.CoreV1().Secrets("").List(
		context.TODO(),
		metav1.ListOptions{})
}

func (s *Server) fetchIngress() (*nv1.IngressList, error) {
	return s.Kube.Client.NetworkingV1().Ingresses("").List(
		context.TODO(),
		metav1.ListOptions{})
}

func (s *Server) panic(e error) {
	fmt.Printf("error: %v", e)
}
