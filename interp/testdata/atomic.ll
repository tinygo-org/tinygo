target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare void @modifyAtomic(ptr)

@rmw.i32 = global i32 0
@rmw.i64 = global i64 0
@rmw.i8 = global i8 0
@rmw.results = global [12 x i32] zeroinitializer
@rmw.max64 = global i64 0
@ptr.target = global i32 0
@ptr.other = global i32 0
@ptr.value = global ptr null
@ptr.old = global ptr null
@ptr.success = global i1 false
@ptr.failure = global i1 false
@once.done = global i32 0
@once.count = global i32 0
@runtime.value = global i32 0
@runtime.old = global i32 0
@runtime.success = global i1 false
@volatile.value = global i32 0
@float.value = global float 0.0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @rmw.init(ptr undef)
  call void @ptr.init(ptr undef)
  call void @once.init(ptr undef)
  call void @runtime.init(ptr undef)
  call void @volatile.init(ptr undef)
  call void @float.init(ptr undef)
  ret void
}

; Same shape as the sync/atomic bodies created in compiler/atomic.go.
define i32 @"sync/atomic.AddInt32"(ptr %addr, i32 %delta, ptr %context) {
  %old = atomicrmw add ptr %addr, i32 %delta syncscope("singlethread") seq_cst, align 4
  %new = add i32 %old, %delta
  ret i32 %new
}

define ptr @"sync/atomic.SwapPointer"(ptr %addr, ptr %new, ptr %context) {
  %old = atomicrmw xchg ptr %addr, ptr %new syncscope("singlethread") seq_cst, align 8
  ret ptr %old
}

define i1 @"sync/atomic.CompareAndSwapPointer"(ptr %addr, ptr %old, ptr %new, ptr %context) {
  %pair = cmpxchg ptr %addr, ptr %old, ptr %new syncscope("singlethread") seq_cst seq_cst, align 8
  %swapped = extractvalue { ptr, i1 } %pair, 1
  ret i1 %swapped
}

define i1 @"sync/atomic.CompareAndSwapUint32"(ptr %addr, i32 %old, i32 %new, ptr %context) {
  %pair = cmpxchg ptr %addr, i32 %old, i32 %new syncscope("singlethread") seq_cst seq_cst, align 4
  %swapped = extractvalue { i32, i1 } %pair, 1
  ret i1 %swapped
}

define internal void @rmw.init(ptr %context) unnamed_addr {
  %add = call i32 @"sync/atomic.AddInt32"(ptr @rmw.i32, i32 12, ptr undef)
  store i32 %add, ptr @rmw.results
  %sub = atomicrmw sub ptr @rmw.i32, i32 2 seq_cst, align 4
  store i32 %sub, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 1)
  %and = atomicrmw and ptr @rmw.i32, i32 6 seq_cst, align 4
  store i32 %and, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 2)
  %or = atomicrmw or ptr @rmw.i32, i32 9 seq_cst, align 4
  store i32 %or, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 3)
  %xor = atomicrmw xor ptr @rmw.i32, i32 5 seq_cst, align 4
  store i32 %xor, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 4)
  %nand = atomicrmw nand ptr @rmw.i32, i32 12 seq_cst, align 4
  store i32 %nand, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 5)
  %max = atomicrmw max ptr @rmw.i32, i32 3 seq_cst, align 4
  store i32 %max, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 6)
  %min = atomicrmw min ptr @rmw.i32, i32 -7 seq_cst, align 4
  store i32 %min, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 7)
  %umax = atomicrmw umax ptr @rmw.i32, i32 4 seq_cst, align 4
  store i32 %umax, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 8)
  %umin = atomicrmw umin ptr @rmw.i32, i32 100 seq_cst, align 4
  store i32 %umin, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 9)
  %xchg = atomicrmw xchg ptr @rmw.i32, i32 42 seq_cst, align 4
  store i32 %xchg, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 10)
  %i8 = atomicrmw add ptr @rmw.i8, i8 200 monotonic, align 1
  %i8b = atomicrmw add ptr @rmw.i8, i8 100 monotonic, align 1
  %i8c = zext i8 %i8b to i32
  store i32 %i8c, ptr getelementptr ([12 x i32], ptr @rmw.results, i32 0, i32 11)
  %i64 = atomicrmw sub ptr @rmw.i64, i64 5 seq_cst, align 8
  %max64 = atomicrmw max ptr @rmw.i64, i64 -10 seq_cst, align 8
  store i64 %max64, ptr @rmw.max64
  ret void
}

define internal void @ptr.init(ptr %context) unnamed_addr {
  %old = call ptr @"sync/atomic.SwapPointer"(ptr @ptr.value, ptr @ptr.target, ptr undef)
  store ptr %old, ptr @ptr.old
  %success = call i1 @"sync/atomic.CompareAndSwapPointer"(ptr @ptr.value, ptr @ptr.target, ptr @ptr.other, ptr undef)
  store i1 %success, ptr @ptr.success
  %failure = call i1 @"sync/atomic.CompareAndSwapPointer"(ptr @ptr.value, ptr @ptr.target, ptr null, ptr undef)
  store i1 %failure, ptr @ptr.failure
  %loaded = load atomic ptr, ptr @ptr.value seq_cst, align 8
  store i32 7, ptr %loaded
  ret void
}

; Like sync.Once, the second call must not run the body.
define internal void @once.init(ptr %context) unnamed_addr {
  call void @once.do(ptr undef)
  call void @once.do(ptr undef)
  ret void
}

define internal void @once.do(ptr %context) unnamed_addr {
entry:
  %first = call i1 @"sync/atomic.CompareAndSwapUint32"(ptr @once.done, i32 0, i32 1, ptr undef)
  br i1 %first, label %body, label %done

body:
  %count = load i32, ptr @once.count
  %count1 = add i32 %count, 1
  store i32 %count1, ptr @once.count
  br label %done

done:
  ret void
}

define internal void @runtime.init(ptr %context) unnamed_addr {
  store atomic i32 3, ptr @runtime.value seq_cst, align 4
  call void @modifyAtomic(ptr @runtime.value)
  %old = atomicrmw add ptr @runtime.value, i32 1 seq_cst, align 4
  store i32 %old, ptr @runtime.old
  %pair = cmpxchg ptr @runtime.value, i32 4, i32 5 seq_cst monotonic, align 4
  %success = extractvalue { i32, i1 } %pair, 1
  store i1 %success, ptr @runtime.success
  ret void
}

define internal void @volatile.init(ptr %context) unnamed_addr {
  %old = atomicrmw volatile add ptr @volatile.value, i32 1 seq_cst, align 4
  ret void
}

define internal void @float.init(ptr %context) unnamed_addr {
  %old = atomicrmw fadd ptr @float.value, float 1.0 seq_cst, align 4
  ret void
}
