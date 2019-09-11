package main

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var (
	ns string = "default"
)

func createPod(name, ns string) *corev1.Pod {
	labels := make(map[string]string)
	labels["app"] = name

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "main",
					Image:   "ubuntu:18.04",
					Command: []string{"sleep", "3600"},
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

	// Create informer to watch for any new pods that come up.
	// See https://godoc.org/k8s.io/client-go/tools/cache#NewInformer
	podNamesSeen := []string{}
	_, podController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				obj, err := clientset.CoreV1().Pods(ns).List(options)
				return runtime.Object(obj), err
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return clientset.CoreV1().Pods(ns).Watch(options)
			},
		},
		&corev1.Pod{},
		time.Millisecond*100,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if pod, ok := obj.(*corev1.Pod); ok {
					fmt.Printf(" - Pod %s %d\n", pod.Name, len(podNamesSeen))
					podNamesSeen = append(podNamesSeen, pod.Name)
				}
			},
		},
	)

	stopCh := make(chan struct{})
	go podController.Run(stopCh)
	defer close(stopCh)

	i := 0
	for {
		// Create a Pod.
		podName := fmt.Sprintf("demo-pod-%d", i)
		pod := createPod(podName, ns)
		_, err := clientset.CoreV1().Pods(ns).Create(pod)
		fmt.Printf("Created Pod %s\n", podName)
		if err != nil {
			panic(err)
		}

		time.Sleep(30 * time.Second)
		i++
	}
}
