package rorclient

import (
	"context"

	"github.com/NorskHelsenett/ror-viti-agent/internal/converter"
	"github.com/NorskHelsenett/ror/pkg/rorresources"
	"github.com/vitistack/common/pkg/v1alpha1"
)

type MachineClient struct {
	client RorClient
}

// deletes the machine in ROR
func (c *MachineClient) Delete(ctx context.Context, machine v1alpha1.Machine) error {
	uuid := converter.GenerateUuidForMachine(machine)

	err := c.client.DeleteRorResource(ctx, uuid.String())
	if err != nil {
		return err
	}

	return nil
}

// updates the provider status with the status from the Machine object
func (c *MachineClient) UpdateProviderStatus(ctx context.Context, machine v1alpha1.Machine) error {

	resource, err := vitiMachineToResource(machine)
	if err != nil {
		return nil
	}

	slice := []*rorresources.Resource{resource}

	err = c.client.UpdateRorResources(slice)
	if err != nil {
		return nil
	}

	return nil
}

func vitiMachineToResource(machine v1alpha1.Machine) (*rorresources.Resource, error) {
	resource, err := converter.ConvertToRorMachine(&machine)
	if err != nil {
		return nil, err
	}

	return resource, nil
}
