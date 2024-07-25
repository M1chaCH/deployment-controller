package caches

type CachedItem interface {
	GetCacheKey() string
}

type ItemsCache[T CachedItem] interface {
	IsInitialized() bool
	Initialize(items []T) error
	Get(key string) (T, bool)
	GetAll(result chan T) // could improve performance with a redis caches, but I don't know how redis works, so might also be irrelevant
	GetAllAsArray() ([]T, error)
	// Store saves the new entity to the caches
	// NOTE: it is recommended to run this in a separate go routine, since this might go to an external caches
	Store(item T) error
	StoreSafeBackground(item T)
	Remove(id string) error
	Clear()
}

func GetCache[T CachedItem]() ItemsCache[T] {
	return &localItemsCache[T]{}
}
