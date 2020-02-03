# Fix bug in apiserver service cidr split

## Description

This write up is on PR https://github.com/kubernetes/kubernetes/pull/85968.
This PR fixed a bug on the Kubernetes API server.

The bug came about by the use of
[strings.Split()](https://godoc.org/strings#Split).

From the documentation:
```go
func Split(s, sep string) []string
```
> Split slices `s` into all substrings separated by `sep` and returns a slice of the substrings between those separators.
> If `s` does not contain `sep` and `sep` is not empty, Split returns a slice of length 1 whose only element is `s`.
> If `sep` is empty, Split splits after each UTF-8 sequence. If both `s` and `sep` are empty, Split returns an empty slice.
> It is equivalent to SplitN with a count of -1.

You could use this function as follows

```go
fmt.Println(strings.Split("a,b,c", ","))
```

In which case it would output a slice with the elements `["a" "b" "c"]`.

In the PR description, it is mentioned that if the `s` input to `strings.Split`
was an empty string (`""`), then the length of the returned slice would be one
and whose only element would be an empty string.
This would cause an issue with the following check

```go
if len(s.GenericServerRunOptions.ExternalHost) == 0 {
  if len(s.GenericServerRunOptions.AdvertiseAddress) > 0 {
    s.GenericServerRunOptions.ExternalHost = s.GenericServerRunOptions.AdvertiseAddress.String()
  } else {
    if hostname, err := os.Hostname(); err == nil {
      s.GenericServerRunOptions.ExternalHost = hostname
		} else {
			return options, fmt.Errorf("error finding host name: %v", err)
		}
	}
	klog.Infof("external host was not specified, using %v", s.GenericServerRunOptions.ExternalHost)
}
```
In the `complete()` function within
[/cmd/kube-apiserver/app/server.go][/cmd/kube-apiserver/app/server.go].

To go further we need some more details :smile:

## ExternalHost

### Options
To better understand how the above check works and what it does we need to
digress a bit.

So the above check, the `if len(s.GenericServerRunOptions.ExternalHost) == 0`
is in a function called `complete()` within
[/cmd/kube-apiserver/app/server.go][/cmd/kube-apiserver/app/server.go].
This function has the following signature:

```go
// Complete set default ServerRunOptions.
// Should be called after kube-apiserver flags parsed.
func Complete(s *options.ServerRunOptions) (completedServerRunOptions, error) {
...
```

As the comment mentions, this functions sets the default values for the
`ServerRunOptions` struct.
This object is defined in
[/cmd/kube-apiserver/app/options/options.go][/cmd/kube-apiserver/app/options/options.go]

```go
// ServerRunOptions runs a kubernetes api server.
type ServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions
	Etcd                    *genericoptions.EtcdOptions
	SecureServing           *genericoptions.SecureServingOptionsWithLoopback
  ...
```

A lot of the fields are in turned defined in the package `genericoptions` which
is `"k8s.io/apiserver/pkg/server/options"`.

You won't find the above package at the root of the Kubernetes repo.
Instead, you will have to search for it within the `staging` directory.
The above package can be found here
[/staging/src/k8s.io/apiserver/pkg/server/options][/staging/src/k8s.io/apiserver/pkg/server/options].

One of the fields from the `ServerRunOptions` struct is the following

```go
SecureServing           *genericoptions.SecureServingOptionsWithLoopback
```

The `SecureServingOptionsWithLoopback` struct is defined in
[/staging/src/k8s.io/apiserver/pkg/server/options/serving_with_loopback.go][/staging/src/k8s.io/apiserver/pkg/server/options/serving_with_loopback.go]

```go
type SecureServingOptionsWithLoopback struct {
	*SecureServingOptions
}
```

This means that it is an "instance" of the `SecureServingOptions` struct which
is defined in
[/staging/src/k8s.io/apiserver/pkg/server/options/serving.go][/staging/src/k8s.io/apiserver/pkg/server/options/serving.go]

### MaybeDefaultWithSelfSignedCerts

Taking a small detour to look at a method of the `SecureServingOptions` struct
from
[/staging/src/k8s.io/apiserver/pkg/server/options/serving.go][/staging/src/k8s.io/apiserver/pkg/server/options/serving.go]:

```go
func (s *SecureServingOptions) MaybeDefaultWithSelfSignedCerts(publicAddress string, alternateDNS []string, alternateIPs []net.IP) error {
```


[/cmd/kube-apiserver/app/server.go]: https://github.com/kubernetes/kubernetes/blob/master/cmd/kube-apiserver/app/server.go
[/cmd/kube-apiserver/app/options/options.go]: https://github.com/kubernetes/kubernetes/blob/master/cmd/kube-apiserver/app/options/options.go
[/staging/src/k8s.io/apiserver/pkg/server/options]: https://github.com/kubernetes/kubernetes/tree/master/staging/src/k8s.io/apiserver/pkg/server/options
[/staging/src/k8s.io/apiserver/pkg/server/options/serving_with_loopback.go]: https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apiserver/pkg/server/options/serving_with_loopback.go
[/staging/src/k8s.io/apiserver/pkg/server/options/serving.go]: https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apiserver/pkg/server/options/serving.go
