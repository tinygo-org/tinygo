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

	// Test atomic operations as deferred values.
	testDefer()
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

func testDefer() {
	n1 := int32(5)
	defer func() {
		println("deferred atomic add:", n1)
	}()
	defer atomic.AddInt32(&n1, 3)
}
