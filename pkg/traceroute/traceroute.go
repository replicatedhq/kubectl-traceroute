package traceroute

import (
	"strconv"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
