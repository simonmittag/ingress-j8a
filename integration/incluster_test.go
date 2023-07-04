package integration

import (
	"context"
	"errors"
	"github.com/simonmittag/ingress-j8a/server"
	"io/ioutil"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"testing"
)

func TestFetchBackendServicePort(t *testing.T) {
	i := upsertIngressResource("ingress-prefixpath.yml", t)

	s := server.NewServer()

	got, _ := s.FetchBackendServicePort(i.Spec.Rules[0].HTTP.Paths[0].Backend)
	want := "80"
	if got != want {
		t.Errorf("port not extracted, got %v want %v", got, want)
	}
}

func upsertIngressResource(res string, t *testing.T) *netv1.Ingress {
	// Load the Kubernetes configuration
	config, err := loadKubeConfig()
	if err != nil {
		t.Errorf("unable to load kube config %v", err)
	}

	// Create a Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Errorf("unable to create kube client %v", err)
	}

	i, e := readResource(res)
	if e != nil {
		t.Errorf("resource not found: %v", e)
	}

	err = deployIngress(i, clientset)
	if err != nil {
		t.Errorf("unable to deploy ingress with named port %v", err)
	}
	return i
}

func readResource(f string) (*netv1.Ingress, error) {
	f1 := "./resources/examples/" + f
	u, e := readIngressFromFile(f1)
	if e == nil {
		return u, e
	} else {
		f2 := "." + f1
		return readIngressFromFile(f2)
	}
}

func readIngressFromFile(filename string) (*netv1.Ingress, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Create a new runtime.Scheme and register the networkingv1 scheme
	scheme := runtime.NewScheme()
	err = netv1.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	// Create a runtime decoder
	decode := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode

	// Decode the YAML file into an unstructured object
	obj, _, err := decode([]byte(data), nil, nil)
	if err != nil {
		return nil, err
	}

	i, ok := obj.(*netv1.Ingress)
	if ok {
		return i, nil
	} else {
		return nil, errors.New("unexpected object type not ingress")
	}
}

func deployIngress(ingress *netv1.Ingress, clientset *kubernetes.Clientset) error {
	// Get the existing Ingress, if it exists
	existingIngress, err := clientset.NetworkingV1().Ingresses(ingress.GetNamespace()).Get(context.TODO(), ingress.GetName(), metav1.GetOptions{})
	if err == nil {
		// Ingress already exists, perform an update
		existingIngress.SetResourceVersion(existingIngress.GetResourceVersion())
		_, updateErr := clientset.NetworkingV1().Ingresses(existingIngress.GetNamespace()).Update(context.TODO(), ingress, metav1.UpdateOptions{})
		if updateErr != nil {
			return updateErr
		}
	} else {
		// Ingress does not exist, perform a create
		_, err = clientset.NetworkingV1().Ingresses(ingress.GetNamespace()).Create(context.TODO(), ingress, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func loadKubeConfig() (*rest.Config, error) {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
