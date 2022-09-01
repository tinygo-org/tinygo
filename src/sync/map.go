package sync

// This file implements just enough of sync.Map to get packages to compile. It
// is no more efficient than a map with a lock.

type Map struct {
	lock Mutex
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

func (m *Map) Range(f func(key, value interface{}) bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.m == nil {
		return
	}

	for k, v := range m.m {
		if !f(k, v) {
			break
		}
	}
}
