// +build !ignore_autogenerated

/*
  This is required for deepcopy-gen to work.

This oilerplate will make a lot more sense in the context of Kubernetes, where
all the source code must have a copyright message at the top.
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	v1alpha2 "github.com/contributing-to-kubernetes/gnosis/stories/kube-apis/apis/v1alpha2"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*Cluster)(nil), (*v1alpha2.Cluster)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Cluster_To_v1alpha2_Cluster(a.(*Cluster), b.(*v1alpha2.Cluster), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*v1alpha2.Cluster)(nil), (*Cluster)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha2_Cluster_To_v1alpha1_Cluster(a.(*v1alpha2.Cluster), b.(*Cluster), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_Cluster_To_v1alpha2_Cluster(in *Cluster, out *v1alpha2.Cluster, s conversion.Scope) error {
	out.Name = in.Name
	out.Nodes = *(*[]string)(unsafe.Pointer(&in.Nodes))
	return nil
}

// Convert_v1alpha1_Cluster_To_v1alpha2_Cluster is an autogenerated conversion function.
func Convert_v1alpha1_Cluster_To_v1alpha2_Cluster(in *Cluster, out *v1alpha2.Cluster, s conversion.Scope) error {
	return autoConvert_v1alpha1_Cluster_To_v1alpha2_Cluster(in, out, s)
}

func autoConvert_v1alpha2_Cluster_To_v1alpha1_Cluster(in *v1alpha2.Cluster, out *Cluster, s conversion.Scope) error {
	out.Name = in.Name
	out.Nodes = *(*[]string)(unsafe.Pointer(&in.Nodes))
	// WARNING: in.Provider requires manual conversion: does not exist in peer-type
	return nil
}
