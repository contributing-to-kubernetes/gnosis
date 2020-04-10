/*
  Welcome!

  Here is where we give control over to ginkgo to begin running our e2e tests!
*/
package e2e

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/klog"

	"github.com/contributing-to-kubernetes/gnosis/stories/e2e-tests/example-e2e-test/framework"
)

// RunID is a unique identifier of the e2e run.
// Beware that this ID is not the same for all tests in the e2e run, because each Ginkgo node creates it separately.
//
// Taken from k/k/test/e2e/framework/util.go.
var RunID = uuid.NewUUID()

// RunE2ETests can be used to setup your ginkgo environment.
// In our case we just register the default failure handler and let test run
// wild.
//
// Inspired by from k/k/test/e2e.go.
func RunE2ETests(t *testing.T) {
	// Test failures will call the following handler.
	gomega.RegisterFailHandler(ginkgo.Fail)

	klog.Infof("Running with context %v", framework.TestContext.Name)
	klog.Infof("Starting e2e run %q on Ginkgo node %d", RunID, config.GinkgoConfig.ParallelNode)
	ginkgo.RunSpecs(t, "Kubernetes e2e suite")
}
