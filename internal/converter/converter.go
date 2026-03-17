package converter

import (
	"errors"

	"github.com/NorskHelsenett/ror/pkg/models/aclmodels"
	"github.com/NorskHelsenett/ror/pkg/models/aclmodels/rorresourceowner"
	"github.com/NorskHelsenett/ror/pkg/rorresources"
	"github.com/NorskHelsenett/ror/pkg/rorresources/rortypes"
	"github.com/vitistack/common/pkg/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

// ConvertToVitiMachine converts the rormachine to the viti machine CR.
// This will ignore the status completely and only change spec.
func ConvertToVitiMachine(rormachine *rorresources.Resource) (*v1alpha1.Machine, error) {
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

func ConvertToVitiMachines(rormachines []*rorresources.Resource) ([]*v1alpha1.Machine, error) {
	output := make([]*v1alpha1.Machine, len(rormachines))

	for _, rormachine := range rormachines {
		res, err := ConvertToVitiMachine(rormachine)
		if err != nil {
			return nil, err
		}
		output = append(output, res)
	}

	return output, nil
}

func ConvertToRorMachine(vitimachine *v1alpha1.Machine) (*rorresources.Resource, error) {

	machine := rortypes.ResourceMachine{
		Spec:   *convertRorMachineSpec(vitimachine),
		Status: *convertRorMachineStatus(vitimachine),
	}
	rorresource := rorresources.NewRorResource("machine", "machine.ror.internal/v1alpha1")
	rorresource.RorMeta.Ownerref = rorresourceowner.RorResourceOwnerReference{
		Scope:   aclmodels.Acl2ScopeUnknown,
		Subject: aclmodels.Acl2Subject(machine.Status.ProviderStatus.ProviderID),
	}

	rorresource.Metadata.UID = types.UID(vitimachine.Status.MachineID)
	rorresource.Metadata.Name = vitimachine.Name
	rorresource.RorMeta.Action = rortypes.K8sActionAdd
	rorresource.SetMachine(&machine)
	return rorresource, nil
}

func ConvertToRorMachines(vitimachines []*v1alpha1.Machine) ([]*rorresources.Resource, error) {
	output := make([]*rorresources.Resource, len(vitimachines))

	for _, vitimachine := range vitimachines {
		res, err := ConvertToRorMachine(vitimachine)
		if err != nil {
			return nil, err
		}
		output = append(output, res)
	}

	return output, nil
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
