package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/replicatedhq/kubectl-traceroute/pkg/logger"
	"github.com/replicatedhq/kubectl-traceroute/pkg/traceroute"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

			log := logger.NewLogger()

			namespace := viper.GetString("namespace")
			if namespace == "" {
				namespace = "default"
			}

			t := traceroute.Traceroute{
				Namespace:           namespace,
				OriginalServiceName: args[0],
			}

			if err := t.Prepare(KubernetesConfigFlags, log); err != nil {
				log.Error(err)
				return err
			}

			if err := t.Run(); err != nil {
				log.Error(err)
				return err
			}

			if !t.Success {
				os.Exit(1)
			}

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
