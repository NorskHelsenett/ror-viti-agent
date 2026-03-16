package viticlient

import "k8s.io/apimachinery/pkg/runtime/schema"

// NewGVR means NewGroupVersionResource, which is used to filter resources using
// the dynamic client.
func NewGVR(group, version, resource string) *schema.GroupVersionResource {
	return &schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
}

func NewGVRV1Namespace() *schema.GroupVersionResource {
	return NewGVR(
		"",
		"v1",
		"namespaces",
	)
}

func NewGVRV1Alpha1Machine() *schema.GroupVersionResource {
	return NewGVR(
		"vitistack.io",
		"v1alpha1",
		"machines",
	)
}
