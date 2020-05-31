# Using MetalLb with Kind

This is my attempt at running step by step through Duffie Cooley's
[Using MetalLb with Kind] blog post.

I used kind v0.6.0-alpha for this tutorial.

Create your cluster:
```
$ kind create cluster --config=./kind.config.yaml
```

Try and ping the worker nodes in the cluster:
```
$ kubectl  get no -o wide
NAME                 STATUS   ROLES    AGE     VERSION   INTERNAL-IP   EXTERNAL-IP   OS-IMAGE                                  KERNEL-VERSION     CONTAINER-RUNTIME
kind-control-plane   Ready    master   3m22s   v1.15.3   172.17.0.3    <none>        Ubuntu Disco Dingo (development branch)   5.0.0-25-generic   containerd://1.2.6-0ubuntu1
kind-worker          Ready    <none>   2m44s   v1.15.3   172.17.0.2    <none>        Ubuntu Disco Dingo (development branch)   5.0.0-25-generic   containerd://1.2.6-0ubuntu1
(base)
```

Ping the worker node:
```
$ ping 172.17.0.2
PING 172.17.0.2 (172.17.0.2) 56(84) bytes of data.
64 bytes from 172.17.0.2: icmp_seq=1 ttl=64 time=0.085 ms
64 bytes from 172.17.0.2: icmp_seq=2 ttl=64 time=0.099 ms
^C
--- 172.17.0.2 ping statistics ---
2 packets transmitted, 2 received, 0% packet loss, time 25ms
rtt min/avg/max/mdev = 0.085/0.092/0.099/0.007 ms
```

Inspect the Docker `bridge` network:
```
$ docker network inspect bridge
```

Specifically, we are looking for the network's IP address management
configuration:
```
$ docker network inspect bridge --format='{{json .IPAM.Config}}'
[{"Subnet":"172.17.0.0/16","Gateway":"172.17.0.1"}]
```

Based on the subnet, Docker's `bridge` network has the CIDR range of
172.17.0.0 - 172.17.255.255.

Create a sample application, an "echo" server based on [InAnimaTe/echo-server]:
```
$ kubectl apply -f sample-app.yaml
```

You should see a Deployment and a Service:
```
$ kubectl get deploy,svc echo
NAME                         READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/echo   3/3     3            3           13s

NAME           TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
service/echo   LoadBalancer   10.104.255.80   <pending>     8080:32445/TCP   13s
```

Notice that the Service has the external IP marked as `<pending>`.

From [MetalLB's installation instructions]:
```
kubectl apply -f https://raw.githubusercontent.com/google/metallb/v0.8.1/manifests/metallb.yaml
```

The following ConfigMap will give MetalLB control over the IPs from
172.17.255.1 to 172.17.255.250, and configure Layer 2 mode:
```
$ kubectl apply -f metallb-config.yaml
configmap/config created
```

At this point, you should see the echo Service get an external IP:
```
$ kubectl  get svc echo
NAME   TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)          AGE
echo   LoadBalancer   10.104.255.80   172.17.255.1   8080:32445/TCP   10m
```

Test it out:
```
curl http://172.17.255.1:8080
```

or open go to http://172.17.255.1:8080/ws in your web browser :smile:


## Notes

### Layer 2

Let's start with some context on what layer 2 mode means:

> Layer 2, also known as the Data Link Layer, is the second level in the
> seven-layer OSI reference model for network protocol design. Layer 2 is
> equivalent to the link layer (the lowest layer) in the TCP/IP network model.
> Layer2 is the network layer used to transfer data between adjacent network
> nodes in a wide area network or between nodes on the same local area network.
> [layer 2 mode]

> In layer 2 mode, one machine in the cluster takes ownership of the service,
> and uses standard address discovery protocols (ARP for IPv4, NDP for IPv6)
> to make those IPs reachable on the local network. From the LAN’s point of
> view, the announcing machine simply has multiple IP addresses.
> [MetalLB docs]

The machine that takes ownership of a service is a "leader".
MetalLB will rely on Kubernetes to figure out the state of pods and nodes
relevant to the service. This information will be used to select a leader and
to act in the case that the leader goes away.

> layer 2 does not implement a load-balancer. Rather, it implements a failover
> mechanism so that a different node can take over should the current leader
> node fail for some reason. [MetalLB layer 2 docs].


[Using MetalLb with Kind]: https://mauilion.dev/posts/kind-metallb/
[InAnimaTe/echo-server]: https://github.com/InAnimaTe/echo-server
[MetalLB's installation instructions]: https://metallb.universe.tf/installation/
[layer 2 mode]: https://www.juniper.net/documentation/en_US/junos/topics/concept/l2-qfx-series-overview.html
[MetalLB docs]: https://metallb.universe.tf/concepts/
[MetalLB layer 2 docs]: https://metallb.universe.tf/concepts/layer2/
