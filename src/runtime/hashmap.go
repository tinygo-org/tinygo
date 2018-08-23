package runtime

// This is a hashmap implementation for the map[T]T type.
// It is very rougly based on the implementation of the Go hashmap:
//
//     https://golang.org/src/runtime/hashmap.go

import (
	"unsafe"
)

// The underlying hashmap structure for Go.
type hashmap struct {
	next       *hashmap       // hashmap after evacuate (for iterators)
	buckets    unsafe.Pointer // pointer to array of buckets
	count      uint
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

// Get FNV-1a hash of this string.
//
// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function#FNV-1a_hash
func stringhash(s *string) uint32 {
	var result uint32 = 2166136261 // FNV offset basis
	for i := 0; i < len(*s); i++ {
		result ^= uint32((*s)[i])
		result *= 16777619 // FNV prime
	}
	return result
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

// Set a specified key to a given value. Grow the map if necessary.
func hashmapSet(m *hashmap, key string, value unsafe.Pointer) {
	hash := stringhash(&key)
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

	// See whether the key already exists somewhere.
	var emptySlotKey *string
	var emptySlotValue unsafe.Pointer
	var emptySlotTophash *byte
	for bucket != nil {
		for i := uintptr(0); i < 8; i++ {
			slotKeyOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*uintptr(i)
			slotKey := (*string)(unsafe.Pointer(bucketAddr + slotKeyOffset))
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
				if key == *slotKey {
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
		*emptySlotKey = key
		memcpy(emptySlotValue, value, uintptr(m.valueSize))
		*emptySlotTophash = tophash
		return
	}
	panic("todo: hashmap: grow bucket")
}

// Get the value of a specified key, or zero the value if not found.
func hashmapGet(m *hashmap, key string, value unsafe.Pointer) {
	hash := stringhash(&key)
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
			slotKey := (*string)(unsafe.Pointer(bucketAddr + slotKeyOffset))
			slotValueOffset := unsafe.Sizeof(hashmapBucket{}) + uintptr(m.keySize)*8 + uintptr(m.valueSize)*uintptr(i)
			slotValue := unsafe.Pointer(bucketAddr + slotValueOffset)
			if bucket.tophash[i] == tophash {
				// This could be the key we're looking for.
				if key == *slotKey {
					// Found the key, copy it.
					memcpy(value, slotValue, uintptr(m.valueSize))
					return
				}
			}
		}
		bucket = bucket.next
	}

	// Did not find the key.
	memzero(value, uintptr(m.valueSize))
}
