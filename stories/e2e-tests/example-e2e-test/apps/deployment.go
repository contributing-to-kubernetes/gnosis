package apps

import (
	"github.com/onsi/ginkgo"

	"k8s.io/klog"

	"github.com/contributing-to-kubernetes/gnosis/stories/e2e-tests/example-e2e-test/framework"
)

var _ = SIGDescribe("Deployment", func() {
	var ns string

	ginkgo.AfterEach(func() {
		klog.Info("Running AfterEach")
	})

	ginkgo.BeforeEach(func() {
		klog.Info("Running BeforeEach")
		ns = framework.TestContext.Namespace
	})

	ginkgo.It("should be able to create", func() {
		testCreateDeployment(ns)
	})

	ginkgo.It("should be able to delete", func() {
		testDeleteDeployment(ns)
	})
})

/*
  TODO: Write actual tests!

  Here, we could make use of client-go to actually run tests instead of just
  printing a string :-)
*/

func testCreateDeployment(ns string) {
	klog.Infof("create deployment in namespace %s", ns)
}

func testDeleteDeployment(ns string) {
	klog.Infof("delete deployment in namespace %s", ns)
}
