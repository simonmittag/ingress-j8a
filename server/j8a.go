package server

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func int32Ptr(i int32) *int32 { return &i }

func (s *Server) createJ8aNamespace() {
	const j8aNamespace string = "j8a"

	nsName := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: j8aNamespace,
		},
	}

	ns, e := s.Kube.Client.CoreV1().Namespaces().
		Create(context.Background(), nsName, metav1.CreateOptions{})
	if e == nil {
		s.Kube.Namespace = ns
		s.Log.Infof("created namespace %v", ns.ObjectMeta.Name)
	} else {
		ns, e := s.Kube.Client.CoreV1().Namespaces().
			Get(context.Background(), j8aNamespace, metav1.GetOptions{})
		if ns != nil {
			s.Kube.Namespace = ns
			s.Log.Infof("detected namespace %v", ns.ObjectMeta.Name)
		}
		if e != nil {
			s.panic(fmt.Errorf("unable to create or locate J8a namespace in cluster, cause %v", e))
		}
	}
}

func (s *Server) createJ8aDeployment() {
	deploymentsClient := s.Kube.Client.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}
