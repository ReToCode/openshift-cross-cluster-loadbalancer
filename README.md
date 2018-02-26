# OpenShift cross cluster load balancer
This is a tcp load balancer that is aware of multiple OpenShift clusters and their exported routes. Also uses pod filters to determine where a HA-Proxy is running. It is a prototype that was created during my masters thesis to prove an idea to update productive OpenShift clusters without any downtime and risk. The full thesis can be found here: [Master Thesis](https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/blob/master/MT_Reto_Lehmann_OpenShiftSmartCrossClusterLoadbalancing.pdf)

## Screenshot
![Image of the UI](https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/blob/master/img/ui.png)

## Main idea
In a large, high load production OpenShift cluster, changes to cluster itself impose a huge risk.
During two years of OpenShift operating experience the idea of a rolling update model for the hole OpenShift cluster formed.
Basically this works the same way the rolling update of an application on OpenShift works,  but instead of creating new containers, you create a hole new OpenShift cluster and automatically migrate the applications based on any selectors.

![Rolling update of the cluster](https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/blob/master/img/rollingClusterUpdate.png)

In our case we would like to move less important projects to the new cluster, test all operations in productive workload and then migrate the important apps.  As everything is containerized and based on yaml configuration files, the migration of applications is quite easy.

For us the missing part is an external load balancer that is aware of multiple OpenShift clusters and can balance to those based on which applications are exposed as routes.

This repo, the OpenShift "smart" load balancer fills that gap.

## Smart load balancer components
![Components of smart load balancer](https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/blob/master/img/internalArchitecture.png)

## High level overview
![High level overview](https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/blob/master/img/architectureOverview.png)

## Setup
- You need two OpenShift clusters with HA-Proxies as routers
- Install and run the smart load balancer on a server
- Then install the smart load balancer plugin as a container on OpenShift:

```bash
# Prerequisite two OpenShift clusters with HA-Proxies

# Install the smart load balancer somewhere
# Get the binary from: https://github.com/ReToCode/openshift-cross-cluster-loadbalancer/releases
./openshift-cross-cluster-loadbalancer

# Deploy the plugin on multiple clusters
oc new-app retocode/origin:v3 --name=smart-lb-plugin -e SMART_LB_API_URLS=http://<ip-of-smart-lb>:8089 -e CLUSTER_KEY=openshift-1
oc new-app retocode/origin:v3 --name=smart-lb-plugin -e SMART_LB_API_URLS=http://ip-of-smart-lb:8089 -e CLUSTER_KEY=openshift-2

# Deploy some sample apps or use your own
# Cluster 1
oc project myproject
oc new-app retocode/node-sample:v2 —name=node-sample -e CLUSTER='OpenShift Cluster 1'
oc expose service node-sample --hostname=ose1-direct-myproject.192.168.99.100.nip.io
oc expose service node-sample --hostname=myapp.mydomain.com --name=shared-route
oc expose service node-sample --hostname=myapp-migrate.mydomain.com --name=shared-route-weight
oc annotate route shared-route-weight smartlb-weight='10'

# Cluster 2
oc project myproject
oc new-app retocode/node-sample:v2 --name=node-sample -e CLUSTER='OpenShift Cluster 2’
oc expose service node-sample --hostname=ose2-direct-myproject.192.168.99.103.nip.io
oc expose service node-sample --hostname=myapp.mydomain.com --name=shared-route
oc expose service node-sample --hostname=myapp-migrate.mydomain.com --name=shared-route-weight
oc annotate route shared-route-weight smartlb-weight='20'

# Check the UI on the smart load balancer http://<ip-of-smart-lb>:8089

# Get the apps directly
curl ose1-direct-myproject.192.168.99.100.nip.io
curl ose2-direct-myproject.192.168.99.103.nip.io

# Get the apps balanced
curl localhost:8080 -H 'Host: myapp.mydomain.com'

# Get the apps during a migration from one cluster to the other (weighted)
curl http://localhost:8080 -H 'Host: myapp-migrate.mydomain.com' 
```

