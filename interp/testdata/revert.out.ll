target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@foo.knownAtRuntime = local_unnamed_addr global i64 0
@bar.knownAtRuntime = local_unnamed_addr global i64 0
@baz.someGlobal = external local_unnamed_addr global [3 x { i64, i32 }]
@baz.someInt = local_unnamed_addr global i32 0
@x.atomicNum = local_unnamed_addr global i32 1
@x.volatileNum = global i32 0
@y.ready = local_unnamed_addr global i32 0
@z.bloom = global i64 0
@z.arr = global [32 x i8] zeroinitializer
@atomic.value = global i32 11
@atomic.old = local_unnamed_addr global i32 7
@atomic.success = local_unnamed_addr global i1 true
@atomic.failure = local_unnamed_addr global i1 false
@atomic.seen = local_unnamed_addr global i32 11
@atomic.runtimeOld = local_unnamed_addr global i32 0
@atomic.runtimeSuccess = local_unnamed_addr global i1 false
@atomic.runtimeSeen = local_unnamed_addr global i32 0

declare void @externalCall(i64) local_unnamed_addr

declare i64 @ptrHash(ptr nocapture) local_unnamed_addr

declare void @modifyAtomic(ptr) local_unnamed_addr

declare i32 @"sync/atomic.SwapUint32"(ptr, i32, ptr) local_unnamed_addr

declare i1 @"sync/atomic.CompareAndSwapUint32"(ptr, i32, i32, ptr) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call fastcc void @baz.init(ptr undef)
  call fastcc void @foo.init(ptr undef)
  %val = load i64, ptr @foo.knownAtRuntime, align 8
  store i64 %val, ptr @bar.knownAtRuntime, align 8
  call void @externalCall(i64 3)
  %y = load volatile i32, ptr @x.volatileNum, align 4
  store volatile i32 %y, ptr @x.volatileNum, align 4
  call void @modifyAtomic(ptr @atomic.value)
  %runtimeOld = call i32 @"sync/atomic.SwapUint32"(ptr @atomic.value, i32 15, ptr undef)
  store i32 %runtimeOld, ptr @atomic.runtimeOld, align 4
  %runtimeSuccess = call i1 @"sync/atomic.CompareAndSwapUint32"(ptr @atomic.value, i32 15, i32 17, ptr undef)
  store i1 %runtimeSuccess, ptr @atomic.runtimeSuccess, align 1
  %runtimeSeen = load atomic i32, ptr @atomic.value seq_cst, align 4
  store i32 %runtimeSeen, ptr @atomic.runtimeSeen, align 4
  store atomic i32 19, ptr @atomic.value seq_cst, align 4
  call fastcc void @y.init(ptr undef)
  call fastcc void @z.set(ptr @z.bloom, ptr @z.bloom)
  call fastcc void @z.setArr(ptr @z.arr, i64 32, ptr @z.bloom)
  ret void
}

define internal fastcc void @foo.init(ptr %context) unnamed_addr {
  store i64 5, ptr @foo.knownAtRuntime, align 8
  unreachable
}

define internal fastcc void @baz.init(ptr %context) unnamed_addr {
  unreachable
}

define internal fastcc void @y.init(ptr %context) unnamed_addr {
entry:
  br label %loop

loop:                                             ; preds = %loop, %entry
  %val = load atomic i32, ptr @y.ready seq_cst, align 4
  %ready = icmp eq i32 %val, 1
  br i1 %ready, label %end, label %loop

end:                                              ; preds = %loop
  ret void
}

define internal fastcc void @z.setArr(ptr %arr, i64 %n, ptr %context) unnamed_addr {
entry:
  br label %loop

loop:                                             ; preds = %loop, %entry
  %prev = phi i64 [ %n, %entry ], [ %idx, %loop ]
  %idx = sub i64 %prev, 1
  %elem = getelementptr i8, ptr %arr, i64 %idx
  call fastcc void @z.set(ptr %elem, ptr %context)
  %done = icmp eq i64 %idx, 0
  br i1 %done, label %end, label %loop

end:                                              ; preds = %loop
  ret void
}

define internal fastcc void @z.set(ptr %ptr, ptr %context) unnamed_addr {
  %hash = call i64 @ptrHash(ptr %ptr)
  %index = lshr i64 %hash, 58
  %bit = shl i64 1, %index
  %old = load i64, ptr %context, align 8
  %new = or i64 %old, %bit
  store i64 %new, ptr %context, align 8
  ret void
}
