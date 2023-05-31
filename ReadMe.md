[![Circleci Builds](https://circleci.com/gh/simonmittag/ingress-j8a.svg?style=shield)](https://circleci.com/gh/simonmittag/ingress-j8a)
[![Dependabot](https://badgen.net/badge/Dependabot/enabled/green?icon=dependabot)](https://github.com/simonmittag/ingress-j8a/pulls?q=is%3Aopen+is%3Apr)
[![Github Issues](https://img.shields.io/github/issues/simonmittag/ingress-j8a)](https://github.com/simonmittag/ingress-j8a/issues)
[![Github Activity](https://img.shields.io/github/commit-activity/m/simonmittag/ingress-j8a)](https://img.shields.io/github/commit-activity/m/simonmittag/ingress-j8a)  
[![Go Report](https://goreportcard.com/badge/github.com/simonmittag/ingress-j8a)](https://goreportcard.com/report/github.com/simonmittag/ingress-j8a)
[![Go Version](https://img.shields.io/github/go-mod/go-version/simonmittag/ingress-j8a)](https://img.shields.io/github/go-mod/go-version/simonmittag/ingress-j8a)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Why?
A kube ingress controller for [j8a](https://github.com/simonmittag/j8a).


# What?
ingress-j8a is a kubernetes ingress controller, exposing ports 80, 443 on a j8a ingress-controller resource
inside the cluster. Its role is to farm traffic to kubernetes services and their associated pods. 
The project is currently under construction and there is currently no public release.

![](art/ingress-j8a.png)

## Goals
* Zero downtime for j8a during updates to ingress resources.
* Redundancy for J8a with multiple instances and a load balancing mechanism
* intelligent defaults for j8a for proxy server params the kubernetes ingress resource does not readily expose.



## How it works 
The basic mechanics of monitoring kubernetes for configuration changes,
then updating J8a's config and it's live traffic routes.

![](art/ingress-j8a-mechanics.png)
1. The user deploys ingress resources to the cluster, or updates them. This also needs to cater for configMap and secrets updates
2. A cache that runs inside ingress-j8a monitors for updates to kube resources in all namespaces. it versions the config.
3. the control loop that continuously waits for config changes is notified.
4. the control loop reads the config out and generates a j8a config object in yml. this will be deployed to the kube cluster as its own configmap object. (We will probably need our own namespace) r i
5. ingress-j8a then tells the kube api server about j8a-config as a configMap
6. kube api server deploys this resource into the cluster into our own namespace. 
7. ingress-j8a then tell kube api server to deploy the latest docker image of j8a into the cluster. 
8. kube api-server creates the deployment. Several problems need to be solved here. 
   * It will need to be configured from the configmap. 
   * it needs to run on some kind of nodeport config on each node? listening on the same port on every node. 
   * we need it's external IP address
   * we may need to create an external NLB for it? (how would we even know about this?)


# Contributions

The ingress-j8a team welcomes all [contributors](https://github.com/simonmittag/ingress-j8a/blob/master/CONTRIBUTING.md). Everyone
interacting with the project's codebase, issue trackers, chat rooms and mailing lists is expected to follow
the [code of conduct](https://github.com/simonmittag/ingress-j8a/blob/master/CODE_OF_CONDUCT.md)
