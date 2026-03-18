package viticlient

import (
	"log/slog"
	"reflect"

	"github.com/vitistack/common/pkg/v1alpha1"
)

func AddFunc(obj any) {

	var machine v1alpha1.Machine
	err := MarshalAnyMachineObject(obj, &machine)
	if err != nil {
		slog.Error("failed to marshal", "error", err)
		return
	}
	slog.Info("added machine", "name", machine.Name)
}

func UpdateFunc(oldObj, newObj any) {
	var oldMachine v1alpha1.Machine
	err := MarshalAnyMachineObject(oldObj, &oldMachine)
	if err != nil {
		slog.Error("failed to marshal old machine", "error", err)
		return
	}

	var newMachine v1alpha1.Machine
	err = MarshalAnyMachineObject(newObj, &newMachine)
	if err != nil {
		slog.Error("failed to marshal old machine", "error", err)
		return
	}

	equality := reflect.DeepEqual(newMachine, oldMachine)
	slog.Info("updated machine", "old_name", oldMachine.Name, "new_machine", newMachine.Name, "equality", equality)

}

func DeleteFunc(obj any) {
	var machine v1alpha1.Machine
	err := MarshalAnyMachineObject(obj, &machine)
	if err != nil {
		slog.Error("failed to marshal", "error", err)
		return
	}
	slog.Info("deleted machine", "name", machine.Name)

}
