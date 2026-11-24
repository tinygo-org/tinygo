//go:build gc.conservative || gc.precise

package runtime

// Memory profiling for the block-based collectors. Sampled objects are found
// by their head block, so sweep can count frees. gcLock guards all of it.

import "unsafe"

const memProfileMaxDepth = 32

// memBucket holds the counts for one allocation stack and size. Keying on the
// size too, like upstream, lets runtime/pprof scale each bucket correctly.
type memBucket struct {
	hash  uintptr
	size  uintptr
	depth uintptr
	stack [memProfileMaxDepth]uintptr

	// Counts as of the last completed GC, which MemProfile reports.
	allocs, frees, allocBytes, freeBytes uint64
	// Counts since the last completed GC.
	pendAllocs, pendFrees, pendAllocBytes, pendFreeBytes uint64
}

// memObject is a sampled object that has not been freed yet.
type memObject struct {
	// key is the address of the head block, bitwise-NOT so the conservative
	// collector does not see it as a pointer. Zero is an empty slot.
	key    uintptr
	size   uintptr
	bucket uintptr
}

// memProfileTableLayout has no pointers like gclayout.NoPtrs, but alloc does
// not sample it, so the profiler can allocate its tables.
var memProfileTableLayout = unsafe.Pointer(uintptr(3<<1 | 1))

var (
	memProfileNext    uintptr     // bytes left until the next sample, 0 before the first
	memProfileBuckets []memBucket // all buckets, len is the number in use
	memProfileIndex   []uint32    // open addressing table of bucket index+1, by stack hash
	memProfileObjects []memObject // open addressing table of sampled objects
	memProfileLive    uintptr     // number of entries in memProfileObjects
	memProfileBits    []byte      // one bit per head block of a sampled object
)

// memProfileCount accounts for an allocation of size bytes and reports whether
// it should be sampled. Call it with gcLock held.
func memProfileCount(size uintptr) bool {
	rate := MemProfileRate
	if rate <= 0 {
		return false
	}
	if rate == 1 {
		return true
	}
	if memProfileNext == 0 {
		memProfileNext = memProfileNextSample(rate)
	}
	if size < memProfileNext {
		memProfileNext -= size
		return false
	}
	memProfileNext = memProfileNextSample(rate)
	return true
}

// memProfileNextSample returns an exponentially distributed distance with the
// given mean, which the scaling in runtime/pprof assumes.
func memProfileNextSample(rate int) uintptr {
	if rate > 0x7000000 {
		rate = 0x7000000
	}
	u := float64(fastrand()>>8+1) / (1 << 24) // in (0, 1]
	return uintptr(-memProfileLog(u)*float64(rate)) + 1
}

// memProfileLog returns the natural logarithm of x > 0, accurate to about
// 1e-6, which is plenty for picking sample distances.
func memProfileLog(x float64) float64 {
	const ln2 = 0.6931471805599453
	bits := float64bits(x)
	exp := int(bits>>52&0x7ff) - 1023
	m := float64frombits(bits&^(0x7ff<<52) | 1023<<52) // in [1, 2)
	// ln(m) = 2*atanh(z), with z in [0, 1/3).
	z := (m - 1) / (m + 1)
	z2 := z * z
	return float64(exp)*ln2 + 2*z*(1+z2*(1.0/3+z2*(1.0/5+z2*(1.0/7+z2*(1.0/9)))))
}

// memProfileAlloc records a sampled allocation. alloc calls it directly, as
// memProfileCallers skips a fixed number of frames.
//
//go:noinline
func memProfileAlloc(ptr unsafe.Pointer, size uintptr) {
	var stk [memProfileMaxDepth]uintptr
	depth := uintptr(memProfileCallers(stk[:]))
	hash := (2166136261 ^ size) * 16777619
	for _, pc := range stk[:depth] {
		hash = (hash ^ pc) * 16777619
	}

	gcLock.Lock()
	head := blockFromAddr(uintptr(ptr)).findHead()
	for memProfileGrow(head) {
	}
	b := memProfileBucket(hash, size, stk[:depth])
	b.pendAllocs++
	b.pendAllocBytes += uint64(size)
	memProfileInsertObject(memObject{key: ^head.address(), size: size, bucket: uintptr(b.index())})
	memProfileBits[head/8] |= 1 << (head % 8)
	gcLock.Unlock()
}

func (b *memBucket) index() uintptr {
	return (uintptr(unsafe.Pointer(b)) - uintptr(unsafe.Pointer(&memProfileBuckets[0]))) / unsafe.Sizeof(memBucket{})
}

// memProfileGrow grows one table that lacks room for this sample, unlocking
// gcLock to allocate. It returns false once nothing needs to grow.
func memProfileGrow(head gcBlock) bool {
	var size uintptr
	var kind int
	switch {
	case len(memProfileBuckets) == cap(memProfileBuckets):
		kind, size = 0, max(64, 2*uintptr(cap(memProfileBuckets)))*unsafe.Sizeof(memBucket{})
	case 2*(len(memProfileBuckets)+1) > len(memProfileIndex):
		kind, size = 1, max(128, 2*uintptr(len(memProfileIndex)))*unsafe.Sizeof(uint32(0))
	case 2*(memProfileLive+1) > uintptr(len(memProfileObjects)):
		kind, size = 2, max(128, 2*uintptr(len(memProfileObjects)))*unsafe.Sizeof(memObject{})
	case uintptr(head)/8 >= uintptr(len(memProfileBits)):
		// Leave some room, as the heap usually keeps growing.
		kind, size = 3, (uintptr(endBlock)/8+1)*5/4
	default:
		return false
	}

	// alloc does not sample allocations with this layout, so growing a table
	// does not recurse.
	gcLock.Unlock()
	mem := alloc(size, memProfileTableLayout)
	gcLock.Lock()

	// Another thread may have grown the table in the meantime.
	switch kind {
	case 0:
		n := size / unsafe.Sizeof(memBucket{})
		if n > uintptr(cap(memProfileBuckets)) {
			buckets := unsafe.Slice((*memBucket)(mem), n)[:len(memProfileBuckets)]
			copy(buckets, memProfileBuckets)
			memProfileBuckets = buckets
		}
	case 1:
		n := size / unsafe.Sizeof(uint32(0))
		if n > uintptr(len(memProfileIndex)) {
			memProfileIndex = unsafe.Slice((*uint32)(mem), n)
			for i := range memProfileBuckets {
				memProfileIndexInsert(uint32(i))
			}
		}
	case 2:
		n := size / unsafe.Sizeof(memObject{})
		if n > uintptr(len(memProfileObjects)) {
			old := memProfileObjects
			memProfileObjects = unsafe.Slice((*memObject)(mem), n)
			memProfileLive = 0
			for _, obj := range old {
				if obj.key != 0 {
					memProfileInsertObject(obj)
				}
			}
		}
	case 3:
		if size > uintptr(len(memProfileBits)) {
			bits := unsafe.Slice((*byte)(mem), size)
			copy(bits, memProfileBits)
			memProfileBits = bits
		}
	}
	return true
}

// memProfileBucket returns the bucket for a stack and size, adding it if needed. There
// must be room for one more bucket.
func memProfileBucket(hash, size uintptr, stk []uintptr) *memBucket {
	mask := uintptr(len(memProfileIndex) - 1)
	for i := hash & mask; ; i = (i + 1) & mask {
		idx := memProfileIndex[i]
		if idx == 0 {
			break
		}
		b := &memProfileBuckets[idx-1]
		if b.hash == hash && b.size == size && b.depth == uintptr(len(stk)) && memProfileSameStack(b.stack[:b.depth], stk) {
			return b
		}
	}
	n := len(memProfileBuckets)
	memProfileBuckets = memProfileBuckets[:n+1]
	b := &memProfileBuckets[n]
	b.hash = hash
	b.size = size
	b.depth = uintptr(len(stk))
	copy(b.stack[:], stk)
	memProfileIndexInsert(uint32(n))
	return b
}

func memProfileSameStack(a, b []uintptr) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func memProfileIndexInsert(bucket uint32) {
	mask := uintptr(len(memProfileIndex) - 1)
	i := memProfileBuckets[bucket].hash & mask
	for memProfileIndex[i] != 0 {
		i = (i + 1) & mask
	}
	memProfileIndex[i] = bucket + 1
}

func memProfileObjectSlot(key uintptr) uintptr {
	// The low bits of a block address are always the same, so mix them out.
	return (key * 0x9e3779b9 >> 8) & uintptr(len(memProfileObjects)-1)
}

func memProfileInsertObject(obj memObject) {
	mask := uintptr(len(memProfileObjects) - 1)
	i := memProfileObjectSlot(obj.key)
	for memProfileObjects[i].key != 0 {
		i = (i + 1) & mask
	}
	memProfileObjects[i] = obj
	memProfileLive++
}

// memProfileSampled reports whether head is the head block of a sampled
// object. Call it with gcLock held.
func memProfileSampled(head gcBlock) bool {
	return memProfileEnabled && memProfileLive != 0 &&
		uintptr(head)/8 < uintptr(len(memProfileBits)) &&
		memProfileBits[head/8]&(1<<(head%8)) != 0
}

// memProfileFree records the free of the sampled object with this head block.
// It must not allocate, as sweep calls it.
func memProfileFree(head gcBlock) {
	memProfileBits[head/8] &^= 1 << (head % 8)
	key := ^head.address()
	mask := uintptr(len(memProfileObjects) - 1)
	i := memProfileObjectSlot(key)
	for memProfileObjects[i].key != key {
		if memProfileObjects[i].key == 0 {
			runtimeFatal("memprofile: sampled object not found")
		}
		i = (i + 1) & mask
	}
	obj := memProfileObjects[i]
	b := &memProfileBuckets[obj.bucket]
	b.pendFrees++
	b.pendFreeBytes += uint64(obj.size)

	// Backward shift deletion keeps the probe sequences intact.
	for {
		memProfileObjects[i] = memObject{}
		j := i
		for {
			j = (j + 1) & mask
			next := memProfileObjects[j]
			if next.key == 0 {
				memProfileLive--
				return
			}
			home := memProfileObjectSlot(next.key)
			// Move next into the hole unless its home slot lies in (i, j].
			if (j > i && (home <= i || home > j)) || (j < i && home <= i && home > j) {
				memProfileObjects[i] = next
				i = j
				break
			}
		}
	}
}

// memProfileFlush publishes the counts since the last GC. runGC calls it once
// sweep is done.
func memProfileFlush() {
	for i := range memProfileBuckets {
		b := &memProfileBuckets[i]
		b.allocs += b.pendAllocs
		b.frees += b.pendFrees
		b.allocBytes += b.pendAllocBytes
		b.freeBytes += b.pendFreeBytes
		b.pendAllocs, b.pendFrees, b.pendAllocBytes, b.pendFreeBytes = 0, 0, 0, 0
	}
}

func memProfile(p []MemProfileRecord, inuseZero bool) (n int, ok bool) {
	gcLock.Lock()
	// Before the first GC nothing is published. Publish what there is, so
	// short programs get a profile too.
	empty := true
	for i := range memProfileBuckets {
		b := &memProfileBuckets[i]
		if b.allocs != 0 || b.frees != 0 {
			empty = false
			break
		}
	}
	if empty {
		memProfileFlush()
	}
	for i := range memProfileBuckets {
		b := &memProfileBuckets[i]
		if inuseZero || b.allocBytes != b.freeBytes {
			n++
		}
	}
	if n <= len(p) {
		ok = true
		j := 0
		for i := range memProfileBuckets {
			b := &memProfileBuckets[i]
			if !inuseZero && b.allocBytes == b.freeBytes {
				continue
			}
			r := &p[j]
			j++
			r.AllocBytes = int64(b.allocBytes)
			r.FreeBytes = int64(b.freeBytes)
			r.AllocObjects = int64(b.allocs)
			r.FreeObjects = int64(b.frees)
			copy(r.Stack0[:], b.stack[:b.depth])
			clear(r.Stack0[b.depth:])
		}
	}
	gcLock.Unlock()
	return
}
