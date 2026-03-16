package viticlient

import (
	"github.com/vitistack/common/pkg/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// MarshalMachineObjects takes a slice of unstructured.Unstructured and converts it to v1alpha1.Machine.
// It checks each resource if it matches the expected GroupVersionKind with the Machine definition,
// and converts any matches to the output.
func MarshalMachineObjects(unstructured []unstructured.Unstructured) ([]*v1alpha1.Machine, error) {
	output := make([]*v1alpha1.Machine, len(unstructured))
	for index, resource := range unstructured {
		if resource.GroupVersionKind() != NewGVRV1Alpha1Machine().GroupVersion().WithKind("Machine") {
			continue
		}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.Object, &output[index])
		if err != nil {
			return nil, err
		}
	}
	return output, nil
}
