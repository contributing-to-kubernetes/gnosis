# Running Kubernetes E2E Tests

Before we get started, I have to give credit to the wonderful people that have
worked on kind and on kind-based CI in Kubernetes since this README is based on
the things I learned while going though
https://github.com/kubernetes-sigs/kind/blob/master/hack/ci/e2e-k8s.sh

So @BenTheElder and @liggitt (authors of `e2e-k8s.sh`), thank you!
And thank you to all the contributors to Kind and Kubernetes - knowledge
doesn't just materialize out of a vacuum but advances step by step until it
reaches us and pushes us a bit forward.
Every contribution to the community, no matter how small has helped us get to
where we are today :evergreen_tree:.

## Overview

Check this out: https://testgrid.k8s.io/sig-release-master-blocking#conformance-ga-only

In Kubernetes there are 2 types of e2e tests: presubmits (the ones that are run
on each PR) and periodics (tests that periodically test Kubernetes).

The above shows you an e2e job.
An e2e job is a
[Prow](https://github.com/kubernetes/test-infra/tree/master/prow)
job that executes one or more e2e tests.

The e2e tests themselves can be found in k/k under the
[test/e2e](https://github.com/kubernetes/kubernetes/tree/master/test/e2e) directory.

All these tests are compiled into an executable that can later be run to test
Kubernetes.
You can try it out yourself by cloning Kubernetes and running the following
command at the root level of the repository

```
make all WHAT='test/e2e/e2e.test'
```

If successful, you should see a binary in
`./_output/local/bin/linux/amd64/e2e.test`.
This binary is what we can run to execute e2e tests.

Here is the kicker: Kubernetes e2e tests are written using
https://github.com/onsi/ginkgo.
To run them, we need to have ginkgo installed.
And because this is such a necessity in the Kubernetes world, there is a way of
doing it by running the following command:

```
make all WHAT='vendor/github.com/onsi/ginkgo/ginkgo'
```

**Note** you can combine the previous 2 commands into

```
make all WHAT='test/e2e/e2e.test vendor/github.com/onsi/ginkgo/ginkgo'
```

That will compile the `e2e.test` and `ginkgo` binaries.


## Setting Up a Cluster

The simplest way to get a cluster up and running is with
https://github.com/kubernetes-sigs/kind.

Follow the instructions in its docs to install it and lets go ahead and create
a cluster.
By now, you should already have a clone of Kubernetes, place the code in
`$GOPATH/src/k8s.io/kubernetes/` (although kind has a the `--kube-root` flag
that you can use to point to your clone of Kubernetes).

But let's go ahead and create a kind `node-image` (a container image used to
run Kubernetes nodes)

```
kind build node-image --type docker
```

kind implements to build types, bazel and docker.

The above will take a bit but eventuallyyyyy, you should see a new container
images named `kindest/node:latest`.

Now that we have our Kubernetes version built, let's create a cluster with it.
Begin by copy-and-pasting the following yaml snippet into a file (i.e.,
`test-run.yaml`):

```yaml
# config for 1 control plane node and 2 workers.
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  ipFamily: ipv4
nodes:
- role: control-plane
- role: worker
- role: worker
kubeadmConfigPatches:
- |
  kind: ClusterConfiguration
  metadata:
    name: config
  apiServer:
    extraArgs:
      runtime-config: ""
```

and do

```
kind create cluster --image kindest/node:latest --config test-run.yaml --wait 5m
```

## Running an E2E test

Promise this is going to make sense in a bit but right now just trust me and
let's get a URL for your control plane by running the following:

```
CONTROL_PLANE_URL=$(kubectl config view -o jsonpath="{.clusters[?(@.name == \"kind-kind\")].cluster.server}")
```
The above command gets the value for the cluster.server field for the cluster
named `kind-kind`.
In my case this was `https://127.0.0.1:46033`.

```
$GOPATH/src/k8s.io/kubernetes/_output/bin/ginkgo $GOPATH/src/k8s.io/kubernetes/_output/bin/e2e.test -- \
  --kubeconfig=$HOME/.kube/config --host="$CONTROL_PLANE_URL" --provider=skeleton \
  --ginkgo.focus="Downward API volume should provide podname only"
```

The magic here is that the e2e.test binary is a precompiled version of
Kubernetes tests.
Ginkgo uses it to essentially run `go test` behind the scenes.
All the flags we used mingle with our Kubernetes e2e test binary to run some Go
code.
For example, the test we ran can be (currently) found here:
https://github.com/kubernetes/kubernetes/blob/master/test/e2e/common/downward_api.go.


## Running Your Own Tests :smile:

Time for us to understand Kubernetes e2e tests better.

If you made it to this part of the readme, you should have an `e2e.test`
binary.
This binary is made up of tests, check this part of the ginkgo docs for the
tl;dr https://onsi.github.io/ginkgo/#precompiling-tests.

The Kubernetes `e2e.test` binary is compiled from this place in k/k (Kubernetes
repo) https://github.com/kubernetes/kubernetes/tree/master/test/e2e.
Just for you, we added a smaller version of it here for you to play around in
[`example-e2e-test`](./example-e2e-test/) (promise to add some non-trivial e2e
tests in the future).

You can run it in exactly the same manner as the Kubernetes `e2e.test` binary.

First off, you will need to build a test binary.
You need to first of all install ginkgo following the official docs here
http://onsi.github.io/ginkgo/.
Once you have ginkgo installed in your machine you can go into
[`example-e2e-test`](./example-e2e-test/)
and run

```
ginkgo build
```

This should result in your own test binary `example-e2e-test.test`.

If you want to see the available command-line flags run

```
ginkgo example-e2e-test.test -- --help
```

To run it, then try

```
ginkgo example-e2e-test.test -- --test-flag cool-framework --ns ci
```
