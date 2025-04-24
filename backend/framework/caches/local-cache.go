package caches

import (
	"errors"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"reflect"
	"time"
)

const ErrCacheNotInitialized = "cache not initialized"

type localCachedItem[T CachedItem] struct {
	Item     T
	CachedAt time.Time
}

type localItemsCache[T CachedItem] struct {
	initialized bool
	lifetime    time.Duration
	hasLifetime bool
	cache       map[string]localCachedItem[T]
}

func (c *localItemsCache[T]) IsInitialized() bool {
	if c.cache == nil {
		return false
	}

	return c.initialized
}

func (c *localItemsCache[T]) Initialize(items []T) error {
	if c.cache == nil {
		c.cache = make(map[string]localCachedItem[T])
	}

	for _, entity := range items {
		c.cache[entity.GetCacheKey()] = localCachedItem[T]{
			Item:     entity,
			CachedAt: time.Now(),
		}
	}

	c.initialized = true
	return nil
}

func (c *localItemsCache[T]) Get(key string) (T, bool) {
	var result T
	if !c.IsInitialized() {
		return result, false
	}

	item, ok := c.cache[key]
	if !ok {
		return result, ok
	}

	return item.Item, c.isItemValid(item)
}

func (c *localItemsCache[T]) GetAll(result chan T) {
	if !c.IsInitialized() {
		close(result)
		return
	}

	for _, value := range c.cache {
		if c.isItemValid(value) {
			result <- value.Item
		}
	}
	close(result)
}

func (c *localItemsCache[T]) GetAllAsArray() ([]T, error) {
	if !c.IsInitialized() {
		return nil, errors.New(ErrCacheNotInitialized)
	}

	result := make([]T, 0, len(c.cache))
	for _, value := range c.cache {
		if c.isItemValid(value) {
			result = append(result, value.Item)
		}
	}

	return result, nil
}

func (c *localItemsCache[T]) Store(entity T) error {
	if !c.IsInitialized() {
		return errors.New(ErrCacheNotInitialized)
	}

	c.cache[entity.GetCacheKey()] = localCachedItem[T]{
		Item:     entity,
		CachedAt: time.Now(),
	}
	return nil
}

func (c *localItemsCache[T]) StoreSafeBackground(entity T) {
	err := c.Store(entity)
	if err != nil {
		typeName := reflect.TypeOf(entity).Name()
		logs.Error(nil, "safe background store failed, cache for '%s' dead, %v", typeName, err)
		c.Clear()
	}
}

func (c *localItemsCache[T]) Remove(id string) error {
	if !c.IsInitialized() {
		return errors.New(ErrCacheNotInitialized)
	}
	delete(c.cache, id)
	return nil
}

func (c *localItemsCache[T]) Clear() {
	c.cache = nil
}

func (c *localItemsCache[T]) SetLifetime(duration time.Duration) {
	c.hasLifetime = true
	c.lifetime = duration
}

func (c *localItemsCache[T]) isItemValid(item localCachedItem[T]) bool {
	if !c.hasLifetime {
		return true
	}

	lastValidAt := time.Now().Add(-c.lifetime)
	itemValid := lastValidAt.After(item.CachedAt)

	if !itemValid {
		delete(c.cache, item.Item.GetCacheKey())
	}

	return itemValid
}
