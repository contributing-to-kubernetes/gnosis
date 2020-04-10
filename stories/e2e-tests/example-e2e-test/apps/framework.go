/*
  Whole thing taken from k/k/test/e2e/apps.
*/
package apps

import "github.com/onsi/ginkgo"

// SIGDescribe annotates the test with the SIG label.
func SIGDescribe(text string, body func()) bool {
	return ginkgo.Describe("[sig-apps] "+text, body)
}
