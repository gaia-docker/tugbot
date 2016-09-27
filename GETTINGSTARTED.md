# Tugbot - Getting Started 
# THIS SECTION IS UNDER CONSTRUCTION!!!

*Tugbot* is Testing in Production (TiP) framework for docker base enviroments. Currently supports Kubernetes, Swarm and a single docker host.
This guild will take you step by step of what you need to do in order to configure your cluster for running TiP to analyze your test results.

## Kubernetes
1. Install Kubernetes. You can run *tugbot* on [minikube](http://kubernetes.io/docs/getting-started-guides/minikube/) 
for start a single node kubernetes cluster locally for purposes of development and testing.
2. Create a kubernetes "test job". A "test job" is a regular job that include additional information, 
so *tugbot* will discover it in runtime and rerun it upon events.
For example, each time there is a deployment, *tugbot* can run a job that will validate service qulety.
Follow job contains `tugbot.kubernetes.events: ReplicaSet.SuccessfulCreate` label, 
which asks *tugbot* to run this job each time there is a succesful deployment. 
See example test job yaml in [tugbot-kube](https://github.com/gaia-docker/tugbot-kubernetes/blob/master/README.md).
3. Run *tugbot-kube* and connect it to kubernetes manster node. 
*Tugbot-Kube* listen to kubernetes master events and running jobs upon those events.
The easiest way is to run *tugbot-kube* inside a Docker container, see [tugbot-kube](https://github.com/gaia-docker/tugbot-kubernetes/blob/master/README.md).
4. Run [*tugbot-collect*](https://github.com/gaia-docker/tugbot-collect) & *tugbot-result*. 
*tugbot-collect* collects exited docker test containers results and publish it using *tugbot-result*.
Currently there are 3 options for publish &  analyze test results: 
live stream (a web UI) - [*tugbot-result*](https://github.com/gaia-docker/tugbot-result-service),
elasticsearch - [*tugbot-result-es*](https://github.com/gaia-docker/tugbot-result-service-es) & 
[*gaia*](https://master.gaiahub.io).
Use follow *tugbot-collect* pod that should be deployed on all kubernetes nodes that running test containers:

TODO - add yaml files for deployment of collect and result services

## Swarm
TODO

## Single Docker Host
TODO
