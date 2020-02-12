package main

import (
	"github.com/replicatedhq/kubectl-traceroute/cmd/traceroute/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // required for GKE
)

func main() {
	cli.InitAndExecute()
}
