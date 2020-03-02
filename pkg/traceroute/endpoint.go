package traceroute

import (
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func GetServiceEndpoints(clientset *kubernetes.Clientset, svc *corev1.Service) (*corev1.Endpoints, error) {
	endpoints, err := clientset.CoreV1().Endpoints(svc.Namespace).Get(svc.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get endpoints")
	}

	return endpoints, nil
}

func GetMatchingStatefulset(clientset *kubernetes.Clientset, svc *corev1.Service) (*appsv1.StatefulSet, error) {
	statefulsets, err := clientset.AppsV1().StatefulSets(svc.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list statefulsets")
	}

	matchingStatefulsets := []*appsv1.StatefulSet{}
	for _, statefulset := range statefulsets.Items {
		matchLabels := labels.Set(statefulset.Spec.Selector.MatchLabels)

		isMatch := true
		for k, v := range svc.Spec.Selector {
			if matchLabels.Get(k) != v {
				isMatch = false
				goto StopLooking
			}
		}

	StopLooking:
		if isMatch {
			matchingStatefulsets = append(matchingStatefulsets, &statefulset)
		}
	}

	if len(matchingStatefulsets) == 0 {
		return nil, nil
	}

	if len(matchingStatefulsets) > 1 {
		return nil, errors.New("too many statefulsets")
	}

	return matchingStatefulsets[0], nil
}

func GetMatchingDeployment(clientset *kubernetes.Clientset, svc *corev1.Service) (*appsv1.Deployment, error) {
	deployments, err := clientset.AppsV1().Deployments(svc.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list deployments")
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
