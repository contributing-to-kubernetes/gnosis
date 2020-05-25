# Testing Kubelets

## Table Of Contents

* [Context](#context)
* [Run Services Mode: The Tales of ETCD, API Server, and Namespace Controller](#run-services-mode-the-tales-of-etcd-api-server-and-namespace-controller)
* [Run Kubelet Mode: Getting a Kubelet Up and Running](#run-kubelet-mode-getting-a-kubelet-up-and-running)
* [Running It All](#running-it-all)
  * [In Case of Failure](#in-case-of-failure)

## Context

Welcome back!
This post is a follow up to [Running Node E2E Tests](../e2e-node-tests/).
The prvious post just went through the motions and of how to work with node e2e
tests which is enough for contributors interested in working withSIG Node.

However, for the sake of curiosity, we will now go in deeper and try to figure
out exactly what needs to happen to be able to test a Kubelet.

this time we will grab some information from the local test runner over at
https://github.com/kubernetes/kubernetes/blob/master/test/e2e_node/runner/local/run_local.go

This is program is way smaller than the remote test runner and at its center is
this command:
```
ginkgo <ginkgoFlags> e2e_node.test -- <testFlags>
```

If you went through [how to run Kubernetes e2e tests with kind](../e2e-tests/),
then this structure should look familiar.
For Kubernetes e2e tests we compile all tests into an executable called
`e2e.test` and its contents come from
https://github.com/kubernetes/kubernetes/tree/master/test/e2e.
In the same manner, we have a test executable whose contents now come from
https://github.com/kubernetes/kubernetes/tree/master/test/e2e_node.

And if you want to compile the node e2e test executable you can do so in the
same manner
```
make all WHAT='test/e2e/e2e_node.test'
```

In the case of standard e2e tests, the test executable assumes that you have a
cluster all set up and the executable just executes tests.
Node tests are particularly interesting because on top of running tests it also
get some Kubelets (and sometimes an API server or an ETCD) configured and
working.

---

## Run Services Mode: The Tales of ETCD, API Server, and Namespace Controller

The first thing we will look at is at the `--run-services-mode` flag.
This flag

> If true, only run services (etcd, apiserver) in current process, and not run test.

To execute our test binary (make sure to run `make WHAT='ginkgo e2e_node.test'
beforehand)

```
OUTPUTDIR=/path/to/make/artifacts/something/like/go/src/k8s.io/kubernetes/_output/local/go/bin

sudo ${OUTPUTDIR}/ginkgo  -nodes=1 -untilItFails=false  ${OUTPUTDIR}/e2e_node.test -- \
  --container-runtime=docker --alsologtostderr --v 4 --report-dir=/tmp/_artifacts/200524T140146 \
  --node-name cool-node \
  --kubelet-flags="--container-runtime=docker" \
  --kubelet-flags="--network-plugin= --cni-bin-dir=" \
  --run-services-mode=true
```


We can check that our test binary is actually taking over some of our ports
(for our API server and our ETCD instance)
```
sudo lsof -i -P -n | grep  LISTEN
```

and we can even send requests to it!!!
```
curl -H 'Content-Type: application/yaml' 'http://127.0.0.1:8080/api/v1/namespaces'
```
```json
{
  "kind": "NamespaceList",
  "apiVersion": "v1",
  "metadata": {
    "selfLink": "/api/v1/namespaces",
    "resourceVersion": "49"
  },
  "items": [
    {
      "metadata": {
        "name": "default",
        "selfLink": "/api/v1/namespaces/default",
        "uid": "e1db06cf-c208-4f91-811f-6fb636c01786",
        "resourceVersion": "40",
        "creationTimestamp": "2020-05-24T18:22:24Z",
        "managedFields": [
          {
            "manager": "e2e_node.test",
            "operation": "Update",
            "apiVersion": "v1",
            "time": "2020-05-24T18:22:24Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {"f:status":{"f:phase":{}}}
          }
        ]
      },
      "spec": {
        "finalizers": [
          "kubernetes"
        ]
      },
      "status": {
        "phase": "Active"
      }
    },
    {
      "metadata": {
        "name": "kube-node-lease",
        "selfLink": "/api/v1/namespaces/kube-node-lease",
        "uid": "2e352367-afbf-4ff6-88fa-e02a3e216903",
        "resourceVersion": "6",
        "creationTimestamp": "2020-05-24T18:22:23Z",
        "managedFields": [
          {
            "manager": "e2e_node.test",
            "operation": "Update",
            "apiVersion": "v1",
            "time": "2020-05-24T18:22:23Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {"f:status":{"f:phase":{}}}
          }
        ]
      },
      "spec": {
        "finalizers": [
          "kubernetes"
        ]
      },
      "status": {
        "phase": "Active"
      }
    },
    {
      "metadata": {
        "name": "kube-public",
        "selfLink": "/api/v1/namespaces/kube-public",
        "uid": "18076030-0f32-46f3-87c6-e522bf8ade8d",
        "resourceVersion": "5",
        "creationTimestamp": "2020-05-24T18:22:23Z",
        "managedFields": [
          {
            "manager": "e2e_node.test",
            "operation": "Update",
            "apiVersion": "v1",
            "time": "2020-05-24T18:22:23Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {"f:status":{"f:phase":{}}}
          }
        ]
      },
      "spec": {
        "finalizers": [
          "kubernetes"
        ]
      },
      "status": {
        "phase": "Active"
      }
    },
    {
      "metadata": {
        "name": "kube-system",
        "selfLink": "/api/v1/namespaces/kube-system",
        "uid": "6a3fa478-707a-462d-a8b2-ce00ecde1eea",
        "resourceVersion": "4",
        "creationTimestamp": "2020-05-24T18:22:23Z",
        "managedFields": [
          {
            "manager": "e2e_node.test",
            "operation": "Update",
            "apiVersion": "v1",
            "time": "2020-05-24T18:22:23Z",
            "fieldsType": "FieldsV1",
            "fieldsV1": {"f:status":{"f:phase":{}}}
          }
        ]
      },
      "spec": {
        "finalizers": [
          "kubernetes"
        ]
      },
      "status": {
        "phase": "Active"
      }
    }
  ]
}
```

Again, time to look under the hood.

If we call our test with `--run-services-mode=true` then we are using an
internal object called `e2eServices`.
Let's walk the code through.

[k/k/test/e2e_node/e2e_node_suite_test.go](https://github.com/kubernetes/kubernetes/blob/4e8b56e6671893757d40e2001a3c615acebc13a2/test/e2e_node/e2e_node_suite_test.go#L120-L125)
is essentially the first place where our test code begins.
Here, we check if the `--run-services-mode` flag is set to true - in this mode
we will run the Kubernetes API server and an ETCD in the current process
without executing any tests.

```go
func TestE2eNode(t *testing.T) {
  if *runServicesMode {
	  // If run-services-mode is specified, only run services in current process.
		services.RunE2EServices(t)
		return
	}
  ...
```

[`services.RunE2eServices()`](https://github.com/kubernetes/kubernetes/blob/4e8b56e6671893757d40e2001a3c615acebc13a2/test/e2e_node/services/services.go#L106-L118)
we set the feature gates for our Kubelet and then instantiate an `e2eServices`
services object and proceed to run it.
```go
// Populate global DefaultFeatureGate with value from TestContext.FeatureGates.
	// This way, statically-linked components see the same feature gate config as the test context.
	if err := utilfeature.DefaultMutableFeatureGate.SetFromMap(framework.TestContext.FeatureGates); err != nil {
		t.Fatal(err)
	}
	e := newE2EServices()
	if err := e.run(t); err != nil {
		klog.Fatalf("Failed to run e2e services: %v", err)
	}
```

From the structure of `e2eServices`, we can coroborate that it inded manages an
ETCD and an API server
```go
// e2eService manages e2e services in current process.
type e2eServices struct {
	rmDirs []string
	// statically linked e2e services
	etcdServer   *etcd3testing.EtcdTestServer
	etcdStorage  *storagebackend.Config
	apiServer    *APIServer
	nsController *NamespaceController
}
```
See
[k/k/test/e2e_node/services/internal_services.go#L30-L38](https://github.com/kubernetes/kubernetes/blob/4e8b56e6671893757d40e2001a3c615acebc13a2/test/e2e_node/services/internal_services.go#L30-L38).

The first step is to start ETCD.
For this we will use a handy Kubernetes library
```go
import (
  etcd3testing "k8s.io/apiserver/pkg/storage/etcd3/testing"
)
```
[k/k/staging/src/k8s.io/apiserver/pkg/storage/etcd3/testing](https://github.com/kubernetes/kubernetes/tree/master/staging/src/k8s.io/apiserver/pkg/storage/etcd3/testing).


After that, we start an API Server locally.
This part configures the options and configuration for the Kubernetes API
Server and proceeds to call the code for the API Server directly.

And our last step is to run a namespace controller!
The namespace controller is the component that manages the lifecycle of
Kubernetes namespaces and it can be found here
https://github.com/kubernetes/kubernetes/tree/master/pkg/controller/namespace.
(You should definetely checkout the code for the namespace controller,
specifically how the `Run()` function starts a pool of workers).

Once all this components are running then we just wait for a termination
signal.
Meanwhile, you will have an API Server up and running.

---

## Run Kubelet Mode: Getting a Kubelet Up and Running

Same test binary different flag for this section!
We will now look at `--run-kubelet-mode=true`.

Again, the entrypoint is the
[`TestE2ENode` function](https://github.com/kubernetes/kubernetes/blob/4e8b56e6671893757d40e2001a3c615acebc13a2/test/e2e_node/e2e_node_suite_test.go#L126-L130):
```go
func TestE2eNode(t *testing.T) {
	...
  if *runKubeletMode {
		// If run-kubelet-mode is specified, only start kubelet.
		services.RunKubelet()
		return
	}
  ...
```

In this case, the
[`services.RunKubelet()`](https://github.com/kubernetes/kubernetes/blob/4e8b56e6671893757d40e2001a3c615acebc13a2/test/e2e_node/services/kubelet.go#L74-L87)
function will use a public object:
```go
func RunKubelet() {
	var err error
	// Enable monitorParent to make sure kubelet will receive termination signal
	// when test process exits.
	e := NewE2EServices(true /* monitorParent */)
	defer e.Stop()
	e.kubelet, err = e.startKubelet()
	if err != nil {
		klog.Fatalf("Failed to start kubelet: %v", err)
	}
	// Wait until receiving a termination signal.
	waitForTerminationSignal()
}
```

This `NewE2EServices()` function creates a
[`E2EServices` object](https://github.com/kubernetes/kubernetes/blob/master/test/e2e_node/services/services.go#L33-L43)
```go
// E2EServices starts and stops e2e services in a separate process. The test
// uses it to start and stop all e2e services.
type E2EServices struct {
	// monitorParent determines whether the sub-processes should watch and die with the current
	// process.
	rmDirs        []string
	monitorParent bool
	services      *server
	kubelet       *server
	logs          logFiles
}
```

This public `E2EServices` object builds on top of the other.
One can proceed with the "run kubelet mode" and start a kubelet
```go
var err error
// Enable monitorParent to make sure kubelet will receive termination signal
// when test process exits.
e := NewE2EServices(true /* monitorParent */)
defer e.Stop()
e.kubelet, err = e.startKubelet()
if err != nil {
  klog.Fatalf("Failed to start kubelet: %v", err)
}
// Wait until receiving a termination signal.
waitForTerminationSignal()
```

Which comes in handy, but one can run the API Server, ETCD, and Namespace
controller, AND a Kubelet!
```go
// Enable monitorParent to make sure kubelet will receive termination signal
// when test process exits.
e := NewE2EServices(true /* monitorParent */)
defer e.Stop()

if err := e.Start(); err != nil {
  // Do something...
}
```

The `Start()` function will

> Start starts the e2e services in another process by calling back into the
> test binary.  Returns when all e2e services are ready or an error.
> 
> We want to statically link e2e services into the test binary, but we don't
> want their glog output to pollute the test result. So we run the binary in
> run-services-mode to start e2e services in another process.
> The function starts 2 processes:
>   * internal e2e services: services which statically linked in the test binary - apiserver, etcd and namespace controller.
>   * kubelet: kubelet binary is outside. (We plan to move main kubelet start logic out when we have standard kubelet launcher)

And we will now go and run the API Server, ETCD, a Namespace controller, and a
Kubelet using the above logic!

---

## Running It All

As previously stated, there is a way to un a Kubelet alongside a Kubernetes API
Server, an ECTD instance, and a Namespace controller.
For that we just have to trigger some part of the code that calls the
`E2EServices` `Start()` function.
Luckily for us, that there is a flag for us to do just that without modifying
any of the code.

The flag we will use is `--stop-services=false`.
This falg is defined in
[k/k/test/e2e_node/util.go](https://github.com/kubernetes/kubernetes/blob/4e8b56e6671893757d40e2001a3c615acebc13a2/test/e2e_node/util.go#L65)
and 

> If true, stop local node services after running tests

We know to use it because the entrypoint references it here
[k/k/test/e2e_node/e2e_node_suite_test.go](https://github.com/kubernetes/kubernetes/blob/master/test/e2e_node/e2e_node_suite_test.go#L192-L197)
```go
if *startServices {
	// If the services are expected to stop after test, they should monitor the test process.
	// If the services are expected to keep running after test, they should not monitor the test process.
	e2es = services.NewE2EServices(*stopServices)
	gomega.Expect(e2es.Start()).To(gomega.Succeed(), "should be able to start node services.")
	klog.Infof("Node services started.  Running tests...")
}
```

`startServices` is another flag but it is true by default, so we only need to
set `stopServices` to false in order to run our test executable and leave an
instance of the Kubernetes API Server, an ETCD, a Namespace controller, and a
Kubelet running for us to test.

Now here comes the command that will do all this for us (keep in mind that it
may fail but below the command we will tell you at least one reason why it may
fail :smile:)

```
sudo ${OUTPUTDIR}/ginkgo -nodes=1 -untilItFails=false ${OUTPUTDIR}/e2e_node.test -- \
  --container-runtime=docker --alsologtostderr --v 4 --report-dir=/tmp/_artifacts/200524T140146 \
  --node-name cool-node --kubelet-flags="--container-runtime=docker" \
  --kubelet-flags="--network-plugin= --cni-bin-dir=" \
  --stop-services=false
```

Exactly the same command as before but with `--stop-services` set to false.

### In Case of Failure

If the above command didn't work for you: you saw a message about a failed
health/readiness check, then you might have also seen the command our program
is using to start the kubelet.

In our case it was something like this
```
sudo /usr/bin/systemd-run --unit=kubelet-20200524T194401.service --slice=runtime.slice \
  --remain-after-exit /home/user/go/src/k8s.io/kubernetes/_output/local/go/bin/kubelet \
  --kubeconfig /home/user/go/src/k8s.io/kubernetes/_output/local/go/bin/kubeconfig \
  --root-dir /var/lib/kubelet --v 4 --logtostderr \
  --dynamic-config-dir /home/user/go/src/k8s.io/kubernetes/_output/local/go/bin/dynamic-kubelet-config \
  --network-plugin=kubenet --cni-bin-dir /home/user/go/src/k8s.io/kubernetes/_output/local/go/bin/cni/bin \
  --cni-conf-dir /home/user/go/src/k8s.io/kubernetes/_output/local/go/bin/cni/net.d \
  --cni-cache-dir /home/user/go/src/k8s.io/kubernetes/_output/local/go/bin/cni/cache \
  --hostname-override cool-node --container-runtime docker \
  --container-runtime-endpoint unix:///var/run/dockershim.sock \
  --config /home/user/go/src/k8s.io/kubernetes/_output/local/go/bin/kubelet-config \
  --container-runtime=docker --network-plugin= --cni-bin-dir=
```

In our case, the kubelet is getting started with systemd.
So we tried copying and pasting that command in another terminal and then
executing it.

Notice that the name of the systemd unit is passed as a value to the `--unit`
flag.
And since we have that info we can check its status:
```
systemctl -l status kubelet-20200524T194401.service
...
server.go:274] failed to run Kubelet: running with swap on is not supported, please disable swap! or set --fail-swap-on fl
```

With that complain about swap, we proceeded to delete the current failed
Kubelet systemd unit
```
systemctl reset-failed kubelet-20200524T194401.service
```

We turned off swap
```
sudo swapoff -a
```

(You can turn it back on with `sudo swapon -a`).

And we got it to work! :tada:
