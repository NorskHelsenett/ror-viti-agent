package converter

import (
	"errors"
	"strings"

	"github.com/NorskHelsenett/ror-viti-agent/internal/config"
	"github.com/NorskHelsenett/ror/pkg/models/aclmodels"
	"github.com/NorskHelsenett/ror/pkg/models/aclmodels/rorresourceowner"
	"github.com/NorskHelsenett/ror/pkg/rorresources"
	"github.com/NorskHelsenett/ror/pkg/rorresources/rortypes"
	"github.com/google/uuid"
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
		Spec:   rortypes.ResourceMachineSpec{},
		Status: *convertRorMachineStatus(vitimachine),
	}
	rorresource := rorresources.NewRorResource("machine", "machine.ror.internal/v1alpha1")
	rorresource.RorMeta.Ownerref = rorresourceowner.RorResourceOwnerReference{
		Scope:   aclmodels.Acl2ScopeMachine,
		Subject: aclmodels.Acl2Subject(machine.Status.ProviderStatus.ProviderID),
	}

	rorresource.Metadata.UID = types.UID(GenerateUuidForMachine(*vitimachine).String())
	rorresource.Metadata.Name = vitimachine.Name
	rorresource.RorMeta.Action = rortypes.K8sActionAdd
	rorresource.SetMachine(&machine)
	return rorresource, nil
}

func ConvertToRorMachines(vitimachines []*v1alpha1.Machine) ([]*rorresources.Resource, error) {
	output := make([]*rorresources.Resource, 0, len(vitimachines))

	for _, vitimachine := range vitimachines {
		res, err := ConvertToRorMachine(vitimachine)
		if err != nil {
			return nil, err
		}
		output = append(output, res)
	}

	return output, nil
}

func convertRorMachineStatus(vitimachine *v1alpha1.Machine) *rortypes.ResourceMachineStatus {
	status := rortypes.ResourceMachineStatus{
		ProviderStatus: &vitimachine.Status,
	}

	return &status
}

// Generates a unique id for the name given. Uses uuidv5 and uses a combination
// of availabilityzone and provider as the namespace.
// The uuid will be unique within an availabilityzone and provider
func GenerateUuidForMachine(machine v1alpha1.Machine) uuid.UUID {
	ns := strings.Join([]string{machine.Spec.ProviderConfig.Zone, machine.Spec.ProviderConfig.Name}, "_")
	namespace := uuid.NewSHA1(config.GetNamespaceId(), []byte(ns))

	uuid := uuid.NewSHA1(namespace, []byte(machine.Status.ProviderID))

	return uuid
}
