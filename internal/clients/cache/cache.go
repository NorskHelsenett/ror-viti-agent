// Package cache provides a generic implementation of a key value cache.
// Source: https://helsegitlab.nhn.no/vm/vm-i-ror/vsphereagent/-/blob/develop/internal/cache/cache.go
package cache

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"
)

var ErrCacheMiss = errors.New("could not get value, key does not exist")
var ErrCacheNotInitialized = errors.New("map is not initialized")
var ErrInvalidKey = errors.New("cannot add value with empty or nil key")

type MissesError[K comparable] struct {
	Misses []K
}

func NewMissesError[K comparable]() MissesError[K] {

	return MissesError[K]{
		Misses: make([]K, 0),
	}
}

func (e *MissesError[K]) Error() string {
	return fmt.Sprintf("cache missed %b times", len(e.Misses))
}

type Cache[K comparable, V any] struct {
	cache   map[K]V
	mutex   sync.RWMutex
	deletes int
	limit   int
}

type Init[K comparable, V any] func(V) K

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		cache:   make(map[K]V),
		mutex:   sync.RWMutex{},
		deletes: 0,
		limit:   100,
	}
}

// Init wraps the handle multiple function
// will panic if it cannot initialize all the values.
func (c *Cache[K, V]) Init(values []V, initfunc Init[K, V]) {
	misses := c.AddMultiple(values, initfunc)
	if len(misses.Misses) > 0 {
		panic("could not initialize cache")
	}
}

// AddMultiple allows us to create multiple entries in the cache, under the
// same lock.
// The key is derived as a function of V. This makes it easier to supply
// multiple key value pairs without having to construct an external map first
//
// example:
//
//	err := s.Cache.AddMultiple(vms, func(vm rortypes.ResourceVirtualMachine) rortypes.VmId {
//		return vm.ExternalId
//	})
func (c *Cache[K, V]) AddMultiple(values []V, initfunc Init[K, V]) MissesError[K] {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var misses = NewMissesError[K]()

	for _, v := range values {
		key := initfunc(v)
		err := c.add(key, v)
		if err != nil {
			misses.Misses = append(misses.Misses, key)
		}
	}

	return misses
}

// ReadMultiple allows us to read multiple entries in the cache, under the
// same lock.
//
// example:
//
//	err := s.Cache.AddMultiple(keys)
func (c *Cache[K, V]) ReadMultiple(keys []K) (map[K]V, MissesError[K]) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var misses = NewMissesError[K]()
	var hits = make(map[K]V)

	for _, key := range keys {
		value, err := c.read(key)
		if err != nil {
			misses.Misses = append(misses.Misses, key)
		}
		hits[key] = *value
	}

	return hits, misses
}

func (c *Cache[K, V]) Add(key K, value V) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.cache == nil {
		return ErrCacheNotInitialized
	}

	return c.add(key, value)
}

func (c *Cache[K, V]) add(key K, value V) error {
	// this will not allow 0 as a key or the "" empty string
	var zero K
	if key == zero {
		return ErrInvalidKey
	}

	c.cache[key] = value
	return nil
}

func (c *Cache[K, V]) Read(key K) (*V, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.read(key)
}

func (c *Cache[K, V]) ReadFunc(resource V, getKeyFunc func(V) K) (*V, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.read(getKeyFunc(resource))
}

func (c *Cache[K, V]) read(key K) (*V, error) {

	if value, ok := c.cache[key]; ok {
		return &value, nil
	}

	return nil, ErrCacheMiss
}

func (c *Cache[K, V]) Delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, key)
	c.deletes++
	if c.deletes > c.limit {
		c.compact()
		c.deletes = 0
	}
}

func (c *Cache[K, V]) ConfirmDelete(key K) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.cache[key]; ok {
		delete(c.cache, key)
		c.deletes++
		if c.deletes > c.limit {
			c.compact()
			c.deletes = 0
		}
		return true
	}

	return false
}

// Keys output all keys in the cache.
func (c *Cache[K, V]) Keys() []K {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]K, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}

	return keys
}

// MissingKeys output any keys not in the input.
func (c *Cache[K, V]) MissingKeys(foundKeys []K) []K {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	missing := []K{}
	cachedKeys := make([]K, 0, len(c.cache))
	for k := range c.cache {
		cachedKeys = append(cachedKeys, k)
	}

	for _, key := range cachedKeys {
		if !slices.Contains(foundKeys, key) {
			missing = append(missing, key)
		}
	}

	return missing
}

// MissingKeysValue output any values of keys not in the input.
func (c *Cache[K, V]) MissingKeysValue(missingKeys []K) *[]V {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	missing := []V{}
	keys := make([]K, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}

	for _, key := range keys {
		if !slices.Contains(missingKeys, key) {
			value, _ := c.read(key)
			if value == nil {
				continue
			}
			missing = append(missing, *value)
		}
	}

	return &missing
}

func (c *Cache[K, V]) compact() {
	newData := make(map[K]V, len(c.cache))
	maps.Copy(newData, c.cache)
	c.cache = newData
}

func AddToCache[K comparable, V any](resources *[]V, cache *Cache[K, V], getIDFunc func(V) K) error {

	err := cache.AddMultiple(*resources, func(resource V) K {
		return getIDFunc(resource)
	})

	if len(err.Misses) > 0 {
		return &err
	}
	return nil
}

func DeleteFromCache[K comparable, V any](ids []K, cache *Cache[K, V]) error {

	for _, id := range ids {
		ok := cache.ConfirmDelete(id)
		if !ok {
			return fmt.Errorf("failed to delete resource %v, could not find object with id", id)
		}
	}
	return nil
}
