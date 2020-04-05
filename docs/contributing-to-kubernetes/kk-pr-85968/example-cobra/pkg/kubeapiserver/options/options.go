/*
  Inspired by k/k/pkg/kubeapiserver/options/options.go.
*/
package options

import (
	"net"
)

// DefaultServiceIPCIDR is a CIDR notation of IP range from which to allocate service cluster IPs
var DefaultServiceIPCIDR net.IPNet = net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}
