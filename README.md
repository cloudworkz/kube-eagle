# Kube eagle

<!-- prettier-ignore -->
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/google-cloud-tools/kube-eagle/master/LICENSE)
[![Docker Repository on Quay](https://quay.io/repository/google-cloud-tools/kube-eagle/status "Docker Repository on Quay")](https://quay.io/repository/google-cloud-tools/kube-eagle)

Kube eagle is a prometheus exporter which exports various metrics of kubernetes pod resource requests, limits and it's actual usages. It was created with the purpose to provide a better overview of your kubernetes cluster resources, so that you can optimize the resource allocation. You can easily build, or use our default grafana dashboard which will help you to achieve this goal:

![Grafana Dashboard for Kubernetes resource monitoring](https://raw.githubusercontent.com/google-cloud-tools/kube-eagle/master/grafana-sample.png)

## Setup

Simply deploy a pod which runs kube-eagle inside the kubernetes cluster you would like to monitor. We recommend using our provided helm chart to deploy kube eagle in your cluster:

Kube eagle helm chart: https://github.com/google-cloud-tools/kube-eagle-helm-chart

### Required permissions

Make sure the pod has a service account attached that has the required permissions. You can use our helm chart which is capable of creating the service account along with the required ClusterRole and ClusterRoleBinding.

### Environment variables

| Variable name | Description                                                                           | Default |
| ------------- | ------------------------------------------------------------------------------------- | ------- |
| PORT          | Port to listen on for the prometheus exporter                                         | 8080    |
| IS_IN_CLUSTER | Whether to use in cluster communication or to look for a kubeconfig in home directory | true    |

### Grafana dashboard

Work in progress

## How does it work

Kube eagle talks to the kubernetes api server using the official kubernetes go client. Every 10s we perform four requests - pod
& node resource objects as well as the pod & node usage list. We aggregate and enrich some of the data so that you can easily
build dashboards matching the purpose of optimizing the resource allocation.

## License

MIT License

Copyright (c) 2018

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
