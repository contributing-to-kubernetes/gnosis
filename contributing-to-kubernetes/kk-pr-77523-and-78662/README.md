# Re-route local traffic going to the LB #77523

## Description

This write up is on [PR #77523](https://github.com/kubernetes/kubernetes/pull/77523) and [PR #78662](https://github.com/kubernetes/kubernetes/pull/78662). This PR re-routes traffic that originates from the local node, so that it goes to the Kubernetes service chain.

The goal of these PRs was to refactor the way that traffic gets routed from the local node to an external LB, so that `externalTrafficPolicy=Local` no longer breaks internal reachability. [Issue #65387](https://github.com/kubernetes/kubernetes/issues/65387).

### What are iptables?

In order to understand what the `iptables proxier` is and does, we first have to understand what `iptables` are and why they are relevant.

According to the [man pages](https://linux.die.net/man/8/iptables) description for `iptables`:

```
"Iptables and ip6tables are used to set up, maintain, and inspect the tables of IPv4 and IPv6 packet filter rules in the Linux kernel. Several different tables may be defined. Each table contains a number of built-in chains and may also contain user-defined chains.

Each chain is a list of rules which can match a set of packets. Each rule specifies what to do with a packet that matches. This is called a 'target', which may be a jump to a user-defined chain in the same table."
```

These user-defined and/or built-in policy chains are what is used by `iptables` to allow or block incoming traffic to a system. If a connection attempts to establish itself to the system, `iptables` will reference the list of rules to confirm whether or not it should block or allow the incoming connection. If there is no rule, it will use the default behavior for these incoming traffic connections.

There is a lot of information that can be found on `iptables`. For the purpose of this write up, the above explanation should suffice for a high-level description.

### What is the iptables proxier?

In the context of `Kubernetes`, the `iptables proxier` is described as follows:
```
Proxier is an `iptables` based proxy that is used for connections between a localhost:lport and services that provide the actual backends.
```

In other words, `iptables` connection rules are relied upon for connections between local traffic and the service chain.

## New Logic Breakdown

The following code block is what was added to address this issue:

```go
// Next, redirect all src-type=LOCAL -> LB IP to the service chain for externalTrafficPolicy=Local
// This allows traffic originating from the host to be redirected to the service correctly,
// otherwise traffic to LB IPs are dropped if there are no local endpoints.
args = append(args[:0], "-A", string(svcXlbChain))
writeLine(proxier.natRules, append(args,
  "-m", "comment", "--comment", fmt.Sprintf(`"masquerade LOCAL traffic for %s LB IP"`, svcNameString),
  "-m", "addrtype", "--src-type", "LOCAL", "-j", string(KubeMarkMasqChain))...)
writeLine(proxier.natRules, append(args,
  "-m", "comment", "--comment", fmt.Sprintf(`"route LOCAL traffic for %s LB IP to service chain"`, svcNameString),
  "-m", "addrtype", "--src-type", "LOCAL", "-j", string(svcChain))...)
```

Now, let's break this down line by line.

The first line:

```go
args = append(args[:0], "-A", string(svcXlbChain))
```

appends to a list of args .....

The next line, writes a line, a.k.a. rule, to the `iptables proxier`. In the context of this code, the `iptables proxier` is represented by a struct, named `Proxier`:

```go
// Proxier is an iptables based proxy for connections between a localhost:lport
// and services that provide the actual backends.
type Proxier struct {
	// endpointsChanges and serviceChanges contains all changes to endpoints and
	// services that happened since iptables was synced. For a single object,
	// changes are accumulated, i.e. previous is state from before all of them,
	// current is state after applying all of those.
	endpointsChanges *proxy.EndpointChangeTracker
	serviceChanges   *proxy.ServiceChangeTracker

	mu           sync.Mutex // protects the following fields
	serviceMap   proxy.ServiceMap
	endpointsMap proxy.EndpointsMap
	portsMap     map[utilproxy.LocalPort]utilproxy.Closeable
	// endpointsSynced and servicesSynced are set to true when corresponding
	// objects are synced after startup. This is used to avoid updating iptables
	// with some partial data after kube-proxy restart.
	endpointsSynced bool
	servicesSynced  bool
	initialized     int32
	syncRunner      *async.BoundedFrequencyRunner // governs calls to syncProxyRules

	// These are effectively const and do not need the mutex to be held.
	iptables       utiliptables.Interface
	masqueradeAll  bool
	masqueradeMark string
	exec           utilexec.Interface
	clusterCIDR    string
	hostname       string
	nodeIP         net.IP
	portMapper     utilproxy.PortOpener
	recorder       record.EventRecorder
	healthChecker  healthcheck.Server
	healthzServer  healthcheck.HealthzUpdater

	// Since converting probabilities (floats) to strings is expensive
	// and we are using only probabilities in the format of 1/n, we are
	// precomputing some number of those and cache for future reuse.
	precomputedProbabilities []string

	// The following buffers are used to reuse memory and avoid allocations
	// that are significantly impacting performance.
	iptablesData             *bytes.Buffer
	existingFilterChainsData *bytes.Buffer
	filterChains             *bytes.Buffer
	filterRules              *bytes.Buffer
	natChains                *bytes.Buffer
	natRules                 *bytes.Buffer

	// endpointChainsNumber is the total amount of endpointChains across all
	// services that we will generate (it is computed at the beginning of
	// syncProxyRules method). If that is large enough, comments in some
	// iptable rules are dropped to improve performance.
	endpointChainsNumber int

	// Values are as a parameter to select the interfaces where nodeport works.
	nodePortAddresses []string
	// networkInterfacer defines an interface for several net library functions.
	// Inject for test purpose.
	networkInterfacer utilproxy.NetworkInterfacer
}
```

Now, if you can re-direct attention to the following line of code within the new logic:
```go
writeLine(proxier.natRules, append(args,
  "-m", "comment", "--comment", fmt.Sprintf(`"masquerade LOCAL traffic for %s LB IP"`, svcNameString),
  "-m", "addrtype", "--src-type", "LOCAL", "-j", string(KubeMarkMasqChain))...)
```

You will notice that it references something called `proxier.natRules`. As shown in the `Proxier` struct definition, it is a `*bytes.Buffer` that is used to, as the name suggests, represent `NAT rules`.

- _For those who are not entirely familiar:_
  - _[NAT](https://www.cisco.com/c/en/us/support/docs/ip/network-address-translation-nat/26704-nat-faq-00.html) or [Network Address Translation](https://www.cisco.com/c/en/us/support/docs/ip/network-address-translation-nat/26704-nat-faq-00.html) is an internet standard/process, in which a network device (i.e. a router) translates IP addresses from a `LAN` (`Local Area Network`) to a single publically available `IP address`. Therefore, `NAT rules` are the mechanism in which we can control where these `IP packets` are be routed._

  - _More info on [NAT rules can be found here.](https://web.mit.edu/rhel-doc/4/RH-DOCS/rhel-sg-en-4/s1-firewall-ipt-fwd.html)_

Continuing on, we will see that there is an `append` made to the `args` list:
```go
append(args,
  "-m", "comment", "--comment", fmt.Sprintf(`"masquerade LOCAL traffic for %s LB IP"`, svcNameString),
  "-m", "addrtype", "--src-type", "LOCAL", "-j", string(KubeMarkMasqChain))...)
```

We have a few things going on here. First, we are masquerading local traffic, for the LB IP, for the service with the name of `svcNameString`. Next, we are defining a rule that, for all traffic originating from host-local, we masquerade by adding the `mark-for-masquerade` policy chain. Which is defined earlier in this file:
```go
// the mark-for-masquerade chain
KubeMarkMasqChain utiliptables.Chain = "KUBE-MARK-MASQ"
```

This is done in order to, as pointed out earlier in the code comment, allow traffic originating from the host to be redirected to the service correctly. If this is not done, traffic to LB IPs are dropped when there are no local endpoints.

Now, we must write a `NAT rule` to route local traffic going to the LB IP, to the service chain, like so:
```go
writeLine(proxier.natRules, append(args,
  "-m", "comment", "--comment", fmt.Sprintf(`"route LOCAL traffic for %s LB IP to service chain"`, svcNameString),
  "-m", "addrtype", "--src-type", "LOCAL", "-j", string(svcChain))...)
```

where `svcChain` is defined as:
```go
svcChain := svcInfo.servicePortChainName
```

## Conclusion

In this PR, we got a glimpse into the use of `NAT rules`, `iptables` and how the `iptabels proxier`, within Kubernetes, handles connections from local traffic, to the LB, to a service.

As a community, it is important to keep on learning and sharing the knowledge we gain from our experiences.

```
"We are only as strong as we are united, as weak as we are divided."

― J.K. Rowling, Harry Potter and the Goblet of Fire
```