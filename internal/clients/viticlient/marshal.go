package viticlient

import (
	"github.com/vitistack/common/pkg/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func MarshalMachineObjects(unstructured []unstructured.Unstructured) ([]*v1alpha1.Machine, error) {
	output := make([]*v1alpha1.Machine, len(unstructured))
	for index, resource := range unstructured {
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(resource.Object, output[index])
		if err != nil {
			return nil, err
		}
	}
	return output, nil
}
