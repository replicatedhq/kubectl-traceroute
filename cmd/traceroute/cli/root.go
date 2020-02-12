package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/replicatedhq/kubectl-traceroute/pkg/logger"
	"github.com/replicatedhq/kubectl-traceroute/pkg/traceroute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "traceroute SERVICE_NAME[:PORT]",
		Short:         "",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Println(cmd.UsageString())
				os.Exit(1)
			}

			// v := viper.GetViper()
			log := logger.NewLogger()
			log.Info("")

			namespace := viper.GetString("namespace")
			if namespace == "" {
				namespace = "default"
			}

			fqdn, err := traceroute.FQDNForArg(namespace, args[0])
			if err != nil {
				log.Error(err)
				return err
			}

			clientset, err := createClientset()
			if err != nil {
				log.Error(err)
				return err
			}

			log.Info("Tracing route to %s", fqdn)

			serviceNameParts := strings.Split(args[0], ":")
			serviceName := serviceNameParts[0]
			servicePort := ""
			if len(serviceNameParts) > 1 {
				servicePort = serviceNameParts[1]
			}

			svc, err := traceroute.FindService(clientset, namespace, serviceName)
			if err != nil {
				log.Error(err)
				return err
			}
			if svc == nil {
				log.Info("  x    x    x    service named %s not found in %s namespace", serviceName, namespace)
				os.Exit(1)
			}
			log.Info("  ✓    ✓    ✓    service named %s found in %s namespace", serviceName, namespace)

			if servicePort != "" {
				ok, err := traceroute.PortExistsOnService(svc, servicePort)
				if err != nil {
					log.Error(err)
					return err
				}

				if !ok {
					log.Info("  x    x    x    port %s not found on service %s", servicePort, serviceName)
					os.Exit(1)
				}

				log.Info("  ✓    ✓    ✓    port %s found on service %s", servicePort, serviceName)

			}

			deployment, err := traceroute.GetMatchingDeployment(clientset, svc)
			if err != nil {
				log.Error(err)
				return err
			}
			if deployment == nil {
				log.Info("  x    x    x    no matching deployment found")
				os.Exit(1)
			}
			log.Info("  ✓    ✓    ✓    Deployment/%s", deployment.Name)
			log.Info("  ✓    ✓    ✓    %d replicas of deployment should be present", *deployment.Spec.Replicas)

			checkCount := 0
			for checkCount < 3 {
				healthy, total, err := traceroute.GetDeploymentReplicaCount(clientset, deployment)
				if err != nil {
					log.Error(err)
					return err
				}

				if checkCount < 2 {
					log.InfoNoNewLine(" %d/%d ", healthy, total)
					time.Sleep(time.Second)
				} else {
					log.Info(" %d/%d   ready replicas of deployment found", healthy, total)
				}

				checkCount++
			}

			log.Info("")

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("TRACEROUTE")
	viper.AutomaticEnv()
}

func createClientset() (*kubernetes.Clientset, error) {
	config, err := KubernetesConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create clientset")
	}

	return clientset, nil
}
