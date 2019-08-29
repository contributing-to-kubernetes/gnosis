# Kind-ly running Contour

This is based on the great post by Steve Sloka,
[Kind-ly running Contour](https://projectcontour.io/kindly-running-contour/).

I used kind v0.6.0-alpha and Contour 0.15.0 for this tutorial.

## Creating a Kubernetes cluster

First off, let's create the cluster:
```
$ kind create cluster --config ./contour/kind.config.yaml
```

The kind config file will create 1 worker node and we are mapping ports 80 and
443 between the worker node and our host.

You should see the worker node come up as a Docker container:
```
ONTAINER ID        IMAGE                  COMMAND                  CREATED              STATUS              PORTS                                      NAMES
aef85177177f        kindest/node:v1.15.3   "/usr/local/bin/entr…"   About a minute ago   Up About a minute   0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp   kind-worker
```

## [heptio/contour]

### Envoy

Let's take a short detour into Envoy before we proceed with Contour.
Hopefully some details will make more sense...

So Envoy...

One of the big features that envoy provides is a set of management APIs.
These management APIs can very easily be implemented by configuration servers
(i.e., Contour) or control plane services (i.e., Istio).
If the control plane / configuration server implements the set amnagement APIs
then it is possible to manage Envoy configuration dynamically (no need to
restart Envoy for changes to occur).

In the Envoy world, specifically looking at the v2 management APIs, these can
do endpoint discovery, cluster discovery, route discovery, listener discovery,
health discovery, aggregated discovery, and secret discovery services.
These APIs are often refered to as xDS (x Discovery Service) APIs.
See [The universal data plane API] for more detail.

### Contour

> Contour is a Kubernetes ingress controller using Lyft's Envoy proxy. https://projectcontour.io

> Contour is an Ingress controller for Kubernetes that works by deploying the Envoy proxy as a reverse proxy and load balancer. Unlike other Ingress controllers, Contour supports dynamic configuration updates out of the box while maintaining a lightweight profile.

In this blog post, we will us a ["split" deployment].
This sort of deployment separates Contour and Envoy resources, versus deploying
them in the same Pod and having them communicate through `localhost`.

We will use one of the examples from the [heptio/contour] repo:
```
# Clone the repo...
$ mkdir ~/go/src/github.com/heptio
$ cd ~/go/src/github.com/heptio
$ git clone https://github.com/heptio/contour.git

# Deploy Contour
$ kubectl apply -f ~/go/src/github.com/heptio/contour/examples/ds-hostnet-split
```

## Deploy a Sample Application

Now that we have Contour deployed, we should have all requests to
`localhost:80` and `localhost:443` be routed to an Envoy pod.
At this point we need an application to which Envoy will route these requests
to.

The [heptio/contour] contains the resources necessary to deploy [kuard]:
```
$ kubectl apply -f ~/go/src/github.com/heptio/contour/examples/example-workload/kuard-ingressroute.yaml
```

**NOTE**: You'll need to add `127.0.0.1 kuard.local` to `/etc/hosts` in order to
access the sample application through your web browser at
`http://kuard.local/`.

### What Just Happened?

The last kubectl apply command should have created the following Kubernetes
resource's:
* `deployment.apps/kuard`
* `service/kuard` - ClusterIP type Service
* `ingressroute.contour.heptio.com/kuard`

`ingressroute.contour.heptio.com/kuard` looks like this:
```yaml
apiVersion: contour.heptio.com/v1beta1                                          
kind: IngressRoute                                                              
metadata:                                                                       
  labels:                                                                       
    app: kuard                                                                  
  name: kuard                                                                   
  namespace: default                                                            
spec:                                                                           
  virtualhost:                                                                  
    fqdn: kuard.local                                                           
  routes:                                                                       
    - match: /                                                                  
      services:                                                                 
        - name: kuard                                                           
          port: 80
```


[heptio/contour]: https://github.com/heptio/contour
[The universal data plane API]: https://blog.envoyproxy.io/the-universal-data-plane-api-d15cec7a
["split" deployment]: https://projectcontour.io/contour-v014/
[kuard]: https://github.com/kubernetes-up-and-running/kuard
[Contour IngressRoute]: https://github.com/heptio/contour/blob/v0.14.2/docs/ingressroute.md#ingress-to-ingressroute
