target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@toSet.0 = internal global [10 x i8] undef
@set.ret.0 = internal global i8* undef
@toSet.1 = internal global [3 x i8] undef
@toSet.2 = internal global [5 x i8] undef

declare i8* @memset(i8* nocapture writeonly, i32, i64) argmemonly nofree nounwind willreturn writeonly

declare void @llvm.memset.p0i8.i64(i8* nocapture writeonly, i8, i64, i1 immarg) argmemonly nofree nounwind willreturn writeonly

define internal void @testMemSet() {
  ; The C memset function truncates the integer argument and returns the destination.
  %memset.ret = call i8* @memset(i8* bitcast ([10 x i8]* @toSet.0 to i8*), i32 321, i64 10)
  store i8* %memset.ret, i8** @set.ret.0
  ; Test the LLVM memset intrinsic.
  call void @llvm.memset.p0i8.i64(i8* bitcast ([3 x i8]* @toSet.1 to i8*), i8 66, i64 3, i1 false)
  ; A volatile memset should be passed through as-is.
  call void @llvm.memset.p0i8.i64(i8* bitcast ([5 x i8]* @toSet.2 to i8*), i8 67, i64 5, i1 true)
  ret void
}

@from = internal constant [94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
@to.1 = internal global [8 x [94 x i8]] undef

declare i8* @memmove(i8* nocapture writeonly, i8* nocapture readonly, i64) argmemonly nofree nounwind willreturn

declare i8* @memcpy(i8* nocapture writeonly, i8* nocapture readonly, i64) argmemonly nofree nounwind willreturn

define internal void @testMemMove1() {
entry: ; Fill to.1 with copies of "from".
  %cpy.ret = call i8* @memcpy(i8* bitcast ([8 x [94 x i8]]* @to.1 to i8*), i8* bitcast ([94 x i8]* @from to i8*), i64 94)
  %dst.init = getelementptr i8, i8* %cpy.ret, i64 94
  br label %loop
loop:
  %dst = phi i8* [ %dst.init, %entry ], [ %dst.2, %loop ]
  %dst.addr = ptrtoint i8* %dst to i64
  %n = sub i64 %dst.addr, ptrtoint ([8 x [94 x i8]]* @to.1 to i64)
  %dst.1 = call i8* @memmove(i8* %dst, i8* bitcast ([8 x [94 x i8]]* @to.1 to i8*), i64 %n)
  %dst.2 = getelementptr i8, i8* %dst.1, i64 %n
  %ok = icmp ult i64 %n, 376
  br i1 %ok, label %loop, label %done
done:
  ret void
}

declare void @llvm.memcpy.p0i8.p0i8.i64(i8* nocapture writeonly, i8* nocapture readonly, i64, i1 immarg) argmemonly nofree nounwind willreturn

declare void @llvm.memmove.p0i8.p0i8.i64(i8* nocapture writeonly, i8* nocapture readonly, i64, i1 immarg) argmemonly nofree nounwind willreturn

@ext = external global i64*

define internal void @testMemMove2() {
  %a = alloca i64*, i32 3
  %v = load i64*, i64** @ext
  store i64* %v, i64** %a
  %a.i8ptr = bitcast i64** %a to i8*
  %a.1 = getelementptr i64*, i64** %a, i64 1
  %a.1.i8ptr = bitcast i64** %a.1 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* %a.1.i8ptr, i8* %a.i8ptr, i64 8, i1 false)
  call void @llvm.memmove.p0i8.p0i8.i64(i8* %a.1.i8ptr, i8* %a.i8ptr, i64 16, i1 false)
  %a.2 = getelementptr i64*, i64** %a, i64 2
  %v.1 = load i64*, i64** %a.2
  store i64* %v.1, i64** @ext
  ret void
}

define void @runtime.initAll(i8* %context) unnamed_addr {
  call void @testMemSet()
  call void @testMemMove1()
  call void @testMemMove2()
  ret void
}
