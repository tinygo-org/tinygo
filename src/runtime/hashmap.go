package runtime

// This is a hashmap implementation for the map[T]T type.
// It is very roughly based on the implementation of the Go hashmap:
//
//     https://golang.org/src/runtime/map.go

import (
	"reflect"
	"unsafe"
)

// The underlying hashmap structure for Go.
type hashmap struct {
	buckets    unsafe.Pointer // pointer to array of buckets
	count      uintptr
	keySize    uint8 // maybe this can store the key type as well? E.g. keysize == 5 means string?
	valueSize  uint8
	bucketBits uint8
	keyEqual   func(x, y unsafe.Pointer, n uintptr) bool
	keyHash    func(key unsafe.Pointer, size uintptr) uint32
}

type hashmapAlgorithm uint8

const (
	hashmapAlgorithmBinary hashmapAlgorithm = iota
	hashmapAlgorithmString
	hashmapAlgorithmInterface
)

// A hashmap bucket. A bucket is a container of 8 key/value pairs: first the
// following two entries, then the 8 keys, then the 8 values. This somewhat odd
// ordering is to make sure the keys and values are well aligned when one of
// them is smaller than the system word size.
type hashmapBucket struct {
	tophash [8]uint8
	next    *hashmapBucket // next bucket (if there are more than 8 in a chain)
	// Followed by the actual keys, and then the actual values. These are
	// allocated but as they're of variable size they can't be shown here.
}

type hashmapIterator struct {
	buckets      unsafe.Pointer
	numBuckets   uintptr
	bucketNumber uintptr
	bucket       *hashmapBucket
	bucketIndex  uint8
}

// Get the topmost 8 bits of the hash, without using a special value (like 0).
func hashmapTopHash(hash uint32) uint8 {
	tophash := uint8(hash >> 24)
	if tophash < 1 {
		// 0 means empty slot, so make it bigger.
		tophash += 1
	}
	return tophash
}

// Create a new hashmap with the given keySize and valueSize.
func hashmapMake(keySize, valueSize uint8, sizeHint uintptr, alg uint8) *hashmap {
	numBuckets := sizeHint / 8
	bucketBits := uint8(0)
	for numBuckets != 0 {
		numBuckets /= 2
		bucketBits++
	}
	bucketBufSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(keySize)*8 + uintptr(valueSize)*8
	buckets := alloc(bucketBufSize*(1<<bucketBits), nil)

	keyHash := hashmapKeyHashAlg(hashmapAlgorithm(alg))
	keyEqual := hashmapKeyEqualAlg(hashmapAlgorithm(alg))

	return &hashmap{
		buckets:    buckets,
		keySize:    keySize,
		valueSize:  valueSize,
		bucketBits: bucketBits,
		keyEqual:   keyEqual,
		keyHash:    keyHash,
	}
}

func hashmapKeyEqualAlg(alg hashmapAlgorithm) func(x, y unsafe.Pointer, n uintptr) bool {
	switch alg {
	case hashmapAlgorithmBinary:
		return memequal
	case hashmapAlgorithmString:
		return hashmapStringEqual
	case hashmapAlgorithmInterface:
		return hashmapInterfaceEqual
	default:
		panic("unknown hashmap equal algorithm")
	}
}

func hashmapKeyHashAlg(alg hashmapAlgorithm) func(key unsafe.Pointer, n uintptr) uint32 {
	switch alg {
	case hashmapAlgorithmBinary:
		return hash32
	case hashmapAlgorithmString:
		return hashmapStringPtrHash
	case hashmapAlgorithmInterface:
		return hashmapInterfacePtrHash
	default:
		panic("unknown hashmap hash algorithm")
	}
}

func hashmapShouldGrow(m *hashmap) bool {
	if m.bucketBits == 32 {
		// can't grow any more
		return false
	}

	// "maximum" number of elements is 0.75 * buckets * elements per bucket
	max := uintptr((3 * ((1 << m.bucketBits) * 8)) / 4)

	return m.count > max
}

// Return the number of entries in this hashmap, called from the len builtin.
// A nil hashmap is defined as having length 0.
//go:inline
func hashmapLen(m *hashmap) int {
	if m == nil {
		return 0
	}
	return int(m.count)
}

// wrapper for use in reflect
func hashmapLenUnsafePointer(p unsafe.Pointer) int {
	m := (*hashmap)(p)
	return hashmapLen(m)
}

// Set a specified key to a given value. Grow the map if necessary.
//go:nobounds
func hashmapSet(m *hashmap, key unsafe.Pointer, value unsafe.Pointer, hash uint32) {
	tophash := hashmapTopHash(hash)

	if m.buckets == nil {
		// No bucket was allocated yet, do so now.
		m.buckets = unsafe.Pointer(hashmapInsertIntoNewBucket(m, key, value, tophash))
		return
	}

	if hashmapShouldGrow(m) {
		hashmapGrow(m)
	}

	numBuckets := uintptr(1) << m.bucketBits
	bucketNumber := (uintptr(hash) & (numBuckets - 1))
	bucketSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*8
	bucketAddr := uintptr(m.buckets) + bucketSize*bucketNumber
	bucket := (*hashmapBucket)(unsafe.Pointer(bucketAddr))
	var lastBucket *hashmapBucket

	// See whether the key already exists somewhere.
	var emptySlotKey unsafe.Pointer
	var emptySlotValue unsafe.Pointer
	var emptySlotTophash *byte
	for bucket != nil {
		for i := uintptr(0); i < 8; i++ {
			slotKeyOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*uintptr(i)
			slotKey := unsafe.Pointer(uintptr(unsafe.Pointer(bucket)) + slotKeyOffset)
			slotValueOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*uintptr(i)
			slotValue := unsafe.Pointer(uintptr(unsafe.Pointer(bucket)) + slotValueOffset)
			if bucket.tophash[i] == 0 && emptySlotKey == nil {
				// Found an empty slot, store it for if we couldn't find an
				// existing slot.
				emptySlotKey = slotKey
				emptySlotValue = slotValue
				emptySlotTophash = &bucket.tophash[i]
			}
			if bucket.tophash[i] == tophash {
				// Could be an existing key that's the same.
				if m.keyEqual(key, slotKey, uintptr(m.keySize)) {
					// found same key, replace it
					memcpy(slotValue, value, uintptr(m.valueSize))
					return
				}
			}
		}
		lastBucket = bucket
		bucket = bucket.next
	}
	if emptySlotKey == nil {
		// Add a new bucket to the bucket chain.
		// TODO: rebalance if necessary to avoid O(n) insert and lookup time.
		lastBucket.next = (*hashmapBucket)(hashmapInsertIntoNewBucket(m, key, value, tophash))
		return
	}
	m.count++
	memcpy(emptySlotKey, key, uintptr(m.keySize))
	memcpy(emptySlotValue, value, uintptr(m.valueSize))
	*emptySlotTophash = tophash
}

// hashmapInsertIntoNewBucket creates a new bucket, inserts the given key and
// value into the bucket, and returns a pointer to this bucket.
func hashmapInsertIntoNewBucket(m *hashmap, key, value unsafe.Pointer, tophash uint8) *hashmapBucket {
	bucketBufSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*8
	bucketBuf := alloc(bucketBufSize, nil)
	// Insert into the first slot, which is empty as it has just been allocated.
	slotKeyOffset := unsafe.Sizeof(hashmapBucket{})
	slotKey := unsafe.Pointer(uintptr(bucketBuf) + slotKeyOffset)
	slotValueOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8
	slotValue := unsafe.Pointer(uintptr(bucketBuf) + slotValueOffset)
	m.count++
	memcpy(slotKey, key, uintptr(m.keySize))
	memcpy(slotValue, value, uintptr(m.valueSize))
	bucket := (*hashmapBucket)(bucketBuf)
	bucket.tophash[0] = tophash
	return bucket
}

func hashmapGrow(m *hashmap) {

	// clone map as empty
	n := *m
	n.count = 0

	// allocate our new buckets twice as big
	n.bucketBits = m.bucketBits + 1
	numBuckets := uintptr(1) << n.bucketBits
	bucketBufSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*8
	n.buckets = alloc(bucketBufSize*numBuckets, nil)

	// use a hashmap iterator to go through the old map
	var it hashmapIterator

	var key = alloc(uintptr(m.keySize), nil)
	var value = alloc(uintptr(m.valueSize), nil)

	for hashmapNext(m, &it, key, value) {
		h := m.keyHash(key, uintptr(m.keySize))
		hashmapSet(&n, key, value, h)
	}

	*m = n
}

// Get the value of a specified key, or zero the value if not found.
//go:nobounds
func hashmapGet(m *hashmap, key, value unsafe.Pointer, valueSize uintptr, hash uint32) bool {
	if m == nil {
		// Getting a value out of a nil map is valid. From the spec:
		// > if the map is nil or does not contain such an entry, a[x] is the
		// > zero value for the element type of M
		memzero(value, uintptr(valueSize))
		return false
	}
	numBuckets := uintptr(1) << m.bucketBits
	bucketNumber := (uintptr(hash) & (numBuckets - 1))
	bucketSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*8
	bucketAddr := uintptr(m.buckets) + bucketSize*bucketNumber
	bucket := (*hashmapBucket)(unsafe.Pointer(bucketAddr))

	tophash := uint8(hash >> 24)
	if tophash < 1 {
		// 0 means empty slot, so make it bigger.
		tophash += 1
	}

	// Try to find the key.
	for bucket != nil {
		for i := uintptr(0); i < 8; i++ {
			slotKeyOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*uintptr(i)
			slotKey := unsafe.Pointer(uintptr(unsafe.Pointer(bucket)) + slotKeyOffset)
			slotValueOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*uintptr(i)
			slotValue := unsafe.Pointer(uintptr(unsafe.Pointer(bucket)) + slotValueOffset)
			if bucket.tophash[i] == tophash {
				// This could be the key we're looking for.
				if m.keyEqual(key, slotKey, uintptr(m.keySize)) {
					// Found the key, copy it.
					memcpy(value, slotValue, uintptr(m.valueSize))
					return true
				}
			}
		}
		bucket = bucket.next
	}

	// Did not find the key.
	memzero(value, uintptr(m.valueSize))
	return false
}

// Delete a given key from the map. No-op when the key does not exist in the
// map.
//go:nobounds
func hashmapDelete(m *hashmap, key unsafe.Pointer, hash uint32) {
	if m == nil {
		// The delete builtin is defined even when the map is nil. From the spec:
		// > If the map m is nil or the element m[k] does not exist, delete is a
		// > no-op.
		return
	}
	numBuckets := uintptr(1) << m.bucketBits
	bucketNumber := (uintptr(hash) & (numBuckets - 1))
	bucketSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*8
	bucketAddr := uintptr(m.buckets) + bucketSize*bucketNumber
	bucket := (*hashmapBucket)(unsafe.Pointer(bucketAddr))

	tophash := uint8(hash >> 24)
	if tophash < 1 {
		// 0 means empty slot, so make it bigger.
		tophash += 1
	}

	// Try to find the key.
	for bucket != nil {
		for i := uintptr(0); i < 8; i++ {
			slotKeyOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*uintptr(i)
			slotKey := unsafe.Pointer(uintptr(unsafe.Pointer(bucket)) + slotKeyOffset)
			if bucket.tophash[i] == tophash {
				// This could be the key we're looking for.
				if m.keyEqual(key, slotKey, uintptr(m.keySize)) {
					// Found the key, delete it.
					bucket.tophash[i] = 0
					m.count--
					return
				}
			}
		}
		bucket = bucket.next
	}
}

// Iterate over a hashmap.
//go:nobounds
func hashmapNext(m *hashmap, it *hashmapIterator, key, value unsafe.Pointer) bool {
	if m == nil {
		// From the spec: If the map is nil, the number of iterations is 0.
		return false
	}

	if it.buckets == nil {
		// initialize iterator
		it.buckets = m.buckets
		it.numBuckets = uintptr(1) << m.bucketBits
	}

	for {
		if it.bucketIndex >= 8 {
			// end of bucket, move to the next in the chain
			it.bucketIndex = 0
			it.bucket = it.bucket.next
		}
		if it.bucket == nil {
			if it.bucketNumber >= it.numBuckets {
				// went through all buckets
				return false
			}
			bucketSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*8
			bucketAddr := uintptr(it.buckets) + bucketSize*it.bucketNumber
			it.bucket = (*hashmapBucket)(unsafe.Pointer(bucketAddr))
			it.bucketNumber++ // next bucket
		}
		if it.bucket.tophash[it.bucketIndex] == 0 {
			// slot is empty - move on
			it.bucketIndex++
			continue
		}

		bucketAddr := uintptr(unsafe.Pointer(it.bucket))
		slotKeyOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*uintptr(it.bucketIndex)
		slotKey := unsafe.Pointer(bucketAddr + slotKeyOffset)
		memcpy(key, slotKey, uintptr(m.keySize))

		if it.buckets == m.buckets {
			// Our view of the buckets is the same as the parent map.
			// Just copy the value we have
			slotValueOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*uintptr(it.bucketIndex)
			slotValue := unsafe.Pointer(bucketAddr + slotValueOffset)
			memcpy(value, slotValue, uintptr(m.valueSize))
			it.bucketIndex++
		} else {
			it.bucketIndex++

			// Our view of the buckets doesn't match the parent map.
			// Look up the key in the new buckets and return that value if it exists
			hash := m.keyHash(key, uintptr(m.keySize))
			ok := hashmapGet(m, key, value, uintptr(m.valueSize), hash)
			if !ok {
				// doesn't exist in parent map; try next key
				continue
			}

			// All good.
		}

		return true
	}
}

// Hashmap with plain binary data keys (not containing strings etc.).

func hashmapBinarySet(m *hashmap, key, value unsafe.Pointer) {
	// TODO: detect nil map here and throw a better panic message?
	hash := hash32(key, uintptr(m.keySize))
	hashmapSet(m, key, value, hash)
}

func hashmapBinaryGet(m *hashmap, key, value unsafe.Pointer, valueSize uintptr) bool {
	if m == nil {
		memzero(value, uintptr(valueSize))
		return false
	}
	hash := hash32(key, uintptr(m.keySize))
	return hashmapGet(m, key, value, valueSize, hash)
}

func hashmapBinaryDelete(m *hashmap, key unsafe.Pointer) {
	if m == nil {
		return
	}
	hash := hash32(key, uintptr(m.keySize))
	hashmapDelete(m, key, hash)
}

// Hashmap with string keys (a common case).

func hashmapStringEqual(x, y unsafe.Pointer, n uintptr) bool {
	return *(*string)(x) == *(*string)(y)
}

func hashmapStringHash(s string) uint32 {
	_s := (*_string)(unsafe.Pointer(&s))
	return hash32(unsafe.Pointer(_s.ptr), uintptr(_s.length))
}

func hashmapStringPtrHash(sptr unsafe.Pointer, size uintptr) uint32 {
	_s := *(*_string)(sptr)
	return hash32(unsafe.Pointer(_s.ptr), uintptr(_s.length))
}

func hashmapStringSet(m *hashmap, key string, value unsafe.Pointer) {
	hash := hashmapStringHash(key)
	hashmapSet(m, unsafe.Pointer(&key), value, hash)
}

func hashmapStringGet(m *hashmap, key string, value unsafe.Pointer, valueSize uintptr) bool {
	hash := hashmapStringHash(key)
	return hashmapGet(m, unsafe.Pointer(&key), value, valueSize, hash)
}

func hashmapStringDelete(m *hashmap, key string) {
	hash := hashmapStringHash(key)
	hashmapDelete(m, unsafe.Pointer(&key), hash)
}

// Hashmap with interface keys (for everything else).

// This is a method that is intentionally unexported in the reflect package. It
// is identical to the Interface() method call, except it doesn't check whether
// a field is exported and thus allows circumventing the type system.
// The hash function needs it as it also needs to hash unexported struct fields.
//go:linkname valueInterfaceUnsafe reflect.valueInterfaceUnsafe
func valueInterfaceUnsafe(v reflect.Value) interface{}

func hashmapFloat32Hash(ptr unsafe.Pointer) uint32 {
	f := *(*uint32)(ptr)
	if f == 0x80000000 {
		// convert -0 to 0 for hashing
		f = 0
	}
	return hash32(unsafe.Pointer(&f), 4)
}

func hashmapFloat64Hash(ptr unsafe.Pointer) uint32 {
	f := *(*uint64)(ptr)
	if f == 0x8000000000000000 {
		// convert -0 to 0 for hashing
		f = 0
	}
	return hash32(unsafe.Pointer(&f), 8)
}

func hashmapInterfaceHash(itf interface{}) uint32 {
	x := reflect.ValueOf(itf)
	if x.RawType() == 0 {
		return 0 // nil interface
	}

	value := (*_interface)(unsafe.Pointer(&itf)).value
	ptr := value
	if x.RawType().Size() <= unsafe.Sizeof(uintptr(0)) {
		// Value fits in pointer, so it's directly stored in the pointer.
		ptr = unsafe.Pointer(&value)
	}

	switch x.RawType().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return hash32(ptr, x.RawType().Size())
	case reflect.Bool, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return hash32(ptr, x.RawType().Size())
	case reflect.Float32:
		// It should be possible to just has the contents. However, NaN != NaN
		// so if you're using lots of NaNs as map keys (you shouldn't) then hash
		// time may become exponential. To fix that, it would be better to
		// return a random number instead:
		// https://research.swtch.com/randhash
		return hashmapFloat32Hash(ptr)
	case reflect.Float64:
		return hashmapFloat64Hash(ptr)
	case reflect.Complex64:
		rptr, iptr := ptr, unsafe.Pointer(uintptr(ptr)+4)
		return hashmapFloat32Hash(rptr) ^ hashmapFloat32Hash(iptr)
	case reflect.Complex128:
		rptr, iptr := ptr, unsafe.Pointer(uintptr(ptr)+8)
		return hashmapFloat64Hash(rptr) ^ hashmapFloat64Hash(iptr)
	case reflect.String:
		return hashmapStringHash(x.String())
	case reflect.Chan, reflect.Ptr, reflect.UnsafePointer:
		// It might seem better to just return the pointer, but that won't
		// result in an evenly distributed hashmap. Instead, hash the pointer
		// like most other types.
		return hash32(ptr, x.RawType().Size())
	case reflect.Array:
		var hash uint32
		for i := 0; i < x.Len(); i++ {
			hash ^= hashmapInterfaceHash(valueInterfaceUnsafe(x.Index(i)))
		}
		return hash
	case reflect.Struct:
		var hash uint32
		for i := 0; i < x.NumField(); i++ {
			hash ^= hashmapInterfaceHash(valueInterfaceUnsafe(x.Field(i)))
		}
		return hash
	default:
		runtimePanic("comparing un-comparable type")
		return 0 // unreachable
	}
}

func hashmapInterfacePtrHash(iptr unsafe.Pointer, size uintptr) uint32 {
	_i := *(*_interface)(iptr)
	return hashmapInterfaceHash(_i)
}

func hashmapInterfaceEqual(x, y unsafe.Pointer, n uintptr) bool {
	return *(*interface{})(x) == *(*interface{})(y)
}

func hashmapInterfaceSet(m *hashmap, key interface{}, value unsafe.Pointer) {
	hash := hashmapInterfaceHash(key)
	hashmapSet(m, unsafe.Pointer(&key), value, hash)
}

func hashmapInterfaceGet(m *hashmap, key interface{}, value unsafe.Pointer, valueSize uintptr) bool {
	hash := hashmapInterfaceHash(key)
	return hashmapGet(m, unsafe.Pointer(&key), value, valueSize, hash)
}

func hashmapInterfaceDelete(m *hashmap, key interface{}) {
	hash := hashmapInterfaceHash(key)
	hashmapDelete(m, unsafe.Pointer(&key), hash)
}
