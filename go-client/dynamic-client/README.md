# Dynamic Client

In this section we will cover a bit about the dynamic client and why it is
useful.

## Theory

Belowe we discussed some of the pieces that you will often see when working
with Kubernetes.

[`runtime.Object`](https://godoc.org/k8s.io/apimachinery/pkg/runtime#Object)
, from "k8s.io/apimachinery/pkg/runtime", is something that you will see very often.
A `runtime.Object` is an interface that implements the following methods:

```go
type Object interface {
    GetObjectKind() schema.ObjectKind
    DeepCopyObject() Object
}
```
> Object interface must be supported by all API types registered with Scheme.
> Since objects in a scheme are expected to be serialized to the wire, the
> interface an Object must provide to the Scheme allows serializers to set the
> kind, version, and group the object is represented as. An Object may choose
> to return a no-op ObjectKindAccessor in cases where it is not expected to be
> serialized.

[`schema.ObjectKind`](https://godoc.org/k8s.io/apimachinery/pkg/runtime/schema#ObjectKind)
from "k8s.io/apimachinery/pkg/runtime/schema".
`schema.ObjectKind` is another interface that allows for setting and getting an
object's group, version, and kind.

```go
type ObjectKind interface {
    // SetGroupVersionKind sets or clears the intended serialized kind of an object. Passing kind nil
    // should clear the current setting.
    SetGroupVersionKind(kind GroupVersionKind)
    // GroupVersionKind returns the stored group, version, and kind of an object, or nil if the object does
    // not expose or provide these fields.
    GroupVersionKind() GroupVersionKind
}
```
> All objects that are serialized from a Scheme encode their type information.
> This interface is used by serialization to set type information from the
> Scheme onto the serialized version of an object. For objects that cannot be
> serialized or have unique requirements, this interface may be a no-op.

[`client.Patch`](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#Patch)
from "sigs.k8s.io/controller-runtime/pkg/client".

```go
type Patch interface {
    // Type is the PatchType of the patch.
    Type() types.PatchType
    // Data is the raw data representing the patch.
    Data(obj runtime.Object) ([]byte, error)
}
```
> Patch is a patch that can be applied to a Kubernetes object.

[`types.PatchType`](https://godoc.org/k8s.io/apimachinery/pkg/types#PatchType)
from "k8s.io/apimachinery/pkg/types".
```go
const (
    JSONPatchType           PatchType = "application/json-patch+json"
    MergePatchType          PatchType = "application/merge-patch+json"
    StrategicMergePatchType PatchType = "application/strategic-merge-patch+json"
    ApplyPatchType          PatchType = "application/apply-patch+yaml"
)
```
> these are constants to support HTTP PATCH utilized by both the client and
> server.

## Lab

In our example progam, we will try to obtain the API group, version, and
resource from an object.
We will use the "standard" clientset and a dynamic client which returns
unstructured objects.

This example shows an interesting caveat of working with Kubernetes objects:

> decoding to go structs drops apiVersion/kind, because the type info is
> inherent in the object. decoding to unstructured objects
> (like the dynamic client does) preserves that information.
> [github.com/kubernetes/client-go/issues/541#issuecomment-452312901](https://github.com/kubernetes/client-go/issues/541#issuecomment-452312901)

### Create a Cluster
```
$ kind create cluster
```

### Our Go program

We will be using code for the newest version of Kubernetes :rocket:
```
$ export GO111MODULE=on
$ go mod init
$ go get k8s.io/client-go@kubernetes-1.16.0
```

To run this (make sure you change the `push` target inside of the
[Makefile](./Makefile) to create your own image):
```
$ make && kubectl apply -f app.yaml
```

This will create a deployment called `cool-kubernetes`.
To get its logs, try:
```
$ kubectl logs deploy/cool-kubernetes -f
```


[godoc pkg/clinet]: https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client
