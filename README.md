# OpenShift cross cluster load balancer
A tcp load balancer that is aware of multiple OpenShift clusters and their exported routes. Also uses pod filters to determine where a HA-Proxy is running.

## Screenshot
![Image of the UI](https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/blob/master/img/ui.png)

## Main idea
In a large, high load production OpenShift cluster changes to cluster itself impose a huge risk.
During two years of OpenShift operating experience the idea of a rolling update model for the hole OpenShift cluster formed.
Basically this works the same way the rolling update of an application on OpenShift works.
But instead of creating new containers, you create a hole new OpenShift cluster and automatically migrate the applications based on any selectors.

![Rolling update of the cluster](https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/blob/master/img/rollingClusterUpdate.png)

In our case we would like to move less important projects to the new cluster, test all operations in productive workload and then migrate the important apps.
As everything is containerized and based on yaml configuration files, the migration of applications is quite easy.

For us the missing part is an external load balancer that is aware of multiple OpenShift clusters and can balance to those based on which applications are exposed as routes.

This repo, the OpenShift "smart" load balancer fills that gap.

## Master thesis
The load balancer was created as a prototype for my thesis during my masters degree. The documentation can be found here:


## Smart load balancer components
![Components of smart load balancer](https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/blob/master/img/internalArchitecture.png)

## High level overview
![High level overview](https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/blob/master/img/architectureOverview.png)

## Setup
- You need two OpenShift clusters with HA-Proxies as routers
- Install and run the smart load balancer on a server
- Then install the smart load balancer plugin as a container on OpenShift:

```bash
# todo
# Permissions for service-account

# on cluster 1
oc new-app retocode/origin:v2 --name=smart-lb-plugin -e SMART_LB_API_URLS=http://<url-of-balancer>:8089 -e CLUSTER_KEY=ose1

# on cluster 2
oc new-app retocode/origin:v2 --name=smart-lb-plugin -e SMART_LB_API_URLS=http://<url-of-balancer>:8089 -e CLUSTER_KEY=ose2
```

- Check the UI on the smart load balancer http://<url-of-balancer>:8089
