target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@foo.knownAtRuntime = local_unnamed_addr global i64 0
@bar.knownAtRuntime = local_unnamed_addr global i64 0
@baz.someGlobal = external local_unnamed_addr global [3 x { i64, i32 }]
@baz.someInt = local_unnamed_addr global i32 0
@x.atomicNum = local_unnamed_addr global i32 0
@x.volatileNum = global i32 0
@y.ready = local_unnamed_addr global i32 0
@z.bloom = global i64 0
@z.arr = global [32 x i8] zeroinitializer

declare void @externalCall(i64) local_unnamed_addr

declare i64 @ptrHash(i8* nocapture) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call fastcc void @baz.init(i8* undef)
  call fastcc void @foo.init(i8* undef)
  %val = load i64, i64* @foo.knownAtRuntime, align 8
  store i64 %val, i64* @bar.knownAtRuntime, align 8
  call void @externalCall(i64 3)
  store atomic i32 1, i32* @x.atomicNum seq_cst, align 4
  %x = load atomic i32, i32* @x.atomicNum seq_cst, align 4
  store i32 %x, i32* @x.atomicNum, align 4
  %y = load volatile i32, i32* @x.volatileNum, align 4
  store volatile i32 %y, i32* @x.volatileNum, align 4
  call fastcc void @y.init(i8* undef)
  call fastcc void @z.set(i8* bitcast (i64* @z.bloom to i8*), i8* bitcast (i64* @z.bloom to i8*))
  call fastcc void @z.setArr(i8* getelementptr inbounds ([32 x i8], [32 x i8]* @z.arr, i32 0, i32 0), i64 32, i8* bitcast (i64* @z.bloom to i8*))
  ret void
}

define internal fastcc void @foo.init(i8* %context) unnamed_addr {
  store i64 5, i64* @foo.knownAtRuntime, align 8
  unreachable
}

define internal fastcc void @baz.init(i8* %context) unnamed_addr {
  unreachable
}

define internal fastcc void @y.init(i8* %context) unnamed_addr {
entry:
  br label %loop

loop:                                             ; preds = %loop, %entry
  %val = load atomic i32, i32* @y.ready seq_cst, align 4
  %ready = icmp eq i32 %val, 1
  br i1 %ready, label %end, label %loop

end:                                              ; preds = %loop
  ret void
}

define internal fastcc void @z.setArr(i8* %arr, i64 %n, i8* %context) unnamed_addr {
entry:
  br label %loop
            
loop:                                             ; preds = %loop, %entry
  %prev = phi i64 [ %n, %entry ], [ %idx, %loop ]
  %idx = sub i64 %prev, 1
  %elem = getelementptr i8, i8* %arr, i64 %idx
  call fastcc void @z.set(i8* %elem, i8* %context)
  %done = icmp eq i64 %idx, 0
  br i1 %done, label %end, label %loop

end:                                              ; preds = %loop
  ret void
}

define internal fastcc void @z.set(i8* %ptr, i8* %context) unnamed_addr {
  %hash = call i64 @ptrHash(i8* %ptr)
  %index = lshr i64 %hash, 58
  %bit = shl i64 1, %index
  %bloom = bitcast i8* %context to i64*
  %old = load i64, i64* %bloom, align 8
  %new = or i64 %old, %bit
  store i64 %new, i64* %bloom, align 8
  ret void
}