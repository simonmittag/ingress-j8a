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
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const Version = "v0.1.1"

var KubeVersionMinimum = Kube{
	Version: KVersion{
		Major: 1,
		Minor: 22,
	},
}

type Server struct {
	Version string
	Kube    *Kube
	J8a     *J8a
	Log     Logger
}

type Deployment struct {
	Name     string
	Replicas int
}

type Pod struct {
	Name  string
	Label map[string]string
}

type J8a struct {
	Version    string
	Image      string
	Namespace  string
	Deployment Deployment
	Service    string
	Pod        Pod
}

type Kube struct {
	Client  *kubernetes.Clientset
	Config  *rest.Config
	Version KVersion
}

type KVersion struct {
	Major int
	Minor int
}

// TODO: this method contains a lot of defaults
func NewServer() *Server {
	return &Server{
		Version: Version,
		Kube: &Kube{
			Client: nil,
			Config: nil,
			Version: KVersion{
				Major: 0,
				Minor: 0,
			},
		},
		J8a: &J8a{
			Version:   "v1.0.1",
			Image:     "simonmittag/j8a",
			Namespace: "j8a",
			Deployment: Deployment{
				Name:     "deployment-j8a",
				Replicas: 3,
			},
			Service: "loadbalancer-j8a",
			Pod: Pod{
				Name:  "j8a",
				Label: map[string]string{"app": "j8a"},
			},
		},
		Log: NewKLoggerWrapper(),
	}
}

func (s *Server) Bootstrap() {
	s.authenticate().
		checkKubeVersion().
		checkPermissions().
		createOrDetectJ8aNamespace().
		createOrDetectJ8aDeployment().
		createOrDetectJ8aServiceTypeLoadBalancer().
		updateJ8aDeploymentWithFullClusterConfig()

	for {
		s.logObjects()

		//klog.Info("accessed config objects")
		time.Sleep(time.Second * 7)
	}
}

func (s *Server) authenticate() *Server {
	//TODO for now always authenticate external first to make development easier.
	//put this behind a flag eventually and do internal as default
	if e := s.authenticateToKubeExternal(); e != nil {
		e1 := s.authenticateToKubeInternal()
		if e1 != nil {
			s.panic(fmt.Errorf("unable to authenticate to cluster, cause %v", e1))
		}
	}
	return s
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
		s.Log.Info("authenticated external to cluster")
		s.Kube.Client = clientset
	} else {
		return err
	}
	return nil
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
		s.Log.Info("authenticated inside cluster")
		s.Kube.Client = clientset
		return nil
	} else {
		return err
	}

}

func (s *Server) detectKubeVersion() {
	dc, _ := discovery.NewDiscoveryClientForConfig(s.Kube.Config)
	vi, e := dc.ServerVersion()
	if e == nil {
		s.Kube.Version.Major, _ = strconv.Atoi(vi.Major)
		s.Kube.Version.Minor, _ = strconv.Atoi(vi.Minor)
	}
}

func (s *Server) checkKubeVersion() *Server {
	s.detectKubeVersion()
	if s.Kube.Version.Major < KubeVersionMinimum.Version.Major ||
		s.Kube.Version.Minor < KubeVersionMinimum.Version.Minor {
		e := errors.New(fmt.Sprintf("detected unsupported Kubernetes version %v.%v", s.Kube.Version.Major, s.Kube.Version.Minor))
		s.panic(e)
	} else {
		s.Log.Infof("detected Kubernetes version %v.%v", s.Kube.Version.Major, s.Kube.Version.Minor)
	}
	return s
}

func (s *Server) checkPermissions() *Server {
	insufficient := "insufficient privileges to access "
	_, e1 := s.fetchConfigMaps()
	if e1 != nil {
		s.panic(fmt.Errorf(insufficient+"configMaps", e1))
	}

	_, e2 := s.fetchServices()
	if e2 != nil {
		s.panic(fmt.Errorf(insufficient+"services", e2))
	}

	_, e3 := s.fetchSecrets()
	if e3 != nil {
		s.panic(fmt.Errorf(insufficient+"secrets", e3))
	}

	_, e4 := s.fetchIngress()
	if e4 != nil {
		s.panic(fmt.Errorf(insufficient+"ingress", e4))
	} else {
		s.Log.Info("successfully checked privileges to access cluster configuration")
	}
	return s
}

func (s *Server) logObjects() {
	cm, _ := s.fetchConfigMaps()
	s.Log.Infof("detected %d config maps in cluster", len(cm.Items))
	sv, _ := s.fetchServices()
	s.Log.Infof("detected %d services in cluster", len(sv.Items))
	sl, _ := s.fetchSecrets()
	s.Log.Infof("detected %d secrets in cluster", len(sl.Items))
	il, _ := s.fetchIngress()
	s.Log.Infof("detected %d ingress in cluster", len(il.Items))
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
	s.Log.Fatalf("cannot continue because: %v", e)
	s.Log.Fatalf("system shutting down")
	os.Exit(-1)
}
