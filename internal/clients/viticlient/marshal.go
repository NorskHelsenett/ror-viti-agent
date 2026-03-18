package viticlient

import (
	"errors"

	"github.com/vitistack/common/pkg/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	ErrUnexpectedGroupVersionKind = errors.New("unexpected group version kind")
)

// MarshalAnyMachineObject converts the object input to an unstructured.Unstructured then to a v1alpha1.Machine.
func MarshalAnyMachineObject(object any, machine *v1alpha1.Machine) error {
	structured, ok := object.(*unstructured.Unstructured)
	if !ok {
		return errors.New("failed to cast to unstructured.Unstructured")

	}
	err := MarshalMachineObject(structured, machine)
	if err != nil {
		return errors.New("failed to cast into machine")

	}
	return nil
}

// MarshalMachineObject converts unstructured.unstructured and converts it to a v1alpha1.Machine.
// It checks the input if it matches the expected GroupVersionKind with the Machine definition,
// and converts any matches to the output.
func MarshalMachineObject(unstructured *unstructured.Unstructured, machine *v1alpha1.Machine) error {
	if unstructured.GroupVersionKind() != NewGVRV1Alpha1Machine().GroupVersion().WithKind("Machine") {
		return ErrUnexpectedGroupVersionKind
	}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, machine)
	if err != nil {
		return err
	}
	return nil
}

// MarshalMachineObjects takes a slice of unstructured.Unstructured and converts it to v1alpha1.Machine.
// It checks each resource if it matches the expected GroupVersionKind with the Machine definition,
// and converts any matches to the output.
func MarshalMachineObjects(unstructured []*unstructured.Unstructured) ([]*v1alpha1.Machine, error) {
	output := make([]*v1alpha1.Machine, len(unstructured))
	for index, resource := range unstructured {
		err := MarshalMachineObject(resource, output[index])
		if err != nil {
			return nil, err
		}
	}
	return output, nil
}
