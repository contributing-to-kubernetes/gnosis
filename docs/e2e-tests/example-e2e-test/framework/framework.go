package framework

import (
	"flag"
	"github.com/onsi/ginkgo/config"
)

type TestContextType struct {
	Name      string
	Namespace string
}

var TestContext TestContextType

// RegisterCommonFlags registers flags common to all e2e test suites.
// The flag set can be flag.CommandLine (if desired).
//
// For tests that have been converted to registering their
// options themselves, copy flags from test/e2e/framework/config
// as shown in HandleFlags.
//
// Inspired by k/k/test/e2e/framework/test_context.go.
func RegisterCommonFlags(flags *flag.FlagSet) {
	// Turn on verbose by default to get spec names
	config.DefaultReporterConfig.Verbose = true

	// Turn on EmitSpecProgress to get spec progress (especially on interrupt)
	config.GinkgoConfig.EmitSpecProgress = true

	// Randomize specs as well as suites
	config.GinkgoConfig.RandomizeAllSpecs = true

	flags.StringVar(&TestContext.Name, "test-flag", "custom-framework", "Sample name for this framework.")

	flags.StringVar(&TestContext.Namespace, "ns", "test", "Test namespace.")
}
