package viticlient

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Event string

const (
	MACHINE_UPDATE Event = "Updated"
	MACHINE_DELETE Event = "Delete"
	MACHINE_CREATE Event = "Create"
)

type MachineEvent struct {
	Type   Event
	Object *unstructured.Unstructured
}

type Watcher interface {
	Watch(ctx context.Context) <-chan MachineEvent
}
