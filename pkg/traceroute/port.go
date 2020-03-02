package traceroute

import (
	"strconv"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kuberneteserrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CheckPort will validate the port exists, and return keepGoing, error
func (t *Traceroute) CheckPort() (bool, error) {
	if t.ServicePort == "" {
		return true, nil
	}

	ok, err := PortExistsOnService(t.svc, t.ServicePort)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if port exists on service")
	}

	if !ok {
		t.log.Info("  x    x    x    port %s not found on service %s", t.ServicePort, t.ServiceName)
		return false, nil
	}

	t.log.Info("  ✓    ✓    ✓    port %s found on service %s", t.ServicePort, t.ServiceName)

	return true, nil
}

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
