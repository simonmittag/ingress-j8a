package server

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var Version = "v0.1"
var kubeClient *kubernetes.Clientset

func Bootstrap() {
	kubeClient = authenticate()
	pods, _ := kubeClient.CoreV1().Pods("").List(
		context.TODO(),
		metav1.ListOptions{})
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
}

func authenticate() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
