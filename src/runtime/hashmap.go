package runtime

// This is a hashmap implementation for the map[T]T type.
// It is very rougly based on the implementation of the Go hashmap:
//
//     https://golang.org/src/runtime/map.go

import (
	"unsafe"
)

// The underlying hashmap structure for Go.
type hashmap struct {
	next       *hashmap       // hashmap after evacuate (for iterators)
	buckets    unsafe.Pointer // pointer to array of buckets
	count      uintptr
	keySize    uint8 // maybe this can store the key type as well? E.g. keysize == 5 means string?
	valueSize  uint8
	bucketBits uint8
}

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
	bucketNumber uintptr
	bucket       *hashmapBucket
	bucketIndex  uint8
}

// Get FNV-1a hash of this key.
//
// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function#FNV-1a_hash
func hashmapHash(ptr unsafe.Pointer, n uintptr) uint32 {
	var result uint32 = 2166136261 // FNV offset basis
	for i := uintptr(0); i < n; i++ {
		c := *(*uint8)(unsafe.Pointer(uintptr(ptr) + i))
		result ^= uint32(c) // XOR with byte
		result *= 16777619  // FNV prime
	}
	return result
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
func hashmapMake(keySize, valueSize uint8) *hashmap {
	bucketBufSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(keySize)*8 + uintptr(valueSize)*8
	bucket := alloc(bucketBufSize)
	return &hashmap{
		buckets:    bucket,
		keySize:    keySize,
		valueSize:  valueSize,
		bucketBits: 0,
	}
}

// Return the number of entries in this hashmap, called from the len builtin.
// A nil hashmap is defined as having length 0.
func hashmapLen(m *hashmap) int {
	if m == nil {
		return 0
	}
	return int(m.count)
}

// Set a specified key to a given value. Grow the map if necessary.
//go:nobounds
func hashmapSet(m *hashmap, key unsafe.Pointer, value unsafe.Pointer, hash uint32, keyEqual func(x, y unsafe.Pointer, n uintptr) bool) {
	numBuckets := uintptr(1) << m.bucketBits
	bucketNumber := (uintptr(hash) & (numBuckets - 1))
	bucketSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*8
	bucketAddr := uintptr(m.buckets) + bucketSize*bucketNumber
	bucket := (*hashmapBucket)(unsafe.Pointer(bucketAddr))

	tophash := hashmapTopHash(hash)

	// See whether the key already exists somewhere.
	var emptySlotKey unsafe.Pointer
	var emptySlotValue unsafe.Pointer
	var emptySlotTophash *byte
	for bucket != nil {
		for i := uintptr(0); i < 8; i++ {
			slotKeyOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*uintptr(i)
			slotKey := unsafe.Pointer(bucketAddr + slotKeyOffset)
			slotValueOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*uintptr(i)
			slotValue := unsafe.Pointer(bucketAddr + slotValueOffset)
			if bucket.tophash[i] == 0 && emptySlotKey == nil {
				// Found an empty slot, store it for if we couldn't find an
				// existing slot.
				emptySlotKey = slotKey
				emptySlotValue = slotValue
				emptySlotTophash = &bucket.tophash[i]
			}
			if bucket.tophash[i] == tophash {
				// Could be an existing value that's the same.
				if keyEqual(key, slotKey, uintptr(m.keySize)) {
					// found same key, replace it
					memcpy(slotValue, value, uintptr(m.valueSize))
					return
				}
			}
		}
		bucket = bucket.next
	}
	if emptySlotKey != nil {
		m.count++
		memcpy(emptySlotKey, key, uintptr(m.keySize))
		memcpy(emptySlotValue, value, uintptr(m.valueSize))
		*emptySlotTophash = tophash
		return
	}
	panic("todo: hashmap: grow bucket")
}

// Get the value of a specified key, or zero the value if not found.
//go:nobounds
func hashmapGet(m *hashmap, key unsafe.Pointer, value unsafe.Pointer, hash uint32, keyEqual func(x, y unsafe.Pointer, n uintptr) bool) bool {
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
				if keyEqual(key, slotKey, uintptr(m.keySize)) {
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
func hashmapDelete(m *hashmap, key unsafe.Pointer, hash uint32, keyEqual func(x, y unsafe.Pointer, n uintptr) bool) {
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
				if keyEqual(key, slotKey, uintptr(m.keySize)) {
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
	numBuckets := uintptr(1) << m.bucketBits
	for {
		if it.bucketIndex >= 8 {
			// end of bucket, move to the next in the chain
			it.bucketIndex = 0
			it.bucket = it.bucket.next
		}
		if it.bucket == nil {
			if it.bucketNumber >= numBuckets {
				// went through all buckets
				return false
			}
			bucketSize := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*8
			bucketAddr := uintptr(m.buckets) + bucketSize*it.bucketNumber
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
		slotValueOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*uintptr(it.bucketIndex)
		slotValue := unsafe.Pointer(bucketAddr + slotValueOffset)
		memcpy(key, slotKey, uintptr(m.keySize))
		memcpy(value, slotValue, uintptr(m.valueSize))
		it.bucketIndex++

		return true
	}
}

// Hashmap with plain binary data keys (not containing strings etc.).

func hashmapBinarySet(m *hashmap, key, value unsafe.Pointer) {
	hash := hashmapHash(key, uintptr(m.keySize))
	hashmapSet(m, key, value, hash, memequal)
}

func hashmapBinaryGet(m *hashmap, key, value unsafe.Pointer) bool {
	hash := hashmapHash(key, uintptr(m.keySize))
	return hashmapGet(m, key, value, hash, memequal)
}

func hashmapBinaryDelete(m *hashmap, key unsafe.Pointer) {
	hash := hashmapHash(key, uintptr(m.keySize))
	hashmapDelete(m, key, hash, memequal)
}

// Hashmap with string keys (a common case).

func hashmapStringEqual(x, y unsafe.Pointer, n uintptr) bool {
	return *(*string)(x) == *(*string)(y)
}

func hashmapStringHash(s string) uint32 {
	_s := (*_string)(unsafe.Pointer(&s))
	return hashmapHash(unsafe.Pointer(_s.ptr), uintptr(_s.length))
}

func hashmapStringSet(m *hashmap, key string, value unsafe.Pointer) {
	hash := hashmapStringHash(key)
	hashmapSet(m, unsafe.Pointer(&key), value, hash, hashmapStringEqual)
}

func hashmapStringGet(m *hashmap, key string, value unsafe.Pointer) bool {
	hash := hashmapStringHash(key)
	return hashmapGet(m, unsafe.Pointer(&key), value, hash, hashmapStringEqual)
}

func hashmapStringDelete(m *hashmap, key string) {
	hash := hashmapStringHash(key)
	hashmapDelete(m, unsafe.Pointer(&key), hash, hashmapStringEqual)
}
