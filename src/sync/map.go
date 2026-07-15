package sync

import (
	"internal/task"
)

// This file implements just enough of sync.Map to get packages to compile. It
// is no more efficient than a map with a lock.

type Map struct {
	lock task.PMutex
	m    map[interface{}]interface{}
}

func (m *Map) Delete(key interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.m, key)
}

func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	value, ok = m.m[key]
	return
}

func (m *Map) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}
	if existing, ok := m.m[key]; ok {
		return existing, true
	}
	m.m[key] = value
	return value, false
}

func (m *Map) LoadAndDelete(key interface{}) (value interface{}, loaded bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	value, ok := m.m[key]
	if !ok {
		return nil, false
	}
	delete(m.m, key)
	return value, true
}

func (m *Map) Store(key, value interface{}) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}
	m.m[key] = value
}

// Range calls f for each key and value in the map.  If f returns false, the iteration stops.
func (m *Map) Range(f func(key, value interface{}) bool) {
	// Iterate over a key snapshot instead of holding the lock across the callback,
	// to prevent deadlock when a Map method is called inside f.
	//
	// Using a key snapshot in Map.Range is sufficient because Go specifies that:
	// - Range only requires that no key is visited more than once, and
	// - Range may reflect any mapping from any point during the Range call.

	m.lock.Lock()
	keys := make([]interface{}, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}
	m.lock.Unlock()

	for _, k := range keys {
		if v, ok := m.Load(k); ok {
			if !f(k, v) {
				break
			}
		}
	}
}

func (m *Map) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	clear(m.m)
}

// Swap replaces the value for the given key, and returns the old value if any.
func (m *Map) Swap(key, value any) (previous any, loaded bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.m == nil {
		m.m = make(map[interface{}]interface{})
	}
	previous, loaded = m.m[key]
	m.m[key] = value
	return
}

// CompareAndSwap swaps the old and new values for an existing key if the value
// stored in the map is equal to old.
func (m *Map) CompareAndSwap(key, old, new any) (swapped bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.m == nil {
		return false
	}
	value, ok := m.m[key]
	if !ok || value != old {
		return false
	}
	m.m[key] = new
	return true
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
func (m *Map) CompareAndDelete(key, old any) (deleted bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.m == nil {
		return false
	}
	value, ok := m.m[key]
	if !ok || value != old {
		return false
	}
	delete(m.m, key)
	return true
}
