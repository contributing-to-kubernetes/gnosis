# Kubernetes-Style APIs

Welcome to yet another adventure through the Kubernetes codebase.

This time we will see how Kubernetes implements its APIs.
You know all those pods, deployments, services, and all that that you used all
the time?
Well, the plan for today is to explore all the API machinery (ha!) that makes
that possible.

Kubernetes-style APIs are at their core a convination of Go `struct`s and some
code generation to wrangle data into and out of said Go `struct`s.

We will start this trip into kubenernetes by eploring the code generating tools
that are currently used.

## Deep Copy Generator

We already hinted that the implementation of Kubernetes-style APIs relies on
handling data and passing it into and out of Go `struct`s.
If you have worked with multiple nested structs then you may have already ran
into the problem of moving data from one struct into another.
If the terms "shallow" and "deep" copy don't ring a bell then we recommend you
give https://flaviocopes.com/go-copying-structs/ a quick read.

In Kubernetes, we have a fancy tool called `deepcopy-gen` which does some code
generation that results in `DeepCopy` methods.
Like everything, hopefully an example will make it make more sense.

`deepcopy-gen` can be found in staging,
https://github.com/kubernetes/kubernetes/tree/master/staging/src/k8s.io/code-generator/cmd/deepcopy-gen.


For the sake of experimentation, we provide a small bash script,
[build.sh](./build.sh) to get the `deepcopy-gen` binary.
The build script will compile a binary for linux on amd64 but you can tweak it
to cross-compile something else.
This script is the simple version of
https://github.com/kubernetes-sigs/kind/blob/edecdfee8878ac00c7ae00485fca3f95d351e1ac/hack/go_container.sh
which is used in https://github.com/kubernetes-sigs/kind.


Anyway, a
```
./build.sh go build k8s.io/code-generator/cmd/deepcopy-gen
```

will get you a `deepcopy-gen` in the current working directory.


The entrypoint to `deepcopy-gen` is currently in
https://github.com/kubernetes/gengo/blob/master/examples/deepcopy-gen/main.go

There isn't a lot of documentation for `deepcopy-gen` but between
https://github.com/kubernetes/gengo/blob/master/examples/deepcopy-gen/main.go
and asking `deepcopy-gen` for help
```
./tools/deepcopy-gen --help
```

We should be able to get by.

A couple details that we will need to explicitly mention are

* `deepcopy-gen` relies on comments to know what types it should build deep
  copy methods for (we will show an example here shortly).


```
./tools/deepcopy-gen -i ./apis/v1alpha1/ -O zz_generated.deepcopy --go-header-file boilerplate.go.txt
```

This will result in the creation of
[pkg/deepcopy_generated.go](./pkg/deepcopy_generated.go), which we already
included in here.

And to test our deep copiers we included [main.go](./main.go).
If you run it, you will see something like the following
```
$ go run main.go
cluster1: &pkg.Cluster{Name:"kube", Nodes:[]string{"1", "2", "3"}}
cluster2: &pkg.Cluster{Name:"kube", Nodes:[]string{"1", "2", "3"}}
cluster3: &pkg.Cluster{Name:"", Nodes:[]string(nil)}
cluster3: &pkg.Cluster{Name:"kube", Nodes:[]string{"1", "2", "3"}}
cluster1 and cluster3 the same: false
```

and if you look at the code,
```go
  c1 := &api.Cluster{Name: "kube", Nodes: []string{"1", "2", "3"}}
  ...

  c2 := c1.DeepCopy()
  ...

  c3 := &api.Cluster{}
  c2.DeepCopyInto(c3)
  ...

  fmt.Printf("cluster1 and cluster3 the same: %v\n", c1 == c3)
```

we essentially create 1 `Cluster` object and then we use the generated
`DeepCopy()` and `DeepCopyInto()` methods to create copies.

The very last line compares if the first and last "cluster" objects are the
same and we indeed find they aren't - which is good, because we want deep
copies (we want the data alone).

## Conversion Generator

```
$ ./tools/conversion-gen -i ./apis/v1alpha1/ -O zz_generated.conversion --go-header-file boilerplate.go.txt
E0718 21:08:28.977033   25990 conversion.go:755] Warning: could not find nor generate a final Conversion function for github.com/contributing-to-kubernetes/gnosis/stories/kube-apis/apis/v1alpha2.Cluster -> github.com/contributing-to-kubernetes/gnosis/stories/kube-apis/apis/v1alpha1.Cluster
E0718 21:08:28.977988   25990 conversion.go:756]   the following fields need manual conversion:
E0718 21:08:28.977996   25990 conversion.go:758]       - Provider
```

The `zz_generated.conversion.go` needs a wrapper to fully convert to v1alpha2
because there are fields that exist in v1alpha2 that don't exist in v1alpha1
and have to be set manually.
