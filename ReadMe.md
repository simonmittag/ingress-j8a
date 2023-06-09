[![Circleci Builds](https://circleci.com/gh/simonmittag/ingress-j8a.svg?style=shield)](https://circleci.com/gh/simonmittag/ingress-j8a)
[![Dependabot](https://badgen.net/badge/Dependabot/enabled/green?icon=dependabot)](https://github.com/simonmittag/ingress-j8a/pulls?q=is%3Aopen+is%3Apr)
[![Github Issues](https://img.shields.io/github/issues/simonmittag/ingress-j8a)](https://github.com/simonmittag/ingress-j8a/issues)
[![Github Activity](https://img.shields.io/github/commit-activity/m/simonmittag/ingress-j8a)](https://img.shields.io/github/commit-activity/m/simonmittag/ingress-j8a)  
[![Go Report](https://goreportcard.com/badge/github.com/simonmittag/ingress-j8a)](https://goreportcard.com/report/github.com/simonmittag/ingress-j8a)
[![CodeClimate Maintainability](https://api.codeclimate.com/v1/badges/854887a41d9d23f0ecec/maintainability)](https://codeclimate.com/github/simonmittag/ingress-j8a/maintainability)
[![CodeClimate Test Coverage](https://api.codeclimate.com/v1/badges/854887a41d9d23f0ecec/test_coverage)](https://codeclimate.com/github/simonmittag/ingress-j8a/code)
[![Go Version](https://img.shields.io/github/go-mod/go-version/simonmittag/ingress-j8a)](https://img.shields.io/github/go-mod/go-version/simonmittag/ingress-j8a)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Why?
This document outlines the design of an Ingress Controller for [j8a](https://github.com/simonmittag/j8a) inside a 
Kubernetes cluster. The Ingress Controller is responsible for managing incoming network traffic to `service` objects within the 
cluster, providing a highly available, j8a based entry point for external clients to access the cluster's applications. The controller utilizes the 
`ingress` resource, along with other Kubernetes objects such as `service`, `configMap`, and `secret`, to facilitate routing 
and load balancing of network traffic.

# What?
`ingress-j8a` is a kubernetes ingress controller pod, exposing ports 80, 443 of managed j8a pods, making the cluster accessible to the internet. It generates the configuration
objects for j8a, keeps those configurations updated and manages instances of j8a within the cluster. 

![](art/ingress-j8a.png)
* `ingress-j8a` talks to kube apiserver via the golang kubernetes client and authenticates internal to the cluster with `j8a-serviceaccount` that is deployed together with the ingresscontroller. The `j8a-serviceaccount` has an associated `j8a-clusterrole` and `j8a-clusterrolebinding` to give it minimum privileges required to access cluster-wide `ingress` `ingressclass` `service` `configMap` and `secret` resources required.
* `ingress-j8a` consumes cluster users `ingress` resources from all namespaces for the `ingressClass` j8a
* `ingress-j8a` creates the ingressClass resource that specifies the controller implementation itself. 
  * J8a metadata (🚧 timeouts?) is controlled by modifying this resource and specifying `spec.parameters.key` that reconfigure j8a
* `ingress-j8a` creates a `deployment` of j8a into the cluster by talking to the kubernetes API server. 
  * once `ingress-j8a` is undeployed, the dependent deployment of j8a pods will remain. upon re-deploy the controller recognizes the existing deployment.
  * Pods use off-the-shelf j8a images from dockerhub.
  * Proxy config is passed via env internally.
  * When proxy config needs to change, the deployment is updated with the contents of the env variable.
* `ingress-j8a` allocates a `service` of type loadbalancer that forwards traffic to the proxy server pods.
* j8a `pod` itself exposes ports 80 and 443 on it's clusterIp (depends on config from ingress.yml). It is accessed externally via the outer load balancer.
* j8a routes traffic to pods that are mapped by translation of `service` urls to actual pods inside the cluster. 

# How?
## Design Goals
* Zero downtime deployments for j8a during updates to all cluster resources.
* Redundancy for j8a with multiple proxy server instances and a load balancing mechanism
* Intelligent defaults for j8a for proxy server params the kubernetes ingress resource does not readily expose.

## Resource Lifecycle
The basic mechanics of monitoring kubernetes for configuration changes,
then updating J8a's config and it's live traffic routes.

![](art/ingress-j8a-mechanics.png)
1. The user deploys `ingress` resources to the cluster, or updates them. This is similar for dependent resources such as `configMap` and `secret` that are used by the `ingress` resources. The user is allowed to deploy these at any time.
2. A cache that runs inside `ingress-j8a` monitors for updates to kube resources in all namespaces. It pulls down the latest resources, caches them, then versions its own config. This mechanism has an idle wait safeguard to protect against versioning too frequently.
3. The control loop inside `ingress-j8a` that continuously waits for config changes is notified (this idea is borrowed from ingress-nginx).
4. The control loop reads the versioned, cached config out and generates a j8a config object in yml format. This is based on a template of the j8a config, filled in using go {{template}} variables. The result will be deployed to the kube cluster as its own configmap object in the j8a namespace.
5. `ingress-j8a` then deploys the `configMap` as a resource to the kube api server and keeps it updated for subsequent changes.
6. kube api server deploys this resource into the cluster and maintains it there as a source of truth for the current config, outside of the cache of `ingress-j8a`. 
7. `ingress-j8a` then tells kube api server to deploy the latest docker image of j8a into the cluster using this config. It updates the current deployment for j8a and deploys new pods into that deploying using a rolling configuration update. 
8. kube apiserver updates the `deployment` using the passed in config via the descriptor. pods are updating by creation of a new `replicaset` (not pictured) that scales up while the old one scales down.
9. kube apiserver updates the `service` with a `labelselector` to tell the loadbalancer about the new proxy pods with their updated config.


# Contributions

The ingress-j8a team welcomes all [contributors](https://github.com/simonmittag/ingress-j8a/blob/master/CONTRIBUTING.md). Everyone
interacting with the project's codebase, issue trackers, chat rooms and mailing lists is expected to follow
the [code of conduct](https://github.com/simonmittag/ingress-j8a/blob/master/CODE_OF_CONDUCT.md)
