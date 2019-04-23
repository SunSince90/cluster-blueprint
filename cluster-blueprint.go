package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	types_core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types_app_v1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	core_v1 "k8s.io/api/core/v1"

	apps_v1 "k8s.io/api/apps/v1"

	"math/rand"
	"time"

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
	clientset         *kubernetes.Clientset
	deploymentsClient types_app_v1.DeploymentInterface
	serviceClient     types_core_v1.ServiceInterface
	nodePortRange     = []int32{int32(30000), int32(32767)}
	data              Data

	stop bool
)

type Data struct {
	Deployments []DepSer
}

type DepSer struct {
	Deployment apps_v1.Deployment
	Service    core_v1.Service
}

func main() {
	stop = false

	fmt.Println("Cluster Blueprint started.")

	kubeconfig := "/home/elis/.kube/config"

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

	//	The clients
	deploymentsClient = clientset.AppsV1().Deployments(meta_v1.NamespaceDefault)
	serviceClient = clientset.CoreV1().Services(meta_v1.NamespaceDefault)

	//	The data
	data = Data{}

	//	Start
	cmd := ""
	reader := bufio.NewReader(os.Stdin)
	for cmd != "exit" {
		fmt.Println("Enter commands.")
		cmd, _ = reader.ReadString('\n')
		cmd = strings.TrimRight(cmd, "\r\n")
		params := strings.Split(cmd, " ")

		switch params[0] {
		case "add":
			d, err := getDeployment(params)
			if err == nil {
				data.Deployments = append(data.Deployments, d)
			}

		case "go":
			deploy()
		}
	}

	/*for cmd != "exit" {

		cmd, _ = reader.ReadString('\n')
		cmd = strings.TrimRight(cmd, "\r\n")
		params := strings.Split(cmd, " ")
		switch params[0] {
		case "add":
			d, err := getDeployment(params)
			if err == nil {
				data.Deployments = append(data.Deployments, d)
			}

		case "go":
			deploy(data)
		}
	}*/
}

func generateRandomPort() int32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int31n(nodePortRange[1]-nodePortRange[0]) + nodePortRange[0]
}

func getDeployment(params []string) (DepSer, error) {

	if len(params) != 4 {
		fmt.Println("Not enough parameters")
		return DepSer{}, errors.New("Not enough parameters")
	}

	oneReplica := int32(1)
	image := params[2]
	name := params[3]

	//	Generate a random port number
	//portNumber := generateRandomPort()
	portNumber := int32(80)

	//	Do the service
	service := core_v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: core_v1.ServiceSpec{
			Ports: []core_v1.ServicePort{
				core_v1.ServicePort{
					Port: portNumber,
				},
			},
			Selector: map[string]string{
				"app": name,
			},
			ClusterIP: "",
		},
	}

	//	Do the deployment
	deployment := apps_v1.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: apps_v1.DeploymentSpec{
			Selector: &meta_v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Replicas: &oneReplica,
			Template: core_v1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{
					Name:      name,
					Namespace: "default",
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: core_v1.PodSpec{
					Containers: []core_v1.Container{
						core_v1.Container{
							Image:           image,
							Name:            name,
							ImagePullPolicy: core_v1.PullAlways,
							Ports: []core_v1.ContainerPort{
								core_v1.ContainerPort{
									ContainerPort: portNumber,
									Name:          name,
								},
							},
							/*VolumeMounts: []core_v1.VolumeMount{
								core_v1.VolumeMount{
									Name:      name + "-persistent-storage",
									MountPath: "/var/lib/" + name,
								},
							},*/
						},
					},
					/*Volumes: []core_v1.Volume{
						core_v1.Volume{
							Name: name + "-persistent-storage",
						},
					},*/
				},
			},
		},
	}
	d := DepSer{
		Deployment: deployment,
		Service:    service,
	}

	return d, nil
}

func deploy() {
	fmt.Println("deploying...")
	for _, deployments := range data.Deployments {
		//	Deploy the service
		if _, err := serviceClient.Create(&deployments.Service); err != nil {
			fmt.Println("Error in deploy service", deployments.Service.Name, ":", err.Error())
		} else {
			fmt.Println("Service", deployments.Service.Name, "deployed!")
		}

		//	Deploy the deployment
		if _, err := deploymentsClient.Create(&deployments.Deployment); err != nil {
			fmt.Println("Error in deploy deployment", deployments.Deployment.Name, ":", err.Error())
		} else {
			fmt.Println("Deployment", deployments.Deployment.Name, "deployed!")
		}
	}
}
