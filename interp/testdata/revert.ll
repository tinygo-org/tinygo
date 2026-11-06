target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare void @externalCall(i64)

declare i64 @ptrHash(ptr nocapture)

declare void @modifyAtomic(ptr)
declare i32 @"sync/atomic.SwapUint32"(ptr, i32, ptr)
declare i1 @"sync/atomic.CompareAndSwapUint32"(ptr, i32, i32, ptr)

@foo.knownAtRuntime = global i64 0
@bar.knownAtRuntime = global i64 0
@baz.someGlobal = external global [3 x {i64, i32}]
@baz.someInt = global i32 0
@x.atomicNum = global i32 0
@x.volatileNum = global i32 0
@y.ready = global i32 0
@z.bloom = global i64 0
@z.arr = global [32 x i8] zeroinitializer
@atomic.value = global i32 0
@atomic.old = global i32 0
@atomic.success = global i1 false
@atomic.failure = global i1 false
@atomic.seen = global i32 0
@atomic.runtimeOld = global i32 0
@atomic.runtimeSuccess = global i1 false
@atomic.runtimeSeen = global i32 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @baz.init(ptr undef)
  call void @foo.init(ptr undef)
  call void @bar.init(ptr undef)
  call void @main.init(ptr undef)
  call void @x.init(ptr undef)
  call void @atomic.init(ptr undef)
  call void @y.init(ptr undef)
  call void @z.init(ptr undef)
  ret void
}

define internal void @foo.init(ptr %context) unnamed_addr {
  store i64 5, ptr @foo.knownAtRuntime
  unreachable ; this triggers a revert of @foo.init.
}

define internal void @bar.init(ptr %context) unnamed_addr {
  %val = load i64, ptr @foo.knownAtRuntime
  store i64 %val, ptr @bar.knownAtRuntime
  ret void
}

define internal void @baz.init(ptr %context) unnamed_addr {
  ; Test extractvalue/insertvalue with more than one index.
  %val = load [3 x {i64, i32}], ptr @baz.someGlobal
  %part = extractvalue [3 x {i64, i32}] %val, 0, 1
  %val2 = insertvalue [3 x {i64, i32}] %val, i32 5, 2, 1
  unreachable ; trigger revert
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  call void @externalCall(i64 3)
  ret void
}


define internal void @x.init(ptr %context) unnamed_addr {
  ; Test atomic and volatile memory accesses.
  store atomic i32 1, ptr @x.atomicNum seq_cst, align 4
  %x = load atomic i32, ptr @x.atomicNum seq_cst, align 4
  store i32 %x, ptr @x.atomicNum
  %y = load volatile i32, ptr @x.volatileNum
  store volatile i32 %y, ptr @x.volatileNum
  ret void
}

define internal void @y.init(ptr %context) unnamed_addr {
entry:
  br label %loop

loop:
  ; Test a wait-loop.
  ; This function must be reverted.
  %val = load atomic i32, ptr @y.ready seq_cst, align 4
  %ready = icmp eq i32 %val, 1
  br i1 %ready, label %end, label %loop

end:
  ret void
}

define internal void @atomic.init(ptr %context) unnamed_addr {
  store atomic i32 7, ptr @atomic.value seq_cst, align 4
  %old = call i32 @"sync/atomic.SwapUint32"(ptr @atomic.value, i32 9, ptr undef)
  store i32 %old, ptr @atomic.old
  %success = call i1 @"sync/atomic.CompareAndSwapUint32"(ptr @atomic.value, i32 9, i32 11, ptr undef)
  store i1 %success, ptr @atomic.success
  %failure = call i1 @"sync/atomic.CompareAndSwapUint32"(ptr @atomic.value, i32 9, i32 13, ptr undef)
  store i1 %failure, ptr @atomic.failure
  %seen = load atomic i32, ptr @atomic.value seq_cst, align 4
  store i32 %seen, ptr @atomic.seen
  call void @modifyAtomic(ptr @atomic.value)
  %runtimeOld = call i32 @"sync/atomic.SwapUint32"(ptr @atomic.value, i32 15, ptr undef)
  store i32 %runtimeOld, ptr @atomic.runtimeOld
  %runtimeSuccess = call i1 @"sync/atomic.CompareAndSwapUint32"(ptr @atomic.value, i32 15, i32 17, ptr undef)
  store i1 %runtimeSuccess, ptr @atomic.runtimeSuccess
  %runtimeSeen = load atomic i32, ptr @atomic.value seq_cst, align 4
  store i32 %runtimeSeen, ptr @atomic.runtimeSeen
  store atomic i32 19, ptr @atomic.value seq_cst, align 4
  ret void
}

define internal void @z.init(ptr %context) unnamed_addr {
  ; This can be safely expanded.
  call void @z.setArr(ptr @z.bloom, i64 1, ptr @z.bloom)

  ; This call should be reverted to prevent unrolling.
  call void @z.setArr(ptr @z.arr, i64 32, ptr @z.bloom)

  ret void
}

define internal void @z.setArr(ptr %arr, i64 %n, ptr %context) unnamed_addr {
entry:
  br label %loop

loop:
  %prev = phi i64 [ %n, %entry ], [ %idx, %loop ]
  %idx = sub i64 %prev, 1
  %elem = getelementptr i8, ptr %arr, i64 %idx
  call void @z.set(ptr %elem, ptr %context)
  %done = icmp eq i64 %idx, 0
  br i1 %done, label %end, label %loop

end:
  ret void
}

define internal void @z.set(ptr %ptr, ptr %context) unnamed_addr {
  ; Insert the pointer into the Bloom filter.
  %hash = call i64 @ptrHash(ptr %ptr)
  %index = lshr i64 %hash, 58
  %bit = shl i64 1, %index
  %old = load i64, ptr %context
  %new = or i64 %old, %bit
  store i64 %new, ptr %context
  ret void
}
