# Kubenet fetches gateway from CNI result instead of calculating gateway from pod cidr #85993

## Description

This write up is on [PR #85993](https://github.com/kubernetes/kubernetes/pull/85993). This PR fixed a regression in kubenet that prevented pods from obtaining IP addresses.

The goal of this PR was to refactor the way that kubenet fetches the gateway. The problem was that the kubenet was having issues allocating addresses for pods, with a node spec in 1.16, as described in [Issue #84541](https://github.com/kubernetes/kubernetes/issues/84541).

### Brief Overview of Kubenet

It is a Linux only network plugin, meant to be basic and simple. It's not expected to implement things such as `cross-node networking` or `network policy` and is typically used in conjunction with a cloud provider that sets up routing rules for node(s).

Here are some things that `kubenet` will do:
- Create a `Linux bridge` named `cbr0`
- Create a `veth pair` for each `pod` connected to `cbr0`
- Assign an `IP address` to the `pod` end of the `veth pair`
  - This `IP address` comes from a `range` that has been assigned to the `node` through configuration or by the `controller-manager`
- Assign an `MTU` to the `cbr0`
  - This `MTU` matches the smallest `MTU` of an `enabled normal interface` on the `host`

More information can be located within the [k8s.io docs](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/). This is also the source in which this overview was derived.

## Old Logic Breakdown

Before the PR, the gateway was derived from the pod cidr by ranging over the list of current pod cidrs:

```go
for idx, currentPodCIDR := range podCIDRs {
	_, cidr, err := net.ParseCIDR(currentPodCIDR)
	if nil != err {
		klog.Warningf("Failed to generate CNI network config with cidr %s at index:%v: %v", currentPodCIDR, idx, err)
		return
	}
	// create list of ips and gateways
	cidr.IP[len(cidr.IP)-1] += 1 // Set bridge address to first address in IPNet
	plugin.podCIDRs = append(plugin.podCIDRs, cidr)
	plugin.podGateways = append(plugin.podGateways, cidr.IP)
}
```

Notice how, in the above logic, we are creating a list of `ips` and `gateways` by setting up a `Linux bridge`, then appending the `pod cidrs` and `gateways` to it. Let's break it down a bit.

First, we range over the `pod cidrs` to get their `IP network` value, a.k.a. [IPNet](https://golang.org/pkg/net/#IPNet), but in this context we call it the `cidr`.
```go
for idx, currentPodCIDR := range podCIDRs {
	_, cidr, err := net.ParseCIDR(currentPodCIDR)
```

Now that we have the `cidr` value, let's create the list of `ips` and `gateways`. We first need to set the `bridge address`. This is done by setting it to the first address in the [IPNet](https://golang.org/pkg/net/#IPNet), a.k.a. that `cidr` value we've just mentioned:
```go
cidr.IP[len(cidr.IP)-1] += 1
```

What that line above does is take an [IPNet.IP](https://golang.org/pkg/net/#IP) value (i.e. `10.0.0.0`) and increment that 32 bit address by 1 (i.e. `10.0.0.1`). Notice that we are mutating the network number from the [IPNet](https://golang.org/pkg/net/#IPNet) when we say `cidr.IP`. If you look at how [IPNet](https://golang.org/pkg/net/#IPNet) works, it's a struct containing the [IP](https://golang.org/pkg/net/#IP) (`network number`) and the [IPMask](https://golang.org/pkg/net/#IPMask). This is how that struct looks in the golang library:
```go
type IPNet struct {
    IP   IP     // network number
    Mask IPMask // network mask
}
```

Now we move on to this line of code:
```go
plugin.podCIDRs = append(plugin.podCIDRS, cidr)
```

The `plugin` part, is a method receiver for the `Event` function that we are currently in, referencing the `kubenetNetworkPlugin` defined earlier in the file as:
```go
type kubenetNetworkPlugin struct {
	network.NoopNetworkPlugin

	host            network.Host
	netConfig       *libcni.NetworkConfig
	loConfig        *libcni.NetworkConfig
	cniConfig       libcni.CNI
	bandwidthShaper bandwidth.Shaper
	mu              sync.Mutex //Mutex for protecting podIPs map, netConfig, and shaper initialization
	podIPs          map[kubecontainer.ContainerID]utilsets.String
	mtu             int
	execer          utilexec.Interface
	nsenterPath     string
	hairpinMode     kubeletconfig.HairpinMode
	// kubenet can use either hostportSyncer and hostportManager to implement hostports
	// Currently, if network host supports legacy features, hostportSyncer will be used,
	// otherwise, hostportManager will be used.
	hostportSyncer    hostport.HostportSyncer
	hostportSyncerv6  hostport.HostportSyncer
	hostportManager   hostport.HostPortManager
	hostportManagerv6 hostport.HostPortManager
	iptables          utiliptables.Interface
	iptablesv6        utiliptables.Interface
	sysctl            utilsysctl.Interface
	ebtables          utilebtables.Interface
	// binDirs is passed by kubelet cni-bin-dir parameter.
	// kubenet will search for CNI binaries in DefaultCNIDir first, then continue to binDirs.
	binDirs           []string
	nonMasqueradeCIDR string
	cacheDir          string
	podCIDRs          []*net.IPNet
	podGateways       []net.IP
}
```

So, when we want to append the updated `cidr` to `plugin.podCIDRs` we are referring to a list of type `*net.IPNet`.

The next line also does some appending,
```go
plugin.podGateways = append(plugin.podGateways, cidr.IP)
```
But, instead of the whole `IPNet`, we are appending `cidr.IP` to a list of type `net.IP`.

## New Logic Breakdown

The logic described above was replaced with changes to the same `kubenet_linux.go` file. Changes were made to the `kubenetNetworkPlugin` struct and to some of the methods on that struct:
- `Event`
- `setup`
- `syncEbtablesDedupRules`
- `getRangesConfig`

First, we removed `podGateways`:
```go
podGateways       []net.IP
```
from the `kubenetNetworkPlugin` struct.

Next, we updated the `Event` method's logic:

```go
for idx, currentPodCIDR := range podCIDRs {
	_, cidr, err := net.ParseCIDR(currentPodCIDR)
	if nil != err {
		klog.Warningf("Failed to generate CNI network config with cidr %s at index:%v: %v", currentPodCIDR, idx, err)
		return
	}
	// create list of ips
	plugin.podCIDRs = append(plugin.podCIDRs, cidr)
}
```

What should stand out is the fact that we are no longer setting up the `bridge address` here or appending `gateway` values to `podGateways` in that `kubenetNetworkPlugin` struct. We've removed that altogether. This essentially removed the dependency on `pod cidrs` to derive a `gateway` value. So, how do we get the `gateway` now??

That's where these next steps come in..

Now, let's update the `setup` method. This method is responsible for setting up networking through [CNI](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/) using the given ns/name and sandbox ID. Let's start off by creating some variables representing lists of type `podGateways` and type `podCidrs`:
```go
var podGateways []net.IP
var podCIDRs []net.IPNet
```

We can update these lists based on whether or not it is an IP4 or IP6 address we're adding:
```go
//TODO: v1.16 (khenidak) update NET_CONFIG_TEMPLATE to CNI version 0.3.0 or later so
// that we get multiple IP addresses in the returned Result structure
if res.IP4 != nil {
	ipv4 = res.IP4.IP.IP.To4()
	podGateways = append(podGateways, res.IP4.Gateway)
	podCIDRs = append(podCIDRs, net.IPNet{IP: ipv4.Mask(res.IP4.IP.Mask), Mask: res.IP4.IP.Mask})
}

if res.IP6 != nil {
	ipv6 = res.IP6.IP.IP
	podGateways = append(podGateways, res.IP6.Gateway)
	podCIDRs = append(podCIDRs, net.IPNet{IP: ipv6.Mask(res.IP6.IP.Mask), Mask: res.IP6.IP.Mask})
}
```

for reference, `res` is a variable defined earlier in this `setup` method as:
```go
// Coerce the CNI result version
res, err := cnitypes020.GetResult(resT)
```

so, when we say
```go
if res.IP4 != nil
```
_or_
```go
if res.IP6 != nil
```
this checks the CNI result to see which IP address type is returned.

Then, at the bottom of this `setup` method, we make a call to a method that eliminates duplicate packets by configuring the rules for [ebtables](https://ebtables.netfilter.org/):
```go
// configure the ebtables rules to eliminate duplicate packets by best effort
plugin.syncEbtablesDedupRules(link.Attrs().HardwareAddr, podCIDRs, podGateways)
```

If you have a sharp eye, you may have noticed the change in the `syncEbtablesDedupRules` method signature. Here's how it was done previously:
```go
plugin.syncEbtablesDedupRules(link.Attrs().HardwareAddr)
```

That is because we need `podCIDRs` and `podGateways` for this `syncEbtablesDedupRules` method when we do the following:
```go
// per gateway rule
for idx, gw := range podGateways {
	klog.V(3).Infof("Filtering packets with ebtables on mac address: %v, gateway: %v, pod CIDR: %v", macAddr.String(), gw.String(), podCIDRs[idx].String())

	bIsV6 := netutils.IsIPv6(gw)
	IPFamily := "IPv4"
	ipSrc := "--ip-src"
	if bIsV6 {
		IPFamily = "IPv6"
		ipSrc = "--ip6-src"
	}
	commonArgs := []string{"-p", IPFamily, "-s", macAddr.String(), "-o", "veth+"}
	_, err = plugin.ebtables.EnsureRule(utilebtables.Prepend, utilebtables.TableFilter, dedupChain, append(commonArgs, ipSrc, gw.String(), "-j", "ACCEPT")...)
	if err != nil {
		klog.Errorf("Failed to ensure packets from cbr0 gateway:%v to be accepted with error:%v", gw.String(), err)
		return

	}
	_, err = plugin.ebtables.EnsureRule(utilebtables.Append, utilebtables.TableFilter, dedupChain, append(commonArgs, ipSrc, podCIDRs[idx].String(), "-j", "DROP")...)
	if err != nil {
		klog.Errorf("Failed to ensure packets from podCidr[%v] but has mac address of cbr0 to get dropped. err:%v", podCIDRs[idx].String(), err)
		return
	}
}
```

The changes made above were to reflect the change in setup where we defined `podCIDRs` as opposed to appending to `plugin.podCIDRs`, which was the `podCIDRs` on the `kubenetNetworkPlugin` struct. Here's the diff for that change:
```go
klog.V(3).Infof("Filtering packets with ebtables on mac address: %v, gateway: %v, pod CIDR: %v", macAddr.String(), gw.String(), plugin.podCIDRs[idx].String()))
```
_to_
```go
klog.V(3).Infof("Filtering packets with ebtables on mac address: %v, gateway: %v, pod CIDR: %v", macAddr.String(), gw.String(), podCIDRs[idx].String())
```
And similarly, a change was made in how we error. We've gone from:
```go
_, err = plugin.ebtables.EnsureRule(utilebtables.Append, utilebtables.TableFilter, dedupChain, append(commonArgs, ipSrc, plugin.podCIDRs[idx].String(), "-j", "DROP")...)
if err != nil {
	klog.Errorf("Failed to ensure packets from podCidr[%v] but has mac address of cbr0 to get dropped. err:%v", plugin.podCIDRs[idx].String(), err)
	return
}
```
_to_
```go
_, err = plugin.ebtables.EnsureRule(utilebtables.Append, utilebtables.TableFilter, dedupChain, append(commonArgs, ipSrc, podCIDRs[idx].String(), "-j", "DROP")...)
if err != nil {
	klog.Errorf("Failed to ensure packets from podCidr[%v] but has mac address of cbr0 to get dropped. err:%v", podCIDRs[idx].String(), err)
	return
}
```

Again, updating `plugin.podCIDRs` to be `podCIDRs` instead.

Finally, let's hop out of the `setup` method and jump to `getRangesConfig`. This was a small method that was updated as well. It gets referenced in the `Event` method that we've talked about earlier. In `Event`, it is used to make the json output for the CNI network config:
```go
json := fmt.Sprintf(NET_CONFIG_TEMPLATE, BridgeName, plugin.mtu, network.DefaultInterfaceName, setHairpin, plugin.getRangesConfig(), plugin.getRoutesConfig())
klog.V(4).Infof("CNI network config set to %v", json)
plugin.netConfig, err = libcni.ConfFromBytes([]byte(json))
```

The `getRangesConfig` method went from this:
```go
// given a n cidrs assigned to nodes,
// create bridge configuration that conforms to them
func (plugin *kubenetNetworkPlugin) getRangesConfig() string {
	createRange := func(thisNet *net.IPNet) string {
		template := `
[{
"subnet": "%s",
"gateway": "%s"
}]`
		return fmt.Sprintf(template, thisNet.String(), thisNet.IP.String())
	}

	ranges := make([]string, len(plugin.podCIDRs))
	for idx, thisCIDR := range plugin.podCIDRs {
		ranges[idx] = createRange(thisCIDR)
	}
	//[{range}], [{range}]
	// each range is a subnet and a gateway
	return strings.Join(ranges[:], ",")
}
```
_to_
```go
func (plugin *kubenetNetworkPlugin) getRangesConfig() string {
	createRange := func(thisNet *net.IPNet) string {
		template := `
[{
"subnet": "%s"
}]`
		return fmt.Sprintf(template, thisNet.String())
	}

	ranges := make([]string, len(plugin.podCIDRs))
	for idx, thisCIDR := range plugin.podCIDRs {
		ranges[idx] = createRange(thisCIDR)
	}
	//[{range}], [{range}]
	// each range contains a subnet. gateway will be fetched from cni result
	return strings.Join(ranges[:], ",")
}
```
