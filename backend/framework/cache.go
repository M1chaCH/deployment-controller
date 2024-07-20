package framework

// TODO test with Reddis cache
type CachedItem interface {
	GetCacheKey() string
}

type ItemsCache[T CachedItem] interface {
	IsInitialized() bool
	Initialize([]T)
	Get(key string) (T, bool)
	GetAll() []T
	// Store saves the new entity to the cache
	// NOTE: it is recommended to run this in a separate go routine, since this might go to an external cache
	Store(entity T)
	Remove(id string)
}

type LocalItemsCache[T CachedItem] struct {
	cache map[string]T
}

func (c *LocalItemsCache[T]) IsInitialized() bool {
	if c.cache == nil {
		c.cache = make(map[string]T)
		return false
	}

	return len(c.cache) > 0
}

func (c *LocalItemsCache[T]) Initialize(entities []T) {
	if c.cache == nil {
		c.cache = make(map[string]T)
	}

	for _, entity := range entities {
		c.cache[entity.GetCacheKey()] = entity
	}
}

func (c *LocalItemsCache[T]) Get(key string) (T, bool) {
	var result T
	if c.cache == nil {
		c.cache = make(map[string]T)
		return result, false
	}

	result, ok := c.cache[key]
	return result, ok
}

func (c *LocalItemsCache[T]) GetAll() []T {
	if c.cache == nil {
		c.cache = make(map[string]T)
		return nil
	}

	var result []T
	for _, value := range c.cache {
		result = append(result, value)
	}
	return result
}

func (c *LocalItemsCache[T]) Store(entity T) {
	if c.cache == nil {
		c.cache = make(map[string]T)
	}

	c.cache[entity.GetCacheKey()] = entity
}

func (c *LocalItemsCache[T]) Remove(id string) {
	if c.cache == nil {
		c.cache = make(map[string]T)
		return
	}
	delete(c.cache, id)
}
