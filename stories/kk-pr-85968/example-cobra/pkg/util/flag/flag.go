/*
  Inspired by k/k/pkg/util/flag/flags.go.
*/
package flag

import (
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

// PrintFlags logs the flags in the flagset
func PrintFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}
