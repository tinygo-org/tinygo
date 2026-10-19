target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@x = global i64 0
@ptr1 = local_unnamed_addr global ptr @x
@ptr2 = local_unnamed_addr global ptr @x
@split = local_unnamed_addr global [2 x i64] zeroinitializer
@inFloat = local_unnamed_addr global double 0.000000e+00
@inBytes = local_unnamed_addr global [16 x i8] zeroinitializer
@copied = local_unnamed_addr global [2 x i64] zeroinitializer
@element = local_unnamed_addr global [2 x i64] [i64 0, i64 ptrtoint (ptr @x to i64)]
@packed = local_unnamed_addr global <{ i8, ptr }> <{ i8 0, ptr @x }>
@copiedElement = local_unnamed_addr global [2 x ptr] [ptr null, ptr @x]

declare void @llvm.memcpy.p0.p0.i64(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i64, i1 immarg) #0

define void @runtime.initAll() unnamed_addr {
entry:
  store i64 ptrtoint (ptr @x to i64), ptr getelementptr inbounds (i8, ptr @split, i32 4), align 8
  store i64 ptrtoint (ptr @x to i64), ptr @inFloat, align 8
  store i64 ptrtoint (ptr @x to i64), ptr getelementptr inbounds (i8, ptr @inBytes, i32 8), align 8
  call void @llvm.memcpy.p0.p0.i64(ptr getelementptr inbounds (i8, ptr @copied, i64 4), ptr @ptr1, i64 8, i1 false)
  ret void
}

attributes #0 = { nocallback nofree nounwind willreturn memory(argmem: readwrite) }
