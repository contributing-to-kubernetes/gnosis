/*
  This is where the main startup logic of the API server lives.
  The flow is that we first defined a cobra command in NewAPIServerCommand.
  This command will parse and define command-line flags (we only care about the
  --service-cluster-ip-range flag).
  Then we will mimick the real kube-apiserver's logic to parse and use this
  value.

  Inspired by k/k/cmd/kube-apiserver/app/server.go.
*/
package app

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/apiserver/pkg/util/term"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog"

	"github.com/contributing-to-kubernetes/gnosis/stories/kk-pr-85968/example-cobra/cmd/app/options"
	"github.com/contributing-to-kubernetes/gnosis/stories/kk-pr-85968/example-cobra/pkg/master"
	utilflag "github.com/contributing-to-kubernetes/gnosis/stories/kk-pr-85968/example-cobra/pkg/util/flag"
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
		// Run but returns an error.
		RunE: func(cmd *cobra.Command, args []string) error {
			utilflag.PrintFlags(cmd.Flags())

			// set default options
			completedOptions, err := Complete(s)
			if err != nil {
				return err
			}

			return Run(completedOptions)
		},
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	//options.AddCustomGlobalFlags(namedFlagSets.FlagSet("generic"))
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	return cmd
}

// Run runs the specified APIServer.
func Run(completeOptions completedServerRunOptions) error {
	klog.Infof("Server run options %#v\n", completeOptions.ServerRunOptions)
	return nil
}

// completedServerRunOptions is a private wrapper that enforces a call of Complete() before Run can be invoked.
type completedServerRunOptions struct {
	*options.ServerRunOptions
}

// Complete set default ServerRunOptions.
// Should be called after kube-apiserver flags parsed.
func Complete(s *options.ServerRunOptions) (completedServerRunOptions, error) {
	var options completedServerRunOptions

	// process s.ServiceClusterIPRange from list to Primary and Secondary
	// we process secondary only if provided by user
	apiServerServiceIP, primaryServiceIPRange, secondaryServiceIPRange, err := getServiceIPAndRanges(s.ServiceClusterIPRanges)
	if err != nil {
		return options, err
	}
	s.PrimaryServiceClusterIPRange = primaryServiceIPRange
	s.SecondaryServiceClusterIPRange = secondaryServiceIPRange
	// Just for kicks.
	s.APIServerServiceIP = apiServerServiceIP

	options.ServerRunOptions = s
	return options, nil
}

// This is where the main action happens and where the bug in
// https://github.com/kubernetes/kubernetes/pull/85937 was resolved.
func getServiceIPAndRanges(serviceClusterIPRanges string) (net.IP, net.IPNet, net.IPNet, error) {
	serviceClusterIPRangeList := []string{}
	// Here is where the bug discovered in
	// https://github.com/kubernetes/kubernetes/pull/85937 is prevented!
	if serviceClusterIPRanges != "" {
		serviceClusterIPRangeList = strings.Split(serviceClusterIPRanges, ",")
	}

	var apiServerServiceIP net.IP
	var primaryServiceIPRange net.IPNet
	var secondaryServiceIPRange net.IPNet
	var err error
	// nothing provided by user, use default range (only applies to the Primary)
	if len(serviceClusterIPRangeList) == 0 {
		var primaryServiceClusterCIDR net.IPNet
		primaryServiceIPRange, apiServerServiceIP, err = master.ServiceIPRange(primaryServiceClusterCIDR)
		if err != nil {
			return net.IP{}, net.IPNet{}, net.IPNet{}, fmt.Errorf("error determining service IP ranges: %v", err)
		}
		return apiServerServiceIP, primaryServiceIPRange, net.IPNet{}, nil
	}

	if len(serviceClusterIPRangeList) > 0 {
		_, primaryServiceClusterCIDR, err := net.ParseCIDR(serviceClusterIPRangeList[0])
		if err != nil {
			return net.IP{}, net.IPNet{}, net.IPNet{}, fmt.Errorf("service-cluster-ip-range[0] is not a valid cidr")
		}

		primaryServiceIPRange, apiServerServiceIP, err = master.ServiceIPRange(*(primaryServiceClusterCIDR))
		if err != nil {
			return net.IP{}, net.IPNet{}, net.IPNet{}, fmt.Errorf("error determining service IP ranges for primary service cidr: %v", err)
		}
	}

	return apiServerServiceIP, primaryServiceIPRange, secondaryServiceIPRange, nil
}
