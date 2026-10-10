package runtime

// Pinner keeps objects alive until Unpin is called. The GC does not move
// objects. The pins are kept in a global map so that a Pinner embedded in
// the object it pins does not become unreachable.
type Pinner struct {
	_ byte
}

var pinned map[*Pinner][]any

func (p *Pinner) Pin(pointer any) {
	if pinned == nil {
		pinned = make(map[*Pinner][]any)
	}
	pinned[p] = append(pinned[p], pointer)
}

func (p *Pinner) Unpin() {
	delete(pinned, p)
}
