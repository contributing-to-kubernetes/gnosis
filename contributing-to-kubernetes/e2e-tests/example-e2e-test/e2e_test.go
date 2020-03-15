/*
  This is the entrypoint to this whole experiment.
  The structure of this mini-project emulates that of the Kubernetes e2e tests.

  The way this will be used is that we will make use of ginkgo to compile our
  tests into a binary that we can run with ginkgo. Then all we need to do is to
  run this binary somewhere within a Kubernetes cluster.
  Think of these compiled tests as a simepl Go binary that you run. As such,
  you can write your tests as if they were normal code.

  To improve the testing UX we create a couple tools that float around the e2e
  test framework.

  The test framework holds global state. We can then use the framework as a
  a wrapper for all complicated things and import it within our tests.

  To see where we actually begin running tests checkout e2e.go.
*/
package e2e

import (
	"flag"
	"os"
	"testing"

	"k8s.io/klog"

	"github.com/contributing-to-kubernetes/cool-kubernetes/contributing-to-kubernetes/e2e-tests/example-e2e-test/framework"

	// Test Sources.
	_ "github.com/contributing-to-kubernetes/cool-kubernetes/contributing-to-kubernetes/e2e-tests/example-e2e-test/apps"
)

// Flags is the flag set that AddOptions adds to. Test authors should
// also use it instead of directly adding to the global command line.
//
// Taken from k/k/test/e2e/framework/config.
var Flags = flag.NewFlagSet("", flag.ContinueOnError)

// CopyFlags ensures that all flags that are defined in the source flag
// set appear in the target flag set as if they had been defined there
// directly. From the flag package it inherits the behavior that there
// is a panic if the target already contains a flag from the source.
//
// Taken from k/k/test/e2e/framework/config.
func CopyFlags(source *flag.FlagSet, target *flag.FlagSet) {
	source.VisitAll(func(flag *flag.Flag) {
		// We don't need to copy flag.DefValue. The original
		// default (from, say, flag.String) was stored in
		// the value and gets extracted by Var for the help
		// message.
		target.Var(flag.Value, flag.Name, flag.Usage)
	})
}

// handleFlags sets up all flags and parses the command line.
//
// Inspired by k/k/test/e2e/e2e_test.go.
func handleFlags() {
	CopyFlags(Flags, flag.CommandLine)
	framework.RegisterCommonFlags(flag.CommandLine)
	flag.Parse()
}

// TestMain is our run of the mill TestMain.
// Wemainly just worry here about parsing command-line flags.
//
// Inspired by k/k/test/e2e/e2e_test.go.
func TestMain(m *testing.M) {
	// Register test flags, then parse flags.
	handleFlags()

	klog.Info("Running test main...")

	os.Exit(m.Run())
}

// TestE2E is how we setup the entrypoint for ginkgo to begin running our e2e
// tests.
//
// Inspired by k/k/test/e2e/e2e_test.go.
func TestE2E(t *testing.T) {
	RunE2ETests(t)
}
