package sync

type Once struct {
	done bool
	m    Mutex
}

func (o *Once) Do(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done {
		return
	}
	o.done = true
	f()
}
