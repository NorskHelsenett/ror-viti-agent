package viticlient

import (
	"fmt"
	"log/slog"

	"github.com/vitistack/common/pkg/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func AddFunc(obj any) {

	var machine v1alpha1.Machine
	structured, ok := obj.(*unstructured.Unstructured)
	if !ok {
		slog.Error("failed to cast to unstructured.Unstructured")
		return
	}
	err := MarshalMachineObject(structured, &machine)
	if err != nil {
		slog.Error("failed to cast into machine")
		return
	}
	slog.Info("added machine", "name", machine.Name)
}
func UpdateFunc(oldObj, newObj any) {
	oldMachine, ok := oldObj.(v1alpha1.Machine)
	if !ok {
		panic(fmt.Errorf("failed to cast to oldMachine"))
	}

	newMachine, ok := newObj.(v1alpha1.Machine)
	if !ok {
		panic(fmt.Errorf("failed to cast to newMachine"))
	}
	slog.Info("updated machine", "old_name", oldMachine.Name, "new_machine", newMachine.Name)

}
func DeleteFunc(obj any) {
	machine, ok := obj.(v1alpha1.Machine)
	if !ok {
		panic(fmt.Errorf("failed to cast to machine"))
	}
	slog.Info("deleted  machine", "name", machine.Name)

}
