package v1alpha2

type Cluster struct {
	Name     string
	Nodes    []string
	Provider string
}
