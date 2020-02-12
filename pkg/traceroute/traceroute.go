package traceroute

import (
	"strconv"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func FindService(clientset *kubernetes.Clientset, namespace string, serviceName string) (*corev1.Service, error) {
	svc, err := clientset.CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
	if kuberneteserrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service")
	}

	return svc, nil
}

func PortExistsOnService(svc *corev1.Service, port string) (bool, error) {
	for _, servicePort := range svc.Spec.Ports {
		if servicePort.Name == port {
			return true, nil
		}

		if strconv.Itoa(int(servicePort.Port)) == port {
			return true, nil
		}
	}

	return false, nil
}

func GetMatchingDeployment(clientset *kubernetes.Clientset, svc *corev1.Service) (*appsv1.Deployment, error) {
	deployments, err := clientset.AppsV1().Deployments(svc.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to find deployments")
	}

	matchingDeployments := []*appsv1.Deployment{}
	for _, deployment := range deployments.Items {
		matchLabels := labels.Set(deployment.Spec.Selector.MatchLabels)

		isMatch := true
		for k, v := range svc.Spec.Selector {
			if matchLabels.Get(k) != v {
				isMatch = false
				goto StopLooking
			}
		}

	StopLooking:
		if isMatch {
			matchingDeployments = append(matchingDeployments, &deployment)
		}
	}

	if len(matchingDeployments) == 0 {
		return nil, nil
	}

	if len(matchingDeployments) > 1 {
		return nil, errors.New("too many deployments")
	}

	return matchingDeployments[0], nil
}

func GetDeploymentReplicaCount(clientset *kubernetes.Clientset, deployment *appsv1.Deployment) (int32, int32, error) {
	foundDeployment, err := clientset.AppsV1().Deployments(deployment.Namespace).Get(deployment.Name, metav1.GetOptions{})
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to get deplyoment")
	}

	return foundDeployment.Status.ReadyReplicas, foundDeployment.Status.Replicas, nil
}
