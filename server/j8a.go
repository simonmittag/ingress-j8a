package server

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

func int32Ptr(i int32) *int32 { return &i }

func (s *Server) createJ8aNamespace() *Server {

	nsName := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.J8a.Namespace,
		},
	}

	ns, e := s.Kube.Client.CoreV1().Namespaces().
		Create(context.Background(), nsName, metav1.CreateOptions{})
	if e == nil {
		s.J8a.Namespace = ns.ObjectMeta.Name
		s.Log.Infof("created namespace %v", ns.ObjectMeta.Name)
	} else {
		ns, e := s.Kube.Client.CoreV1().Namespaces().
			Get(context.Background(), s.J8a.Namespace, metav1.GetOptions{})
		if ns != nil {
			s.J8a.Namespace = ns.ObjectMeta.Name
			s.Log.Infof("detected namespace %v", ns.ObjectMeta.Name)
		}
		if e != nil {
			s.panic(fmt.Errorf("unable to create or namespace J8a in cluster, cause %v", e))
		}
	}
	return s
}

func (s *Server) createJ8aServiceTypeLoadBalancer() *Server {

	servicesClient := s.Kube.Client.CoreV1().Services(s.J8a.Namespace)

	// Define the service
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.J8a.Service,
			Namespace: s.J8a.Namespace,
		},
		Spec: apiv1.ServiceSpec{
			Selector: s.J8a.Pod.Label,
			Type:     apiv1.ServiceTypeLoadBalancer,
			Ports: []apiv1.ServicePort{
				{
					Name:       "http",
					Protocol:   apiv1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(80), // Assuming the pods expose port 80
				},
				{
					Name:       "https",
					Protocol:   apiv1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(443), // Assuming the pods expose port 443
				},
			},
		},
	}

	result, err := servicesClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		s.Log.Fatalf("unable to create service %v in cluster, cause %v", s.J8a.Service, err)
	}
	s.Log.Infof("created service %v in cluster.", result.GetObjectMeta().GetName())

	return s
}

func (s *Server) createJ8aDeployment() *Server {
	var v string
	if strings.HasPrefix(s.J8a.Version, "v") {
		v = s.J8a.Version[1:]
	} else {
		v = s.J8a.Version
	}

	deploymentsClient := s.Kube.Client.AppsV1().Deployments(s.J8a.Namespace)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.J8a.Deployment.Name,
			Namespace: s.J8a.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(int32(s.J8a.Deployment.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: s.J8a.Pod.Label,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: s.J8a.Pod.Label,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  s.J8a.Pod.Name,
							Image: s.J8a.Image + ":" + v,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
								{
									Name:          "https",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 443,
								},
							},
							Env: []apiv1.EnvVar{{
								Name:  "J8ACFG_YML",
								Value: getInitialJ8aConfig(),
							}},
						},
					},
				},
			},
		},
	}

	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		s.Log.Fatalf("unable to created deployment %v in cluster, cause %v", s.J8a.Deployment.Name, err)
	}
	s.Log.Infof("created deployment %v in cluster.", result.GetObjectMeta().GetName())

	return s
}

func getInitialJ8aConfig() string {
	return `---
            connection:
              downstream:
                http:
                  port: 80
            routes:
              - path: "/"
                resource: placeholder
            resources:
              placeholder:
                - url:
                    scheme: http
                    host: localhost
                    port: 59999`
}
