# Why Kubernetes

If you found yourself in here then you are either interested in getting an
overview of Kubernetes of you clicked some random linkin Twitter.
Either way, here, we will discuss why Kubernetes looks the way it does.

## Problem Statement

First thing you see when you land in
[k/k](https://github.com/kubernetes/kubernetes)
(the https://github.com/kubernetes/kubernetes repo) is this line:

> Production-Grade Container Scheduling and Management

The big idea is that you can write portable programs (portable in linux-like
systems, at least) by packagin them in a container image
(the thing that comes out of a `docker build`) and subsecuently using that
container image to run a container (the thing that happens when you `docker run`).

Maybe you are familiar with
[docker compose](https://docs.docker.com/compose/) and have some ecperience
orchestrating containers.
If so, you know that sometimes you have to restart docker compose, if your
machine goes down.
Also, you dont get to scale out to multiple machines easily.
This was one of the reasons why [Docker Swarm](https://github.com/docker/swarm)
was invented.

There are plenty of tools out there to do container orchestration!
One that a lot of people may be familiar with as well is
[Fargate](https://aws.amazon.com/fargate/).

> AWS Fargate is a compute engine for Amazon ECS that allows you to run
> containers without having to manage servers or clusters. With AWS Fargate,
> you no longer have to provision, configure, and scale clusters of virtual
> machines to run containers. This removes the need to choose server types,
> decide when to scale your clusters, or optimize cluster packing. AWS Fargate
> removes the need for you to interact with or think about servers or clusters.
> Fargate lets you focus on designing and building your applications instead of
> managing the infrastructure that runs them.

Sounds pretty damn good, doesn't it?
The big thing that we want to offer people is the capacity to develop and
deploy their applications in a reliable and scalable platform that doesn't
require the developer to know squat about the machines, networking, or anything
of the sort.
The motivation is that developers should be free to do what they need, when
they need.

And like all the container orchestration platforms, Kubernetes aims to do this.
However, here is where the design of Kubernetes really shines.
In Kubernetes, container orchestration doesn't solely imply that your
containers will be kept alive and running.
It also means that any service that your application needs will be available
and managed by the same sort of APIs that manage your application!

This may not sound like much but let's take an example.
If you want to horizontally scale your application (increase the number of
replicas), in Fargate, you need to set up an "app autoscaling target", a couple
"autoscaling policies", along with some "cloudwatch alarms", see
[`bradford-hamilton/terraform-ecs-fargate`](https://github.com/bradford-hamilton/terraform-ecs-fargate/blob/master/terraform/auto_scaling.tf)
for an example.

In Kubernetes, there is an API that is designed to help you out.

```yaml
apiVersion: autoscaling/v1
kind: HorizontalPodAutoscaler
metadata:
  name: php-apache
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: php-apache
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 50
```

This holds for the most part, a lot of container orchestration solutions rely
on the cloud provider to work, and you have to twiddle with those bits to get
your container to do what you want.
In Kubernetes, applications, infrastructure, and cloud providers are kept away
from one another (unless you build an operator to manage your cloud provider's
infrastructure!).
This is where the thing really shines, and this is why so many people even
bother.

Most container orchestration platforms orchestrate the container and period.
Kubernetes orchestrates everything that your application may need to be useful.
