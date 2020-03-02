package traceroute

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/replicatedhq/kubectl-traceroute/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

type Traceroute struct {
	Namespace           string
	OriginalServiceName string

	Success bool

	FQDN        string
	ServiceName string
	ServicePort string

	clientset *kubernetes.Clientset
	log       *logger.Logger

	svc                 *corev1.Service
	selectedDeployment  *appsv1.Deployment
	selectedStatefulset *appsv1.StatefulSet
}

// Prepare will parse the params of traceroute (namespace and servicename)
// and set up the traceroute to run
func (t *Traceroute) Prepare(kubernetesConfigFlags *genericclioptions.ConfigFlags, log *logger.Logger) error {
	fqdn, err := FQDNForArg(t.Namespace, t.OriginalServiceName)
	if err != nil {
		return errors.Wrap(err, "failed to parse fqdn")
	}
	t.FQDN = fqdn

	serviceNameParts := strings.Split(t.OriginalServiceName, ":")
	t.ServiceName = serviceNameParts[0]
	servicePort := ""
	if len(serviceNameParts) > 1 {
		servicePort = serviceNameParts[1]
	}
	t.ServicePort = servicePort

	config, err := kubernetesConfigFlags.ToRESTConfig()
	if err != nil {
		return errors.Wrap(err, "failed to read kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to create clientset")
	}
	t.clientset = clientset

	t.log = log

	return nil
}

func (t *Traceroute) Run() error {
	t.log.Info("Tracing route to %s", t.FQDN)

	svc, err := FindService(t.clientset, t.Namespace, t.ServiceName)
	if err != nil {
		log.Error(err)
		return err
	}
	if svc == nil {
		t.log.Info("  x    x    x    service named %s not found in %s namespace", t.ServiceName, t.Namespace)
		os.Exit(1)
	}
	t.log.Info("  ✓    ✓    ✓    service named %s found in %s namespace", t.ServiceName, t.Namespace)

	// keep a reference to this for other functions
	t.svc = svc

	keepGoing, err := t.CheckPort()
	if err != nil {
		return errors.Wrap(err, "failed to check port")
	}
	if !keepGoing {
		t.Success = false
		return nil
	}

	deployment, err := GetMatchingDeployment(t.clientset, t.svc)
	if err != nil {
		return errors.Wrap(err, "failed to get matching deployment")
	}
	t.selectedDeployment = deployment

	endpoints, err := GetServiceEndpoints(t.clientset, svc)
	if err != nil {
		log.Error(err)
		return err
	}
	if len(endpoints.Subsets) == 0 {
		t.log.Info("  x    x    x    no endpoints on service")

		// Print helpful information here
		t.log.Info("\n")
		t.log.Info(`No endpoints found mean...`)
		t.log.Info("\n")

		os.Exit(1)
	}

	readyEndpointCount := 0
	notReadyEndpointCount := 0
	for _, subset := range endpoints.Subsets {
		readyEndpointCount += len(subset.Addresses)
		notReadyEndpointCount += len(subset.NotReadyAddresses)
	}
	if readyEndpointCount == 0 {
		t.log.Info("  x    x    x    no endpoints are ready")

		// Print helpful information here
		os.Exit(1)
	}
	t.log.Info("  ✓    ✓    ✓    %d endpoint(s) exist", len(endpoints.Subsets))

	// checkCount := 0
	// for checkCount < 3 {
	// 	healthy, total, err := traceroute.GetDeploymentReplicaCount(clientset, deployment)
	// 	if err != nil {
	// 		log.Error(err)
	// 		return err
	// 	}

	// 	if checkCount < 2 {
	// 		log.InfoNoNewLine(" %d/%d ", healthy, total)
	// 		time.Sleep(time.Second)
	// 	} else {
	// 		log.Info(" %d/%d   ready replicas of deployment found", healthy, total)
	// 	}

	// 	checkCount++
	// }

	t.log.Info("")

	return nil
}
