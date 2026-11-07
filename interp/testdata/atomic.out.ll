target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@rmw.i32 = local_unnamed_addr global i32 42
@rmw.i64 = local_unnamed_addr global i64 -5
@rmw.i8 = local_unnamed_addr global i8 44
@rmw.results = local_unnamed_addr global [12 x i32] [i32 12, i32 12, i32 10, i32 2, i32 11, i32 14, i32 -13, i32 3, i32 -7, i32 -7, i32 100, i32 200]
@rmw.max64 = local_unnamed_addr global i64 -5
@ptr.target = local_unnamed_addr global i32 0
@ptr.other = global i32 7
@ptr.value = local_unnamed_addr global ptr @ptr.other
@ptr.old = local_unnamed_addr global ptr null
@ptr.success = local_unnamed_addr global i1 true
@ptr.failure = local_unnamed_addr global i1 false
@once.done = local_unnamed_addr global i32 1
@once.count = local_unnamed_addr global i32 1
@runtime.value = global i32 3
@runtime.old = local_unnamed_addr global i32 0
@runtime.success = local_unnamed_addr global i1 false
@volatile.value = global i32 0
@float.value = global float 0.000000e+00

declare void @modifyAtomic(ptr) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @modifyAtomic(ptr @runtime.value)
  %0 = atomicrmw add ptr @runtime.value, i32 1 seq_cst, align 4
  store i32 %0, ptr @runtime.old, align 4
  %1 = cmpxchg ptr @runtime.value, i32 4, i32 5 seq_cst monotonic, align 4
  %success = extractvalue { i32, i1 } %1, 1
  store i1 %success, ptr @runtime.success, align 1
  %2 = atomicrmw volatile add ptr @volatile.value, i32 1 seq_cst, align 4
  call fastcc void @float.init(ptr undef)
  ret void
}

define i32 @"sync/atomic.AddInt32"(ptr %addr, i32 %delta, ptr %context) local_unnamed_addr {
  %old = atomicrmw add ptr %addr, i32 %delta syncscope("singlethread") seq_cst, align 4
  %new = add i32 %old, %delta
  ret i32 %new
}

define ptr @"sync/atomic.SwapPointer"(ptr %addr, ptr %new, ptr %context) local_unnamed_addr {
  %old = atomicrmw xchg ptr %addr, ptr %new syncscope("singlethread") seq_cst, align 8
  ret ptr %old
}

define i1 @"sync/atomic.CompareAndSwapPointer"(ptr %addr, ptr %old, ptr %new, ptr %context) local_unnamed_addr {
  %pair = cmpxchg ptr %addr, ptr %old, ptr %new syncscope("singlethread") seq_cst seq_cst, align 8
  %swapped = extractvalue { ptr, i1 } %pair, 1
  ret i1 %swapped
}

define i1 @"sync/atomic.CompareAndSwapUint32"(ptr %addr, i32 %old, i32 %new, ptr %context) local_unnamed_addr {
  %pair = cmpxchg ptr %addr, i32 %old, i32 %new syncscope("singlethread") seq_cst seq_cst, align 4
  %swapped = extractvalue { i32, i1 } %pair, 1
  ret i1 %swapped
}

define internal fastcc void @float.init(ptr %context) unnamed_addr {
  %old = atomicrmw fadd ptr @float.value, float 1.000000e+00 seq_cst, align 4
  ret void
}
