# Fix bug in apiserver service cidr split #85968

## Description

This write up is on PR https://github.com/kubernetes/kubernetes/pull/85968.
This PR fixed a bug on the Kubernetes API server.

The issue discussed in this PR revolves around the use of 
[strings.Split()](https://godoc.org/strings#Split) (which takes a string and a
separator as input) on a string that is possibly empty.

Let's look at the following example:

```go
import (
  "string"
)

func main() {
  output := strings.Split("", ",")
}
```

In this case, `output` will result in a slice of length 1 with only and empty string in it.

As it turns out, a command-line flag of the Kubernetes API server, `--service-cluster-ip-range`,
ran into an issue with the expected outcome of `strings.Split()`.
The rest of this write up will walk us through the problem in more depth.

### Toy Model

If you want to mess around with the logic or to more closely follow the code we
added a sample version of the kube-apiserver here in
[`example-cobra`](./example-cobra).

This toy model of the kube apiserver has only the core logic dealing with the
issue at hand but it mirrors the code organization and flow to what the authors
believe to be the key details.

You can compile it with a simple `go build`.
Then run it as any other executable:

```
go build && ./example-cobra --help
```

Checkout the `Misc flags` :smile:.

## A Tour of the Kube API Server

Kubernetes has various components, the API server, the kubelet, etc.

All these components are defined and implemented in the
[kubernetes/kubernetes](https://github.com/kubernetes/kubernetes), k/k, repo.

The way the code is organized is that all the Kubernetes core components are
defined within the
[cmd/](https://github.com/kubernetes/kubernetes/tree/master/cmd) directory at
the root of the repo.

In our case, we can find the Kube API server in
[k/k/cmd/kube-apiserver](https://github.com/kubernetes/kubernetes/tree/master/cmd/kube-apiserver).
Within this directory, we will see right away the entrypoint (how the apis erver runs) defined in
[k/k/cmd/kube-apiserver/apiserver.go](https://github.com/kubernetes/kubernetes/blob/master/cmd/kube-apiserver/apiserver.go).

If you open up the file, you should see the following:

```go
package main

import (
  ...
  "k8s.io/kubernetes/cmd/kube-apiserver/app"
)

func main() {
  ...
  command := app.NewAPIServerCommand()
  
  ....
  if err := command.Execute(); err != nil {
    os.Exit(1)
  }
}
```

Noticed how the API server is implemented as a CLI.
Its entrypoint is defined as part of a "command" that is executed in order to
run the server.
The command itself is defined within the app subpackage,
[k8s.io/kubernetes/cmd/kube-apiserver/app](https://github.com/kubernetes/kubernetes/blob/master/cmd/kube-apiserver/app/),
and it is implemented as a [spf13/cobra](https://github.com/spf13/cobra) command.

The definition of the cobra can be found in
[k/k/cmd/kube-apiserver/app/server.go](https://github.com/kubernetes/kubernetes/blob/master/cmd/kube-apiserver/app/server.go):

```go
import (
  ...
  "k8s.io/kubernetes/cmd/kube-apiserver/app/options"
  ...
)

// NewAPIServerCommand creates a *cobra.Command object with default parameters
func NewAPIServerCommand() *cobra.Command {
  s := options.NewServerRunOptions()
  cmd := &cobra.Command{
    Use: "kube-apiserver",
    Long: `The Kubernetes API server validates and configures data
for the api objects which include pods, services, replicationcontrollers, and
others. The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
    RunE: func(cmd *cobra.Command, args []string) error {
      verflag.PrintAndExitIfRequested()
      utilflag.PrintFlags(cmd.Flags())
      
      // set default options
      completedOptions, err := Complete(s)
      if err != nil {
        return err
      }
      
      // validate options
      if errs := completedOptions.Validate(); len(errs) != 0 {
        return utilerrors.NewAggregate(errs)
      }
      
      return Run(completedOptions, genericapiserver.SetupSignalHandler())
    },
  }
  ...
}
```

There are a couple things we should take away when reading through this code.
For our purposes, these are

```go
s := options.NewServerRunOptions()
```

and 

```
// set default options
completedOptions, err := Complete(s)
if err != nil {
  return err
}
```


### Server Options

The first bit we will look at is

```go
s := options.NewServerRunOptions()
```

from the cobra command definion.

This `options.NewServerRunOptions()` refers to the function
`NewServerRunOptions()` of
[k8s.io/kubernetes/cmd/kube-apiserver/app/options](https://github.com/kubernetes/kubernetes/blob/master/cmd/kube-apiserver/app/options/options.go).

As the name entails, this package implements the options (or configuration) for
the kube API server to run and it does it through the `ServerRunOptions`
struct:

```go
// ServerRunOptions runs a kubernetes api server.
type ServerRunOptions struct {
  GenericServerRunOptions *genericoptions.ServerRunOptions
  Etcd                    *genericoptions.EtcdOptions
  ...
  
  Features                *genericoptions.FeatureOptions
  ...
  
  CloudProvider           *kubeoptions.CloudProviderOptions
  ...
  
  KubeletConfig             kubeletclient.KubeletClientConfig
  ...
  
  // ServiceClusterIPRange is mapped to input provided by user
  ServiceClusterIPRanges string
  ...
}
```

In here you can see that the `ServerRunOptions` struct specifies a
`ServiceClusterIPRanges` field which is a string.
And this value is mapped to input provided by the user.
In this case, this value corresponds to the command-line flag

```
--service-cluster-ip-range string
A CIDR notation IP range from which to assign service cluster IPs. This must not overlap with any IP ranges assigned to nodes for pods.
```

Please see
[kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/)
for more details.

This flag is used to tell the API server the range of IP addresses that can be
used to allocate `ClusterIP`s for Kubernetes services, see
[kubernetes.io/docs/concepts/services-networking/service/](https://kubernetes.io/docs/concepts/services-networking/service/).

Within the API server's cobra command, after being defined, we pass our server
options through a function called `Complete()`, this function is defined in the
same file,

```go
// set default options
completedOptions, err := Complete(s)
if err != nil {
  return err
}
```

And here is where the issue comes in!

## The Bug

The `Complete()` function, defined in the same file as the cobra command

```go
// Complete set default ServerRunOptions.
// Should be called after kube-apiserver flags parsed.
func Complete(s *options.ServerRunOptions) (completedServerRunOptions, error)
```

Takes the defined server options and populates them with all the necessary
information to get the API server up and running.

Before [PR #85968](https://github.com/kubernetes/kubernetes/pull/85968), the
`Complete()` function had the following code in it:

```go
serviceClusterIPRangeList := strings.Split(s.ServiceClusterIPRanges, ",")

var apiServerServiceIP net.IP
var serviceIPRange net.IPNet
var err error
// nothing provided by user, use default range (only applies to the Primary)
if len(serviceClusterIPRangeList) == 0 {
  var primaryServiceClusterCIDR net.IPNet
  serviceIPRange, apiServerServiceIP, err = master.ServiceIPRange(primaryServiceClusterCIDR)
  if err != nil {
    return options, fmt.Errorf("error determining service IP ranges: %v", err)
  }
  s.PrimaryServiceClusterIPRange = serviceIPRange
}

if len(serviceClusterIPRangeList) > 0 {
  _, primaryServiceClusterCIDR, err := net.ParseCIDR(serviceClusterIPRangeList[0])
  if err != nil {
    return options, fmt.Errorf("service-cluster-ip-range[0] is not a valid cidr")
  }

  serviceIPRange, apiServerServiceIP, err = master.ServiceIPRange(*(primaryServiceClusterCIDR))
  if err != nil {
    return options, fmt.Errorf("error determining service IP ranges for primary service cidr: %v", err)
  }
  s.PrimaryServiceClusterIPRange = serviceIPRange
}
```

Notice how `s.ServiceClusterIPRanges` is being split!
The problem with this is that if there is no value provided by the user through
`--service-cluster-ip-range` then `ServiceClusterIPRanges`will default to its
zero value which is an empty string.

**Note**: Zero values are a Go thing. From [A Tour of Go: Zero
Values](https://tour.golang.org/basics/12)

> Variables declared without an explicit initial value are given their zero value.
>
> The zero value is:
>
> 0 for numeric types,
> false for the boolean type, and
> "" (the empty string) for strings.

So if the user doesn't provide a value for the service cluster IP ranges, then
we will end splitting an empty string.
This will result in `serviceClusterIPRangeList` having length 1 and only
containing an empty string.

However, if you read through the rest, the code expects
`serviceClusterIPRangeList` to have a length of 0 when the user did not specify
a value!!!
