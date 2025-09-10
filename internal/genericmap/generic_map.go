package genericmap

import (
	"sync"
)

// GenericMap provides thread-safe storage for Value's type objects
// indexed by Key.
type GenericMap[TKey comparable, TVal any] interface {
	// Store saves a value with the given key
	Store(key TKey, val TVal)
	// Load retrieves a value by key
	Load(key TKey) (val TVal, exists bool)
	// Delete removes a value by key
	Delete(key TKey)
	// LoadAndDelete retrieves and removes a value by key in one operation
	LoadAndDelete(key TKey) (val TVal, exists bool)
	// Clear removes all stored Value
	Clear()
	// Size returns the number of stored Value
	Size() int
}

type genericMap[TKey comparable, TVal any] struct {
	mu sync.RWMutex
	m  map[TKey]TVal
}

func NewGenericMap[TKey comparable, TVal any]() GenericMap[TKey, TVal] {
	return &genericMap[TKey, TVal]{
		m: make(map[TKey]TVal),
	}
}

func (cm *genericMap[TKey, TVal]) Store(key TKey, val TVal) {
	cm.mu.Lock()
	cm.m[key] = val
	cm.mu.Unlock()
}

func (cm *genericMap[TKey, TVal]) Load(key TKey) (val TVal, exists bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	val, exists = cm.m[key]
	return
}

func (cm *genericMap[TKey, TVal]) Delete(key TKey) {
	cm.mu.Lock()
	delete(cm.m, key)
	cm.mu.Unlock()
}

func (cm *genericMap[TKey, TVal]) LoadAndDelete(key TKey) (val TVal, exists bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	val, exists = cm.m[key]
	if exists {
		delete(cm.m, key)
	}
	return
}

func (cm *genericMap[TKey, TVal]) Clear() {
	cm.mu.Lock()
	cm.m = make(map[TKey]TVal)
	cm.mu.Unlock()
}

func (cm *genericMap[TKey, TVal]) Size() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.m)
}
