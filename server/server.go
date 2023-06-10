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
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const Version = "v0.1.2"

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
	Attr    map[Attr]Attr
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
	Version      string
	Image        string
	Namespace    string
	IngressClass string
	Deployment   Deployment
	Service      string
	Pod          Pod
}

type Kube struct {
	Client  kubernetes.Interface
	Config  *rest.Config
	Version KVersion
}

type KVersion struct {
	Major int
	Minor int
}

type Attr string

const (
	TestNoExit Attr = "testNoExit"
)

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
			IngressClass: "ingress-j8a",
			Service:      "loadbalancer-j8a",
			Pod: Pod{
				Name:  "j8a",
				Label: map[string]string{"app": "j8a"},
			},
		},
		Log:  NewKLoggerWrapper(),
		Attr: make(map[Attr]Attr),
	}
}

func (s *Server) Bootstrap() {
	s.authenticate().
		checkKubeVersion().
		checkPermissions().
		createOrDetectJ8aNamespace().
		createOrDetectJ8aIngressClass().
		createOrDetectJ8aDeployment().
		createOrDetectJ8aServiceTypeLoadBalancer().
		updateJ8aDeploymentWithFullClusterConfig()

	for i := 0; i < 10; i++ {
		s.logObjects()

		//klog.Info("accessed config objects")
		time.Sleep(time.Second * 6)
	}
}

func (s *Server) authenticate() *Server {
	//TODO for now always authenticate external first to make development easier.
	//put this behind a flag eventually and do internal as default
	if e := s.authenticateToKubeExternal(); e != nil {
		e = s.authenticateToKubeInternal()
		if e != nil {
			s.panic(e)
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
		s.Kube.Client = clientset
	} else {
		return err
	}

	apiserver, err := clientset.CoreV1().Services("default").Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err == nil {
		s.Log.Infof("authenticated external to kubernetes control plane running at %v uid: %v", config.Host, apiserver.ObjectMeta.UID)
	} else {
		return err
	}
	return nil
}

func (s *Server) authenticateToKubeInternal() error {
	const intErrMsg = "unable to authenticate internal to kubernetes control plane running at %v, cause: %v"

	config, err := rest.InClusterConfig()
	if err == nil {
		s.Kube.Config = config
		clientset, err := kubernetes.NewForConfig(config)
		if err == nil {
			s.Kube.Client = clientset
		} else {
			s.panic(fmt.Errorf(intErrMsg, config.Host, err))
		}
		apiserver, err := clientset.CoreV1().Services("default").Get(context.TODO(), "kubernetes", metav1.GetOptions{})
		if err == nil {
			s.Log.Infof("authenticated internal to kubernetes control plane running at %v uid: %v", config.Host, apiserver.ObjectMeta.UID)
		} else {
			s.panic(fmt.Errorf(intErrMsg, config.Host, err))
		}
	} else {
		host := "'undefined'"
		if config != nil {
			host = config.Host
		}
		return fmt.Errorf(intErrMsg, host, err)
	}
	return nil
}

func (s *Server) detectKubeVersion() {
	defer func() {
		if err := recover(); err != nil {
			// Handle the panic or error here
			klog.Error("unable to detect kube version")
		}
	}()

	dc, e := discovery.NewDiscoveryClientForConfig(s.Kube.Config)

	vi, e := dc.ServerVersion()
	if e == nil && vi != nil {
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
		s.Log.Info("successfully checked privileges to access cluster configuration objects in all namespaces")
	}
	return s
}

func (s *Server) logObjects() {
	cm, _ := s.fetchConfigMaps()
	s.Log.Infof("detected %d config maps", len(cm.Items))
	sv, _ := s.fetchServices()
	s.Log.Infof("detected %d services", len(sv.Items))
	sl, _ := s.fetchSecrets()
	s.Log.Infof("detected %d secrets", len(sl.Items))
	il, _ := s.fetchIngress()
	s.Log.Infof("detected %d ingress", len(il.Items))
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
	if !s.hasAttr(TestNoExit) {
		msg := "shutdown cause: %v"
		s.Log.Fatalf(msg, e)
		os.Exit(-1)
	}
}

func (s *Server) hasAttr(attr Attr) bool {
	_, ok := s.Attr[attr]
	return ok
}
