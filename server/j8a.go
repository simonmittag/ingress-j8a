package server

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func int32Ptr(i int32) *int32 { return &i }

func (s *Server) createJ8aNamespace() *Server {
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
	return s
}

func (s *Server) createJ8aDeployment() *Server {
	var v string
	if strings.HasPrefix(s.J8a.Version, "v") {
		v = s.J8a.Version[1:]
	} else {
		v = s.J8a.Version
	}

	deploymentsClient := s.Kube.Client.AppsV1().Deployments(s.Kube.Namespace.ObjectMeta.Name)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "j8a-deployment",
			Namespace: s.Kube.Namespace.ObjectMeta.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(int32(s.J8a.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "j8a",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "j8a",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "j8a",
							Image: "simonmittag/j8a:" + v,
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
		s.Log.Fatalf("unable to deploy j8a to cluster, cause %v", err)
	}
	s.Log.Infof("deployed %v to cluster.", result.GetObjectMeta().GetName())

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
