package main

import (
	"sync/atomic"
	"unsafe"
)

func main() {
	i32 := int32(-5)
	println("AddInt32:", atomic.AddInt32(&i32, 8), i32)

	i64 := int64(-5)
	println("AddInt64:", atomic.AddInt64(&i64, 8), i64)

	u32 := uint32(5)
	println("AddUint32:", atomic.AddUint32(&u32, 8), u32)

	u64 := uint64(5)
	println("AddUint64:", atomic.AddUint64(&u64, 8), u64)

	uptr := uintptr(5)
	println("AddUintptr:", uint64(atomic.AddUintptr(&uptr, 8)), uint64(uptr))

	println("SwapInt32:", atomic.SwapInt32(&i32, 33), i32)
	println("SwapInt64:", atomic.SwapInt64(&i64, 33), i64)
	println("SwapUint32:", atomic.SwapUint32(&u32, 33), u32)
	println("SwapUint64:", atomic.SwapUint64(&u64, 33), u64)
	println("SwapUintptr:", uint64(atomic.SwapUintptr(&uptr, 33)), uint64(uptr))
	ptr := unsafe.Pointer(&i32)
	println("SwapPointer:", atomic.SwapPointer(&ptr, unsafe.Pointer(&u32)) == unsafe.Pointer(&i32), ptr == unsafe.Pointer(&u32))

	i32 = int32(-5)
	println("CompareAndSwapInt32:", atomic.CompareAndSwapInt32(&i32, 5, 3), i32)
	println("CompareAndSwapInt32:", atomic.CompareAndSwapInt32(&i32, -5, 3), i32)

	i64 = int64(-5)
	println("CompareAndSwapInt64:", atomic.CompareAndSwapInt64(&i64, 5, 3), i64)
	println("CompareAndSwapInt64:", atomic.CompareAndSwapInt64(&i64, -5, 3), i64)

	u32 = uint32(5)
	println("CompareAndSwapUint32:", atomic.CompareAndSwapUint32(&u32, 4, 3), u32)
	println("CompareAndSwapUint32:", atomic.CompareAndSwapUint32(&u32, 5, 3), u32)

	u64 = uint64(5)
	println("CompareAndSwapUint64:", atomic.CompareAndSwapUint64(&u64, 4, 3), u64)
	println("CompareAndSwapUint64:", atomic.CompareAndSwapUint64(&u64, 5, 3), u64)

	uptr = uintptr(5)
	println("CompareAndSwapUintptr:", atomic.CompareAndSwapUintptr(&uptr, 4, 3), uint64(uptr))
	println("CompareAndSwapUintptr:", atomic.CompareAndSwapUintptr(&uptr, 5, 3), uint64(uptr))

	ptr = unsafe.Pointer(&i32)
	println("CompareAndSwapPointer:", atomic.CompareAndSwapPointer(&ptr, unsafe.Pointer(&u32), unsafe.Pointer(&i64)), ptr == unsafe.Pointer(&i32))
	println("CompareAndSwapPointer:", atomic.CompareAndSwapPointer(&ptr, unsafe.Pointer(&i32), unsafe.Pointer(&i64)), ptr == unsafe.Pointer(&i64))

	println("LoadInt32:", atomic.LoadInt32(&i32))
	println("LoadInt64:", atomic.LoadInt64(&i64))
	println("LoadUint32:", atomic.LoadUint32(&u32))
	println("LoadUint64:", atomic.LoadUint64(&u64))
	println("LoadUintptr:", uint64(atomic.LoadUintptr(&uptr)))
	println("LoadPointer:", atomic.LoadPointer(&ptr) == unsafe.Pointer(&i64))

	atomic.StoreInt32(&i32, -20)
	println("StoreInt32:", i32)

	atomic.StoreInt64(&i64, -20)
	println("StoreInt64:", i64)

	atomic.StoreUint32(&u32, 20)
	println("StoreUint32:", u32)

	atomic.StoreUint64(&u64, 20)
	println("StoreUint64:", u64)

	atomic.StoreUintptr(&uptr, 20)
	println("StoreUintptr:", uint64(uptr))

	atomic.StorePointer(&ptr, unsafe.Pointer(&uptr))
	println("StorePointer:", ptr == unsafe.Pointer(&uptr))

	// test atomic.Value load/store operations
	testValue(int(3), int(-2))
	testValue("", "foobar", "baz")

	// test deferred calls to atomic operations
	testDeferredAtomics()
}

func testValue(values ...interface{}) {
	var av atomic.Value
	for _, val := range values {
		av.Store(val)
		loadedVal := av.Load()
		if loadedVal != val {
			println("val store/load didn't work, expected", val, "but got", loadedVal)
		}
	}
}

// testDeferredAtomics tests direct deferred calls to atomic/*
//
// For example:
//   defer atomic.StoreUint32(&global, 1)
//
// Not:
//   defer func () {
//     // ...
// 		 atomic.StoreUint32(&global, 1)`
//     // ...
// 	 }()
func testDeferredAtomics() {
	println("Int32 --")
	testInt32()
	println("Uint32 --")
	testUint32()
	println("Int64 --")
	testInt64()
	println("Uint64 --")
	testUint64()
	println("Uintptr --")
	testUintptr()
	println("Pointer --")
	testPointer()
}

func deferredAddInt32(addr *int32, delta int32) int32 {
	defer atomic.AddInt32(addr, delta)
	return atomic.LoadInt32(addr)
}

func deferredSwapInt32(addr *int32, new int32) int32 {
	defer atomic.SwapInt32(addr, new)
	return atomic.LoadInt32(addr)
}

func deferredCompareAndSwapInt32(addr *int32, a, b int32) int32 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value to compare, we test it for completeness
	defer atomic.CompareAndSwapInt32(addr, a, b)
	return atomic.LoadInt32(addr)
}

func deferredLoadInt32(addr *int32) int32 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.LoadInt32(addr)
	return atomic.LoadInt32(addr)
}

func deferredStoreInt32(addr *int32, new int32) int32 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.StoreInt32(addr, new)
	return atomic.LoadInt32(addr)
}

func testInt32() {
	var value int32
	var before int32
	var after int32
	var success bool

	value = -9
	before = deferredAddInt32(&value, 11)
	after = atomic.LoadInt32(&value)
	success = before == -9 && after == 2
	println("deferred AddInt32:", success, before, after)

	value = -7
	before = deferredSwapInt32(&value, 100)
	after = atomic.LoadInt32(&value)
	success = before == -7 && after == 100
	println("deferred SwapInt32:", success, before, after)

	value = -9
	before = deferredCompareAndSwapInt32(&value, -9, 58)
	after = atomic.LoadInt32(&value)
	success = before == -9 && after == 58
	println("deferred CompareAndSwapInt32:", success, before, after)

	value = -11
	before = deferredLoadInt32(&value)
	after = atomic.LoadInt32(&value)
	success = before == -11 && after == -11
	println("deferred LoadInt32:", success, before, after)

	value = -11
	before = deferredStoreInt32(&value, 15)
	after = atomic.LoadInt32(&value)
	success = before == -11 && after == 15
	println("deferred StoreInt32:", success, before, after)
}

func deferredAddInt64(addr *int64, delta int64) int64 {
	defer atomic.AddInt64(addr, delta)
	return atomic.LoadInt64(addr)
}

func deferredSwapInt64(addr *int64, new int64) int64 {
	defer atomic.SwapInt64(addr, new)
	return atomic.LoadInt64(addr)
}

func deferredCompareAndSwapInt64(addr *int64, a, b int64) int64 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value to compare, we test it for completeness
	defer atomic.CompareAndSwapInt64(addr, a, b)
	return atomic.LoadInt64(addr)
}

func deferredLoadInt64(addr *int64) (old int64) {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.LoadInt64(addr)
	return atomic.LoadInt64(addr)
}

func deferredStoreInt64(addr *int64, new int64) int64 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.StoreInt64(addr, new)
	return atomic.LoadInt64(addr)
}

func testInt64() {
	var value int64
	var before int64
	var after int64
	var success bool

	value = -9
	before = deferredAddInt64(&value, 11)
	after = atomic.LoadInt64(&value)
	success = before == -9 && after == 2
	println("deferred AddInt64:", success, before, after)

	value = -7
	before = deferredSwapInt64(&value, 100)
	after = atomic.LoadInt64(&value)
	success = before == -7 && after == 100
	println("deferred SwapInt64:", success, before, after)

	value = -9
	before = deferredCompareAndSwapInt64(&value, -9, 58)
	after = atomic.LoadInt64(&value)
	success = before == -9 && after == 58
	println("deferred CompareAndSwapInt64:", success, before, after)

	value = -11
	before = deferredLoadInt64(&value)
	after = atomic.LoadInt64(&value)
	success = before == -11 && after == -11
	println("deferred LoadInt64:", success, before, after)

	value = -11
	before = deferredStoreInt64(&value, 15)
	after = atomic.LoadInt64(&value)
	success = before == -11 && after == 15
	println("deferred StoreInt64:", success, before, after)
}

func deferredAddUint32(addr *uint32, delta uint32) uint32 {
	defer atomic.AddUint32(addr, delta)
	return atomic.LoadUint32(addr)
}

func deferredSwapUint32(addr *uint32, new uint32) uint32 {
	defer atomic.SwapUint32(addr, new)
	return atomic.LoadUint32(addr)
}

func deferredCompareAndSwapUint32(addr *uint32, a, b uint32) uint32 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value to compare, we test it for completeness
	defer atomic.CompareAndSwapUint32(addr, a, b)
	return atomic.LoadUint32(addr)
}

func deferredLoadUint32(addr *uint32) uint32 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.LoadUint32(addr)
	return atomic.LoadUint32(addr)
}

func deferredStoreUint32(addr *uint32, new uint32) uint32 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.StoreUint32(addr, new)
	return atomic.LoadUint32(addr)
}

func testUint32() {
	var value uint32
	var before uint32
	var after uint32
	var success bool

	value = 9
	before = deferredAddUint32(&value, 11)
	after = atomic.LoadUint32(&value)
	success = before == 9 && after == 20
	println("deferred AddUint32:", success, before, after)

	value = 7
	before = deferredSwapUint32(&value, 100)
	after = atomic.LoadUint32(&value)
	success = before == 7 && after == 100
	println("deferred SwapUint32:", success, before, after)

	value = 9
	before = deferredCompareAndSwapUint32(&value, 9, 58)
	after = atomic.LoadUint32(&value)
	success = before == 9 && after == 58
	println("deferred CompareAndSwapUint32:", success, before, after)

	value = 11
	before = deferredLoadUint32(&value)
	after = atomic.LoadUint32(&value)
	success = before == 11 && after == 11
	println("deferred LoadUint32:", success, before, after)

	value = 11
	before = deferredStoreUint32(&value, 15)
	after = atomic.LoadUint32(&value)
	success = before == 11 && after == 15
	println("deferred StoreUint32:", success, before, after)
}

func deferredAddUint64(addr *uint64, delta uint64) uint64 {
	defer atomic.AddUint64(addr, delta)
	return atomic.LoadUint64(addr)
}

func deferredSwapUint64(addr *uint64, new uint64) uint64 {
	defer atomic.SwapUint64(addr, new)
	return atomic.LoadUint64(addr)
}

func deferredCompareAndSwapUint64(addr *uint64, a, b uint64) uint64 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value to compare, we test it for completeness
	defer atomic.CompareAndSwapUint64(addr, a, b)
	return atomic.LoadUint64(addr)
}

func deferredLoadUint64(addr *uint64) (old uint64) {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.LoadUint64(addr)
	return atomic.LoadUint64(addr)
}

func deferredStoreUint64(addr *uint64, new uint64) uint64 {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.StoreUint64(addr, new)
	return atomic.LoadUint64(addr)
}

func testUint64() {
	var value uint64
	var before uint64
	var after uint64
	var success bool

	value = 9
	before = deferredAddUint64(&value, 11)
	after = atomic.LoadUint64(&value)
	success = before == 9 && after == 20
	println("deferred AddUint64:", success, before, after)

	value = 7
	before = deferredSwapUint64(&value, 100)
	after = atomic.LoadUint64(&value)
	success = before == 7 && after == 100
	println("deferred SwapUint64:", success, before, after)

	value = 9
	before = deferredCompareAndSwapUint64(&value, 9, 58)
	after = atomic.LoadUint64(&value)
	success = before == 9 && after == 58
	println("deferred CompareAndSwapUint64:", success, before, after)

	value = 11
	before = deferredLoadUint64(&value)
	after = atomic.LoadUint64(&value)
	success = before == 11 && after == 11
	println("deferred LoadUint64:", success, before, after)

	value = 11
	before = deferredStoreUint64(&value, 15)
	after = atomic.LoadUint64(&value)
	success = before == 11 && after == 15
	println("deferred StoreUint64:", success, before, after)
}

func deferredSwapUintptr(addr *uintptr, new uintptr) uintptr {
	defer atomic.SwapUintptr(addr, new)
	return atomic.LoadUintptr(addr)
}

func deferredCompareAndSwapUintptr(addr *uintptr, a, b uintptr) uintptr {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value to compare, we test it for completeness
	defer atomic.CompareAndSwapUintptr(addr, a, b)
	return atomic.LoadUintptr(addr)
}

func deferredLoadUintptr(addr *uintptr) uintptr {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.LoadUintptr(addr)
	return atomic.LoadUintptr(addr)
}

func deferredStoreUintptr(addr *uintptr, new uintptr) uintptr {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.StoreUintptr(addr, new)
	return atomic.LoadUintptr(addr)
}

func testUintptr() {
	var value uintptr
	var before uintptr
	var after uintptr
	var success bool

	value = 7
	before = deferredSwapUintptr(&value, 100)
	after = atomic.LoadUintptr(&value)
	success = before == 7 && after == 100
	println("deferred SwapUintptr:", success, before, after)

	value = 9
	before = deferredCompareAndSwapUintptr(&value, 9, 58)
	after = atomic.LoadUintptr(&value)
	success = before == 9 && after == 58
	println("deferred CompareAndSwapUintptr:", success, before, after)

	value = 11
	before = deferredLoadUintptr(&value)
	after = atomic.LoadUintptr(&value)
	success = before == 11 && after == 11
	println("deferred LoadUintptr:", success, before, after)

	value = 11
	before = deferredStoreUintptr(&value, 15)
	after = atomic.LoadUintptr(&value)
	success = before == 11 && after == 15
	println("deferred StoreUintptr:", success, before, after)
}

func deferredSwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer {
	defer atomic.SwapPointer(addr, new)
	return atomic.LoadPointer(addr)
}

func deferredCompareAndSwapPointer(addr *unsafe.Pointer, a, b unsafe.Pointer) unsafe.Pointer {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value to compare, we test it for completeness
	defer atomic.CompareAndSwapPointer(addr, a, b)
	return atomic.LoadPointer(addr)
}

func deferredLoadPointer(addr *unsafe.Pointer) (old unsafe.Pointer) {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.LoadPointer(addr)
	return atomic.LoadPointer(addr)
}

func deferredStorePointer(addr *unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer {
	// Even though this function doesn't make a lot of sense to be deferred, since you would
	// not be able to use the value loaded, we test it for completeness
	defer atomic.StorePointer(addr, new)
	return atomic.LoadPointer(addr)
}

func testPointer() {
	var value uint64
	var valueP unsafe.Pointer
	var after unsafe.Pointer
	var success bool

	valueIn := func(p unsafe.Pointer) uint64 {
		return *(*uint64)(p)
	}

	value = 7
	valueP = unsafe.Pointer(&value)
	var swapTo uintptr = 100
	swapToP := unsafe.Pointer(&swapTo)
	deferredSwapPointer(&valueP, swapToP)
	after = atomic.LoadPointer(&valueP)
	success = value == 7 && after == swapToP && valueIn(swapToP) == 100
	println("deferred SwapPointer:", success)

	value = 9
	valueP = unsafe.Pointer(&value)
	swapTo = 58
	swapToP = unsafe.Pointer(&swapTo)
	deferredCompareAndSwapPointer(&valueP, valueP, swapToP)
	after = atomic.LoadPointer(&valueP)
	success = value == 9 && after == swapToP && valueIn(swapToP) == 58
	println("deferred CompareAndSwapPointer:", success)

	value = 11
	valueP = unsafe.Pointer(&value)
	deferredLoadPointer(&valueP)
	after = atomic.LoadPointer(&valueP)
	success = value == 11 && after == valueP && valueIn(valueP) == 11
	println("deferred LoadPointer:", success)

	value = 11
	valueP = unsafe.Pointer(&value)
	swapTo = 15
	swapToP = unsafe.Pointer(&swapTo)
	deferredStorePointer(&valueP, unsafe.Pointer(&swapTo))
	after = atomic.LoadPointer(&valueP)
	success = value == 11 && after == swapToP && valueIn(swapToP) == 15
	println("deferred StorePointer:", success)
}
