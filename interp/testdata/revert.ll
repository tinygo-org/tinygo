target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare void @externalCall(i64)

declare i64 @ptrHash(i8* nocapture)

@foo.knownAtRuntime = global i64 0
@bar.knownAtRuntime = global i64 0
@baz.someGlobal = external global [3 x {i64, i32}]
@baz.someInt = global i32 0
@x.atomicNum = global i32 0
@x.volatileNum = global i32 0
@y.ready = global i32 0
@z.bloom = global i64 0
@z.arr = global [32 x i8] zeroinitializer

define void @runtime.initAll() unnamed_addr {
entry:
  call void @baz.init(i8* undef)
  call void @foo.init(i8* undef)
  call void @bar.init(i8* undef)
  call void @main.init(i8* undef)
  call void @x.init(i8* undef)
  call void @y.init(i8* undef)
  call void @z.init(i8* undef)
  ret void
}

define internal void @foo.init(i8* %context) unnamed_addr {
  store i64 5, i64* @foo.knownAtRuntime
  unreachable ; this triggers a revert of @foo.init.
}

define internal void @bar.init(i8* %context) unnamed_addr {
  %val = load i64, i64* @foo.knownAtRuntime
  store i64 %val, i64* @bar.knownAtRuntime
  ret void
}

define internal void @baz.init(i8* %context) unnamed_addr {
  ; Test extractvalue/insertvalue with more than one index.
  %val = load [3 x {i64, i32}], [3 x {i64, i32}]* @baz.someGlobal
  %part = extractvalue [3 x {i64, i32}] %val, 0, 1
  %val2 = insertvalue [3 x {i64, i32}] %val, i32 5, 2, 1
  unreachable ; trigger revert
}

define internal void @main.init(i8* %context) unnamed_addr {
entry:
  call void @externalCall(i64 3)
  ret void
}


define internal void @x.init(i8* %context) unnamed_addr {
  ; Test atomic and volatile memory accesses.
  store atomic i32 1, i32* @x.atomicNum seq_cst, align 4
  %x = load atomic i32, i32* @x.atomicNum seq_cst, align 4
  store i32 %x, i32* @x.atomicNum
  %y = load volatile i32, i32* @x.volatileNum
  store volatile i32 %y, i32* @x.volatileNum
  ret void
}

define internal void @y.init(i8* %context) unnamed_addr {
entry:
  br label %loop

loop:
  ; Test a wait-loop.
  ; This function must be reverted.
  %val = load atomic i32, i32* @y.ready seq_cst, align 4
  %ready = icmp eq i32 %val, 1
  br i1 %ready, label %end, label %loop

end:
  ret void
}

define internal void @z.init(i8* %context) unnamed_addr {
  %bloom = bitcast i64* @z.bloom to i8*

  ; This can be safely expanded.
  call void @z.setArr(i8* %bloom, i64 1, i8* %bloom)

  ; This call should be reverted to prevent unrolling.
  call void @z.setArr(i8* bitcast ([32 x i8]* @z.arr to i8*), i64 32, i8* %bloom)

  ret void
}

define internal void @z.setArr(i8* %arr, i64 %n, i8* %context) unnamed_addr {
entry:
  br label %loop

loop:
  %prev = phi i64 [ %n, %entry ], [ %idx, %loop ]
  %idx = sub i64 %prev, 1
  %elem = getelementptr i8, i8* %arr, i64 %idx
  call void @z.set(i8* %elem, i8* %context)
  %done = icmp eq i64 %idx, 0
  br i1 %done, label %end, label %loop

end:
  ret void
}

define internal void @z.set(i8* %ptr, i8* %context) unnamed_addr {
  ; Insert the pointer into the Bloom filter.
  %hash = call i64 @ptrHash(i8* %ptr)
  %index = lshr i64 %hash, 58
  %bit = shl i64 1, %index
  %bloom = bitcast i8* %context to i64*
  %old = load i64, i64* %bloom
  %new = or i64 %old, %bit
  store i64 %new, i64* %bloom
  ret void
}
