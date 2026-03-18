package rorclient

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/NorskHelsenett/ror-viti-agent/internal/clients/cache"
	"github.com/NorskHelsenett/ror/pkg/rorresources"
)

// RorUUID format, v5 UUID generated from util.GenerateUuidv5.
type RorUUID string

type rorcache struct {
	rorResources *cache.Cache[RorUUID, rorresources.Resource]
	context      context.Context
}

func newRorCache(ctx context.Context) *rorcache {
	return &rorcache{
		rorResources: cache.NewCache[RorUUID, rorresources.Resource](),
		context:      ctx,
	}
}

var (
	ErrExistingResource error = errors.New("resource already exists")
)

// addResource adds a resource to the cache, if the resource already exists it returns.
func (r *rorcache) addResources(resources []*rorresources.Resource) error {

	localresources := make([]rorresources.Resource, 0, len(resources))
	for _, resource := range resources {
		localresources = append(localresources, *resource)
	}

	err := r.rorResources.AddMultiple(localresources, func(resource rorresources.Resource) RorUUID {
		return RorUUID(resource.GetUID())
	})
	if len(err.Misses) > 0 {
		return fmt.Errorf("failed to add %d resources to cache. %s", len(err.Misses), err.Error())
	}
	return nil
}

// getResource attempts to get to get the resourceID from the cache.
func (r *rorcache) getResourceID(resourceID RorUUID) (*rorresources.Resource, error) {
	resource, err := r.rorResources.Read(resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource from cache. %w", err)
	}
	return resource, nil
}

type compareResponse struct {
	Error     error
	Update    []*rorresources.Resource
	Unchanged []*rorresources.Resource
	Delete    []*rorresources.Resource
}

func (r *rorcache) CompareRorResources(updateResources []*rorresources.Resource) compareResponse {

	output := compareResponse{
		Update:    make([]*rorresources.Resource, 0),
		Unchanged: make([]*rorresources.Resource, 0),
		Delete:    make([]*rorresources.Resource, 0),
		Error:     nil,
	}

	ids := make([]RorUUID, 0)
	for _, resource := range updateResources {
		ids = append(ids, RorUUID(resource.GetUID()))
		existing, err := r.getResourceID(RorUUID(resource.GetUID()))
		if err != nil {
			if errors.Is(err, cache.ErrCacheMiss) {
				output.Update = append(output.Update, resource)
				continue
			}
		}
		if existing != nil {
			equality := areEqualResources(resource, existing)
			if equality {
				output.Unchanged = append(output.Unchanged, resource)
			} else {
				output.Update = append(output.Update, resource)
				continue
			}
		}
	}

	expiredResources := r.rorResources.MissingKeysValue(ids)
	for _, resource := range *expiredResources {
		output.Delete = append(output.Delete, &resource)
	}

	return output
}

// areEqualResources compare 2 rorresources.Resource with eachother, if there's a difference it returns false.
func areEqualResources(resource1 *rorresources.Resource, resource2 *rorresources.Resource) bool {

	r1buffer := bytes.Buffer{}
	r2buffer := bytes.Buffer{}
	enc1 := gob.NewEncoder(&r1buffer)
	enc2 := gob.NewEncoder(&r2buffer)

	if err := enc1.Encode(resource1); err != nil {
		return false
	}

	if err := enc2.Encode(resource2); err != nil {
		return false
	}
	hash1 := sha256.Sum256(r1buffer.Bytes())
	hash2 := sha256.Sum256(r2buffer.Bytes())

	return hash1 == hash2
}
