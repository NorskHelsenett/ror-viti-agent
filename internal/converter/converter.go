package converter

import (
	"errors"

	"github.com/NorskHelsenett/ror/pkg/rorresources"
	"github.com/NorskHelsenett/ror/pkg/rorresources/rortypes"
	"github.com/vitistack/common/pkg/v1alpha1"
)

// ConvertToVitiMachine converts the rormachine to the viti machine CR.
// This will ignore the status completely and only change spec.
func ConvertToVitiMachine(rormachine rorresources.Resource) (*v1alpha1.Machine, error) {
	machine := rormachine.Machine().Get()
	if machine == nil {
		return nil, errors.New("failed to get machine from ror resource, resource did not contain a machine resource")
	}

	output := v1alpha1.Machine{
		TypeMeta:   rormachine.TypeMeta,
		ObjectMeta: rormachine.GetMetadata(),
		Spec:       *machine.Spec.ProviderSpec,
		Status:     v1alpha1.MachineStatus{},
	}
	return &output, nil
}

func ConvertToRorMachine(vitimachine *v1alpha1.Machine) (*rorresources.Resource, error) {

	machine := rortypes.ResourceMachine{
		Spec:   *convertRorMachineSpec(vitimachine),
		Status: *convertRorMachineStatus(vitimachine),
	}
	rorresource := rorresources.NewRorResource("machine", "machine.ror.internal/v1alpha1")
	rorresource.SetMachine(&machine)
	return rorresource, nil
}

func convertRorMachineSpec(vitimachine *v1alpha1.Machine) *rortypes.ResourceMachineSpec {
	spec := rortypes.ResourceMachineSpec{
		ProviderSpec: &vitimachine.Spec,
	}

	return &spec
}

func convertRorMachineStatus(vitimachine *v1alpha1.Machine) *rortypes.ResourceMachineStatus {
	status := rortypes.ResourceMachineStatus{
		ProviderStatus: &vitimachine.Status,
	}

	return &status
}
