package caches

import (
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"reflect"
)

type localItemsCache[T CachedItem] struct {
	initialized bool
	cache       map[string]T
}

func (c *localItemsCache[T]) IsInitialized() bool {
	if c.cache == nil {
		return false
	}

	return c.initialized
}

func (c *localItemsCache[T]) Initialize(items []T) error {
	if c.cache == nil {
		c.cache = make(map[string]T)
	}

	for _, entity := range items {
		c.cache[entity.GetCacheKey()] = entity
	}

	c.initialized = true
	return nil
}

func (c *localItemsCache[T]) Get(key string) (T, bool) {
	var result T
	if !c.IsInitialized() {
		return result, false
	}

	result, ok := c.cache[key]
	return result, ok
}

func (c *localItemsCache[T]) GetAll(result chan T) {
	if !c.IsInitialized() {
		close(result)
		return
	}

	for _, value := range c.cache {
		result <- value
	}
	close(result)
}

func (c *localItemsCache[T]) GetAllAsArray() ([]T, error) {
	if !c.IsInitialized() {
		return nil, errors.New("cache not initialized")
	}

	result := make([]T, 0, len(c.cache))
	for _, value := range c.cache {
		result = append(result, value)
	}

	return result, nil
}

func (c *localItemsCache[T]) Store(entity T) error {
	if !c.IsInitialized() {
		return errors.New("cache not initialized")
	}

	c.cache[entity.GetCacheKey()] = entity
	return nil
}

func (c *localItemsCache[T]) StoreSafeBackground(entity T) {
	err := c.Store(entity)
	if err != nil {
		typeName := reflect.TypeOf(entity).Name()
		logs.Severe(fmt.Sprintf("safe background store failed, cache for '%s' dead, %v", typeName, err))
		c.Clear()
	}
}

func (c *localItemsCache[T]) Remove(id string) error {
	if !c.IsInitialized() {
		return errors.New("cache not initialized")
	}
	delete(c.cache, id)
	return nil
}

func (c *localItemsCache[T]) Clear() {
	c.cache = nil
}
