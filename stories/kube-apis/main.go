package main

import (
	"fmt"

	apiv1 "github.com/contributing-to-kubernetes/gnosis/stories/kube-apis/apis/v1alpha1"
)

func main() {
	c1 := &apiv1.Cluster{Name: "kube", Nodes: []string{"1", "2", "3"}}
	fmt.Printf("cluster1: %#v\n", c1)

	c2 := c1.DeepCopy()
	fmt.Printf("cluster2: %#v\n", c2)

	c3 := &apiv1.Cluster{}
	fmt.Printf("cluster3: %#v\n", c3)
	c2.DeepCopyInto(c3)
	fmt.Printf("cluster3: %#v\n", c3)

	fmt.Printf("cluster1 and cluster3 the same: %v\n", c1 == c3)
}
