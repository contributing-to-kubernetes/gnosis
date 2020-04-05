package main

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	ns string = "default"
)

func int32Ptr(i int32) *int32 { return &i }

func createDeploy(name, ns string) *appsv1.Deployment {
	labels := make(map[string]string)
	labels["app"] = name

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "nginx:1.12",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									"cpu":    resource.MustParse("500m"),
									"memory": resource.MustParse("64m"),
								},
								Requests: corev1.ResourceList{
									"cpu":    resource.MustParse("500m"),
									"memory": resource.MustParse("64m"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// Create a dynamic client.
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create Deployment.
	fmt.Println("Creating deployment...")
	deploy := createDeploy("nginx-server", ns)
	result, err := clientset.AppsV1().Deployments(ns).Create(deploy)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	time.Sleep(5 * time.Second)

	// List deployments.
	fmt.Printf("Listing deployments in namespace %q:\n", ns)
	list, err := clientset.AppsV1().Deployments(ns).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
		fmt.Printf("   * objectkind: %v\n", d.GetObjectKind())
		fmt.Printf("   * groupversion: %v\n", d.GetObjectKind().GroupVersionKind())
	}

	deployResource := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	unstructuredObj, err := dynamicClient.Resource(deployResource).Namespace(ns).Get("nginx-server", metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(" * Getting Object kind through a dynamic client...")
	fmt.Printf("   * objectkind: %v\n", unstructuredObj.GetObjectKind())
	fmt.Printf("   * groupversion: %v\n", unstructuredObj.GetObjectKind().GroupVersionKind())

	time.Sleep(10 * time.Minute)
}
