package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// AddToScheme adds all types to the given scheme.
func AddToScheme(s *runtime.Scheme) error {
	if err := localSchemeBuilder.AddToScheme(s); err != nil {
		return err
	}
	return nil
}
