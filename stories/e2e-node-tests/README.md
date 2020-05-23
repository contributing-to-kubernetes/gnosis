# Running E2E Node Tests

Welcome to another adventure down the Kubernetes test directory.
Last time we cover
[how to run Kubernetes e2e tests with kind](../e2e-tests/), this time we will
cover node e2e tests.
These tests are owned by
[SIG Node](https://github.com/kubernetes/community/tree/master/sig-node).

Node e2e tests are interesting because they test components such as the Kubelet
and they do so without much else than ETCD.
This post will be the first in a series of SIG Node tests as we figure out a
reliable and easy-ish way to run these tests locally.
For this post, we will cover how to run them remotely in GCP.

## Getting Started

### Configuring Your CLI
As we go down this rabbit hole, we will begin by setting up our environment for
remote tests.

To get started, please go through
https://github.com/kubernetes/community/blob/master/contributors/devel/sig-node/e2e-node-tests.md
and make sure to install gcloud, for example
https://cloud.google.com/deployment-manager/docs/step-by-step-guide/installation-and-setup

To see if your account is all set do a
```
gcloud auth list
```

Else, run a `gcloud config set account <my-cool-self@gmail.com>`.
To chose a project, do a `gcloud config set project <project ID>`.

GCP, like other cloud providers has regions and zones.
You can browse and chose anyone you want but what matters is that you chose
one.
To do so, you will need to run a command that looks like this

```
gcloud config set compute/region us-central1
gcloud config set compute/zone us-central1-f
```

### Preparing Your VMs

In GCP, in order to SSH into a machine one can add a public SSH key as part of
the metadata associated with VMs.
The "metadata" section is a section in the Google cloud console in the "Compute
Engine" section.

So we will now go and add an SSH key for us to be able to SSH into our logins -
this is used by the official e2e test node runner in order to "upload" binaries
and to "download" test results and logs.

To add an SSH public key please see this guide:
https://cloud.google.com/compute/docs/instances/adding-removing-ssh-keys

You can add your own key (you will need to format it as stated in the link
above) or you can use this gcloud commmand which will generate a prvate and
public SSH keys in your machine and will add the public one as metadata for
your VMs

```
gcloud compute config-ssh
```

Ladt step before we exectute an actual test is to obtain the credentials for a
service account for the node test runner to use.
For that, please follow this guide:
https://cloud.google.com/docs/authentication/production#cloud-console

Once you have it, download it and put it somewhere.
You will need to set the following environment variable for the node test
runner to use:
```
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-credentials.json
```

## Running Node Tests

At this point, you should have a terminal open and be at the root of the
Kubernetes repo.
If are there and if you have gone through the steps above then you should be
able to execute this command

```
time make test-e2e-node FOCUS="\[Flaky\]" SKIP="" PARALLELISM=1 REMOTE=true DELETE_INSTANCES=true
```

This will run Ginkgo tests labeled as "Flaky", it will run them serially, and
it will delete the VM instance after the e2e tests finish.
(These tests will also probably fail but we will tell you why and how to make
them pass later on :microscope:).

But, if you wanna see the entire set of arguments and all run a
```
make test-e2e-node PRINT_HELP=y
```

## Into the Rabbit Hole

Ight, so you just ran a node e2e test.
Running any other e2e test is just a matter of figuring out what to `FOCUS` on
and what to `SKIP`.

So now we will try and see what all is going on behind the scenes - this is the
long section :smile:.

Like normal e2e tests, the make command will compile Ginkgo from the vendor
directory (along with some other dependencies) and it will compile a test
binary from the source code at
https://github.com/kubernetes/kubernetes/tree/master/test/e2e_node.
The bash that does all this for us can be found in
https://github.com/kubernetes/kubernetes/blob/master/hack/make-rules/test-e2e-node.sh.
In that file you will see how parallelism is determined, where the runtime is
specified, ginkgo flags, etc.

The thing you should notice is this command
```bash
go run test/e2e_node/runner/remote/run_remote.go
```

The e2e node test runner has a directory for local tests and another one for
remote tests.
For the sake of experimentation we will only focus here on the remote test
runner.

The test runner begins by checking the value of the test suite.
The test suite is a utility interface created for remote tests,
https://github.com/kubernetes/kubernetes/blob/master/test/e2e_node/remote/types.go.
It sets up the environment and it runs the actual tests.

In the case of node e2e tests we can use the "default" test suite, the
"cadvisor" test suite, or the "conformance" one.

The default test suite can be found here
https://github.com/kubernetes/kubernetes/blob/master/test/e2e_node/remote/node_e2e.go
and that is all that happened during our previous e2e tests.
But before diving in there, we will first see what else the test runner does.

One of the details we will cover here are the node configuration files.
If you did run the node e2e tests in the previous example, you should have seen
three tests fail.
The reason why they fail is that by default, the test runner will run tests in
a n1-standard-1 machine which has 1 vCPU and about 4.5 GB while the tests we
ran require at least 15 vCPUs and some 18 or more GBs.
You can see https://github.com/kubernetes/kubernetes/issues/91263 for more
details.

One of the things we can tune do in the test runner is to specify the machine
instance and that can be done with a configuration file that looks a bit like
this:

```yaml
images:
  cosbeta-resource1:
    image: cos-beta-81-12871-44-0
    project: cos-cloud
    machine: n1-standard-1
    metadata: "user-data<test/e2e_node/jenkins/gci-init.yaml,gci-update-strategy=update_disabled"
    tests:
      - 'resource tracking for 0 pods per node \[Benchmark\]'
```

The configuration file defines 1 or more images to use for a test.
Images in this sense are the images used to create a VM.
Each image is given a name, `cosbeta-resource1` in our case.
This image is based off `cos-beta-81-12871-44-0`, this imageis part of the
`cos-cloud` project (try running `gcloud compute images list --project cos-cloud`).
The machine field is the type of machine you wanna use and the metadata filed
requires a good read of the following link:

https://cloud.google.com/container-optimized-os/docs/how-to/create-configure-instance

Promise to give a more thorough walk through the metadata section and
cloud-init in general later on.
Finally, the tests field is the name of the e2e tests we want to execute (this
is similar to the `FOCUS` functionality of ginkgo).

And for a complete reference, this is the official explanation given in the
test runner
```go
// ImageConfig specifies what images should be run and how for these tests.
// It can be created via the `--images` and `--image-project` flags, or by
// specifying the `--image-config-file` flag, pointing to a json or yaml file
// of the form:
//
//     images:
//       short-name:
//         image: gce-image-name
//         project: gce-image-project
//         machine: for benchmark only, the machine type (GCE instance) to run test
//         tests: for benchmark only, a list of ginkgo focus strings to match tests
```

If you see a configuration file with multiple images, the test runner will go
through each of them and will execute the tests in the tests filed.

Using what we know up to this point, we are now able to fix the tests we
previously ran.
As we mentioned, the issue is that we are trying to run workloads that require
some 15 vCPUs and around 48 GB.
An n1-standard-1 machine will obviously not do it but an n1-standard-16 with 16
vCPUs and 60 GB will do it.

The only node e2e tests labeled as flaky are
* [sig-node] Node Performance Testing [Serial] [Slow] [Flaky] Run node performance testing with pre-defined workloads NAS parallel benchmark (NPB) suite - Embarrassingly Parallel (EP) workload
* [sig-node] Node Performance Testing [Serial] [Slow] [Flaky] Run node performance testing with pre-defined workloads TensorFlow workload
* [sig-node] Node Performance Testing [Serial] [Slow] [Flaky] Run node performance testing with pre-defined workloads NAS parallel benchmark (NPB) suite - Integer Sort (IS) workload

All these tests have "Node Performance Testing" in common.
With that, the following node configuration file should do it:

```yaml
images:
  cosbeta-resource1:
    image: cos-beta-81-12871-44-0
    project: cos-cloud
    machine: n1-standard-16
    metadata: "user-data<test/e2e_node/jenkins/gci-init.yaml,gci-update-strategy=update_disabled"
    tests:
      - 'Node Performance Testing'
```

and then, we can run our tests like this:
```
make test-e2e-node FOCUS="\[Flaky\]" SKIP="" PARALLELISM=1 REMOTE=true DELETE_INSTANCES=true IMAGE_CONFIG_FILE=node-test.yaml
```

After some 20 minutes, you should see 3 green tests show up in your screen.
