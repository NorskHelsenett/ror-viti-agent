package rorclient

import (
	"slices"

	"github.com/NorskHelsenett/ror/pkg/rorresources"
)

func chunkResourceToSet(resources []*rorresources.Resource, size int) []*rorresources.ResourceSet {

	output := make([]*rorresources.ResourceSet, 0)
	resourcechunks := chunkToMax(resources, size)

	for _, resourcechunk := range resourcechunks {
		set := rorresources.NewResourceSet()
		for _, resource := range resourcechunk {
			set.Add(resource)
		}
		output = append(output, set)
	}

	return output
}

// ChunkToMax slices an slice of items into chunks of maxSize.
func chunkToMax(items []*rorresources.Resource, maxSize int) [][]*rorresources.Resource {
	return slices.Collect(slices.Chunk(items, maxSize))
}
