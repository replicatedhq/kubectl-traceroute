package traceroute

import (
	"fmt"
	"strings"
)

func FQDNForArg(namespace string, serviceName string) (string, error) {
	serviceNameParts := strings.Split(serviceName, ":")
	return fmt.Sprintf("%s.%s.svc.cluster.local", serviceNameParts[0], namespace), nil
}
