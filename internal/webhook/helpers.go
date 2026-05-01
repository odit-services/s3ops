package webhook

import (
	"fmt"
	"os"
	"strconv"

	s3oditservicesv1alpha1 "github.com/odit-services/s3ops/api/v1alpha1"
)

func validateServerRefCrossNamespace(serverRef s3oditservicesv1alpha1.ServerReference, resourceNamespace string) error {
	enableCrossNS := os.Getenv("ENABLE_CROSS_NAMESPACE_VALIDATION")
	enable, _ := strconv.ParseBool(enableCrossNS)

	if !enable && serverRef.Namespace != "" && serverRef.Namespace != resourceNamespace {
		return fmt.Errorf("cross-namespace ServerRef is not allowed: server %q is in namespace %q but resource is in namespace %q. Set ENABLE_CROSS_NAMESPACE_VALIDATION=true to allow",
			serverRef.Name, serverRef.Namespace, resourceNamespace)
	}
	return nil
}
