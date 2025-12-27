target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@moveDst = local_unnamed_addr global [3 x i8] c"foo"
@copyDst = local_unnamed_addr global [3 x i8] c"foo"
@externalSrc = external local_unnamed_addr global [2 x i8]
@moveExternalDst = local_unnamed_addr global [2 x i8] zeroinitializer
@moveEscapedSrc = global [4 x i8] c"abcd"
@moveEscapedDst = local_unnamed_addr global [4 x i8] zeroinitializer
@volatileSrc = global [2 x i8] c"xy"
@volatileDst = global [2 x i8] zeroinitializer

declare void @use(ptr) local_unnamed_addr

define void @runtime.initAll() local_unnamed_addr {
  call void @llvm.memmove.p0.p0.i64(ptr @moveExternalDst, ptr @externalSrc, i64 2, i1 false)
  call void @use(ptr @moveEscapedSrc)
  call void @llvm.memmove.p0.p0.i64(ptr @moveEscapedDst, ptr @moveEscapedSrc, i64 4, i1 false)
  call void @llvm.memcpy.p0.p0.i64(ptr @volatileDst, ptr @volatileSrc, i64 2, i1 true)
  ret void
}

declare void @llvm.memmove.p0.p0.i64(ptr nocapture writeonly, ptr nocapture readonly, i64, i1 immarg) #0

declare void @llvm.memcpy.p0.p0.i64(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i64, i1 immarg) #0

attributes #0 = { nocallback nofree nounwind willreturn memory(argmem: readwrite) }
