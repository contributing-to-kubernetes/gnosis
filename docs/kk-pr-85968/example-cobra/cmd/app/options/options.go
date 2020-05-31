/*
  Inspired by k/k/cmd/kube-apiserver/app/options/options.go.
*/
package options

import (
	"net"

	cliflag "k8s.io/component-base/cli/flag"
)

// ServerRunOptions runs a kubernetes api server.
type ServerRunOptions struct {
	// ServiceClusterIPRange is mapped to input provided by user
	ServiceClusterIPRanges string
	//PrimaryServiceClusterIPRange and SecondaryServiceClusterIPRange are the results
	// of parsing ServiceClusterIPRange into actual values
	PrimaryServiceClusterIPRange   net.IPNet
	SecondaryServiceClusterIPRange net.IPNet

	// Adding this one just for kicks.
	APIServerServiceIP net.IP
}

// NewServerRunOptions creates a new ServerRunOptions object with default parameters
func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{}

	return &s
}

// Flags returns flags for a specific APIServer by section name
func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("misc")

	// TODO (khenidak) change documentation as we move IPv6DualStack feature from ALPHA to BETA
	fs.StringVar(&s.ServiceClusterIPRanges, "service-cluster-ip-range", s.ServiceClusterIPRanges, ""+
		"A CIDR notation IP range from which to assign service cluster IPs. This must not "+
		"overlap with any IP ranges assigned to nodes for pods.")

	return fss
}
