package main

import (
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	//"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	basePath = "http://127.0.0.1:9000/polycube/v1"

	vPodsRangeDefault            = "10.10.0.0/16"
	vtepsRangeDefault            = "10.18.0.0/16"
	serviceClusterIPRangeDefault = "10.96.0.0/12"
	serviceNodePortRangeDefault  = "30000-32767"
)

var (
	// connection to the k8s apif
	clientset *kubernetes.Clientset

	stop bool
)

func main() {
	stop = false

	fmt.Println("Cluster Blueprint started.")

	kubeconfig := "/home/elis/.kube/kubeconfig.conf"

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	var err1 error
	clientset, err1 = kubernetes.NewForConfig(config)
	if err1 != nil {
		panic(err1.Error())
	}

	//	Get all Deployments
	deploymentsClient := clientset.AppsV1().Deployments(meta_v1.NamespaceAll)

	deploymentsList, err := deploymentsClient.List(meta_v1.ListOptions{})
	if err != nil {
		fmt.Println("Could not list deployments:", err)
		return
	}

	for _, deployment := range deploymentsList.Items {
		fmt.Println("Deployment:", deployment.Name, "on", deployment.Namespace)
	}

}
