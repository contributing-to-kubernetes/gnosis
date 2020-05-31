# Overview of Kubernetes

Let's get to it!

It is possible that you may have heard about Pods, Deployments, Configmaps, and
the like.
If you haven't, then you are about to! :smile:

**Note**: the examples we reference here can also be found in the
[examples](./examples) sudirectory.

## Pods
By now we know that Kubernetes is a "Production-Grade Container Scheduling and
Management".
Now, we have to figure out what to put in containers...

Containers are designed for you to run 1 process in them.
They are not entire machines!
They should not ron daemonized applications, you shouldn't have nginx sending
over reuests to your web app, or anything of the sort.
You get one process so use it wisely!

However, you may need to have multiple processes working together.
Have you heard of Istio, Linkerd, or any of the other tens of service meshes?
What they do, is that they run a sidecar container with your "application".
This sidecar container may handle networking, telemetry, what have you.
Thus you get to focus on your app and the othe container that runs with your
app will help you out (no need to bake additional logic into your application).

To enable this pattern, Kubernetes packages one or more containers in a `Pod`.
(If you are familiar with ECS/Fargate you may have workedwith tasks).

The most common way of running thins in Kubernetes is by using YAML.
Time for an example!
Let's run a container that uses ubuntu 18.04:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - image: ubuntu:18.04
    name: bash
    args:
    - "sleep"
    - "999"
```

Now, create the pod!
```bash
kubectl apply -f examples/pod.yaml
```

You can look at the pod by running
```bash
kubectl get pods
```

and delete it by running
```bash
kubectl delete -f examples/pod.yaml
```

## ReplicaSets

tl;dr: pods package applications.

Another thing you may see or hear every now and then are `ReplicaSets`.
ReplicaSets help you scale your app.
ReplicaSets have a similar structure to Pods:
```yaml
apiVersion: ...
kind: ...
metadata:
  ...
spec:
  replicas: N
  selector:
    matchLabels:
      ...
  template:
```

The `apiVersion`, `kind`, `metadata`, and `spec` are in almost all Kuberentes
resources, so this you will see most of the time.
Within the `spec` field, however you see some new fields:
* `replicas`: specifies the number of replicas of a given pod that you want to
  run.
* `selector.matchLabels`: like many objects in Kubernetes, ReplicaSets are
  managed by a Kubernetes component (a "controller"). This
  controller is an application, and this application finds pods
  through "labels" (labels are attached to Kubernetes resources, they are
  essentially more metadata). So in the `selector.matchLabels` field you tell
  Kubernetes how the Pods for this ReplicaSet will be labeled.
* `template`: this is where you specify the pod template (or pod spec!!) that the replicaset is
  to use to create pods!

Let's now create a replicaset that runs 2 replicas of the pod in the previous
section.
The replicaset definition will now look like this:
```yaml
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: my-rs
spec:
  replicas: 2
  selector:
    matchLabels:
      app: ubuntu
  template:
    metadata:
      labels:
        app: ubuntu
    spec:
      containers:
      - image: ubuntu:18.04
        name: bash
        args:
        - "sleep"
        - "999"
```

Create it
```bash
kubectl apply -f examples/rs.yaml
```

Inspect it
```
kubectl get replicasets
```

Inspect its pods
```
kubectl get pods --show-labels
```

And delete it
```
kubectl delete -f examples/rs.yaml
```

## Deployments

tl;dr pods package applications and replicasets allow you to scale your pods.

Deployments are where it is at! :rocket:
Deployments are an abstraction on top of ReplicaSets; Deployments manage
ReplicaSets!
So for the most part, you want to be writing Deployments instead of
ReplicaSets.

We previously mentioned that replicasets manage pod replicas.
Well, deployments manage replicasets!

There are many features that Deployments bring to the table but the number one
is the ability to manage version updates.
what if you want to bump your app to its latest version?
If you already have a deployment running, then that deployment will have a
replicaset, and that replicaset will be "managing" some pods.
If you now update your deployment to use a new container image version then the
deployment will create a new replicaset, and this new replicaset will then create
pods based on the newer version.
As the new pods come to live, the old replicaset will slowly decresed its
desired replica count, thus killing the old pods until we replace all the old
pods with ones running the latest version.

And now time for another example!
A deployment looks very much like a replicaset.
You can compare this to our previous example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deploy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ubuntu
  template:
    metadata:
      labels:
        app: ubuntu
    spec:
      containers:
      - image: ubuntu:18.04
        name: bash
        args:
        - "sleep"
        - "999"
```

You can create the deployment
```bash
kubectl apply -f examples/deploy.yaml
```

You can inspect the deployment
```bash
kubectl get deployments
```

You can inspect its replicaset
```bash
kubectl get replicasets --show-labels
```

Its pods
```bash
kubectl get pods --show-labels
```

and to clean it up
```
kubectl delete -f examples/deploy.yaml
```
