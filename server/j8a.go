package server

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
	"strings"
)

func int32Ptr(i int32) *int32 { return &i }

func (s *Server) createOrDetectJ8aNamespace() *Server {

	nsName := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.J8a.Namespace,
		},
	}

	ns, e := s.Kube.Client.CoreV1().Namespaces().
		Create(context.Background(), nsName, metav1.CreateOptions{})
	if e == nil {
		s.J8a.Namespace = ns.ObjectMeta.Name
		s.Log.Infof("created namespace '%v'", ns.ObjectMeta.Name)
	} else {
		ns, e := s.Kube.Client.CoreV1().Namespaces().
			Get(context.Background(), s.J8a.Namespace, metav1.GetOptions{})
		if ns != nil {
			s.J8a.Namespace = ns.ObjectMeta.Name
			s.Log.Infof("detected namespace '%v'", ns.ObjectMeta.Name)
		}
		if e != nil {
			s.panic(fmt.Errorf("unable to create or namespace %v, cause: %v", s.J8a.Namespace, e))
		}
	}
	return s
}

func (s *Server) createOrDetectJ8aServiceTypeLoadBalancer() *Server {

	servicesClient := s.Kube.Client.CoreV1().Services(s.J8a.Namespace)

	// Define the service
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.J8a.Service,
			Namespace: s.J8a.Namespace,
			Annotations: map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
				//"service.beta.kubernetes.io/aws-load-balancer-internal": "false",
			},
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
		result, err := servicesClient.Get(context.TODO(), s.J8a.Service, metav1.GetOptions{})
		if err != nil {
			s.Log.Fatalf("unable to create or detect service '%v', cause: %v", s.J8a.Service, err)
		} else {
			s.Log.Infof("detected service '%v'", result.ObjectMeta.Name)
		}
	} else {
		s.Log.Infof("created service '%v'", result.GetObjectMeta().GetName())
	}

	return s
}

func (s *Server) createOrDetectJ8aDeployment() *Server {
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
		result, err := deploymentsClient.Get(context.TODO(), s.J8a.Deployment.Name, metav1.GetOptions{})
		if err != nil {
			s.Log.Fatalf("unable to create or detect deployment '%v', cause: %v", s.J8a.Deployment.Name, err)
		} else {
			s.Log.Infof("detected deployment '%v'", result.ObjectMeta.Name)
			r := int(*result.Spec.Replicas)
			if r != s.J8a.Deployment.Replicas {
				//remember the current deployment scale of j8a
				s.J8a.Deployment.Replicas = r
				s.Log.Infof("j8a replicas configuration set to %v based on current value of deployment '%v'", r, result.ObjectMeta.Name)
			}
		}
	} else {
		s.Log.Infof("created deployment '%v'", result.GetObjectMeta().GetName())
	}

	return s
}

func (s *Server) createOrDetectJ8aIngressClass() *Server {
	ingressClassClient := s.Kube.Client.NetworkingV1().IngressClasses()

	// Create the IngressClass resource
	ingressClass := &netv1.IngressClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ingress-j8a",
			Annotations: map[string]string{
				"ingressclass.kubernetes.io/is-default-class": "true",
			},
		},
		Spec: netv1.IngressClassSpec{
			//TODO: what does this point to? the docker image name?
			Controller: "github.com/simonmittag/ingress-j8a",
		},
	}

	// Create the IngressClass using the client
	ic, err := ingressClassClient.Create(context.TODO(), ingressClass, metav1.CreateOptions{})
	if err != nil {
		result, err := ingressClassClient.Get(context.TODO(), "ingress-j8a", metav1.GetOptions{})
		if err != nil {
			s.Log.Fatalf("unable to create or detect ingress class '%v', cause %v", s.J8a.Deployment.Name, err)
		} else {
			s.Log.Infof("detected ingressClass '%v'", result.ObjectMeta.Name)
		}
	} else {
		s.Log.Infof("created ingressClass '%v'", ic.ObjectMeta.Name)
	}

	return s
}

// TODO: this may not work in the future but for initial config.
// server cannot process listener callbacks while this is running (but it could queue them)
func (s *Server) updateJ8aDeploymentWithFullClusterConfig() {
	il, _ := s.fetchIngress()
	s.updateCacheFromIngressList(il)

	// get services
	// get secrets
}

func (s *Server) updateCacheFromIngressList(il *netv1.IngressList) {
	for _, igrs := range il.Items {
		//we only process ingress where class is specified and points to J8a. Older kube versions without
		//ingressclass are not supported.
		if igrs.Spec.IngressClassName != nil && *igrs.Spec.IngressClassName == s.J8a.IngressClass {
			for _, r := range igrs.Spec.Rules {
				if r.HTTP != nil {
					for _, p := range r.HTTP.Paths {
						//TODO: these routes need to be collected in a list
						s.Cache.update(*NewRouteFrom(p.Path, r.Host, p.PathType, s.findServiceDNSName(p.Backend)))
					}
				}
			}
		}
	}
}

func getTemplateJ8aConfig() string {
	f := "resources/j8a/configtemplate.yml"
	t, e := os.ReadFile(f)
	if e != nil {
		t, _ = os.ReadFile("../" + f)
	}
	return string(t)
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
