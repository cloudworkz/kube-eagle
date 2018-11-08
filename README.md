# Kube eagle

Kube eagle is a prometheus exporter which exports various metrics of kubernetes pod resource requests, limits and it's 
actual usages. It was created with the purpose to provide a better overview of your kubernetes cluster resources, so that
you can optimize the resource allocation.

## Setup

Simply deploy a pod which runs kube-eagle inside the kubernetes cluster you would like to monitor. Make sure the pod has a
service account attached that has read-only access to the cluster.

### Environment variables

| Environment variable | Description                                                                           | Default |
| -------------------- | ------------------------------------------------------------------------------------- | ------- |
| PORT                 | Port to listen on for the prometheus exporter                                         | 8080    |
| IS_IN_CLUSTER        | Whether to use in cluster communication or to look for a kubeconfig in home directory | true    |

## How does it work

Kube eagle talks to the kubernetes api server using the official kubernetes go client. Every 10s we perform 4 requests - pod
& node resource objects as well as the pod & node usage list. We aggregate and enrich some of the data so that you can easily
build dashboards matching the purpose of optimizing the resource allocation.
