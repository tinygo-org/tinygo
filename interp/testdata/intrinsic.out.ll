; ModuleID = 'testdata/intrinsic.ll'
source_filename = "testdata/intrinsic.ll"
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@toSet.0 = internal global [10 x i8] c"AAAAAAAAAA"
@set.ret.0 = internal global i8* getelementptr inbounds ([10 x i8], [10 x i8]* @toSet.0, i32 0, i32 0)
@toSet.1 = internal global [3 x i8] c"BBB"
@toSet.2 = internal global [5 x i8] undef
@from = internal constant [94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
@to.1 = internal global [8 x [94 x i8]] [[94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", [94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", [94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", [94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", [94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", [94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", [94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~", [94 x i8] c" !#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"]
@ext = external global i64*

; Function Attrs: argmemonly nofree nounwind willreturn writeonly
declare i8* @memset(i8* nocapture writeonly, i32, i64) #0

; Function Attrs: argmemonly nofree nounwind willreturn writeonly
declare void @llvm.memset.p0i8.i64(i8* nocapture writeonly, i8, i64, i1 immarg) #0

define internal void @testMemSet() {
  %memset.ret = call i8* @memset(i8* getelementptr inbounds ([10 x i8], [10 x i8]* @toSet.0, i32 0, i32 0), i32 321, i64 10)
  store i8* %memset.ret, i8** @set.ret.0, align 8
  call void @llvm.memset.p0i8.i64(i8* getelementptr inbounds ([3 x i8], [3 x i8]* @toSet.1, i32 0, i32 0), i8 66, i64 3, i1 false)
  call void @llvm.memset.p0i8.i64(i8* getelementptr inbounds ([5 x i8], [5 x i8]* @toSet.2, i32 0, i32 0), i8 67, i64 5, i1 true)
  ret void
}

; Function Attrs: argmemonly nofree nounwind willreturn
declare i8* @memmove(i8* nocapture writeonly, i8* nocapture readonly, i64) #1

; Function Attrs: argmemonly nofree nounwind willreturn
declare i8* @memcpy(i8* nocapture writeonly, i8* nocapture readonly, i64) #1

define internal void @testMemMove1() {
entry:
  %cpy.ret = call i8* @memcpy(i8* getelementptr inbounds ([8 x [94 x i8]], [8 x [94 x i8]]* @to.1, i32 0, i32 0, i32 0), i8* getelementptr inbounds ([94 x i8], [94 x i8]* @from, i32 0, i32 0), i64 94)
  %dst.init = getelementptr i8, i8* %cpy.ret, i64 94
  br label %loop

loop:                                             ; preds = %loop, %entry
  %dst = phi i8* [ %dst.init, %entry ], [ %dst.2, %loop ]
  %dst.addr = ptrtoint i8* %dst to i64
  %n = sub i64 %dst.addr, ptrtoint ([8 x [94 x i8]]* @to.1 to i64)
  %dst.1 = call i8* @memmove(i8* %dst, i8* getelementptr inbounds ([8 x [94 x i8]], [8 x [94 x i8]]* @to.1, i32 0, i32 0, i32 0), i64 %n)
  %dst.2 = getelementptr i8, i8* %dst.1, i64 %n
  %ok = icmp ult i64 %n, 376
  br i1 %ok, label %loop, label %done

done:                                             ; preds = %loop
  ret void
}

; Function Attrs: argmemonly nofree nounwind willreturn
declare void @llvm.memcpy.p0i8.p0i8.i64(i8* noalias nocapture writeonly, i8* noalias nocapture readonly, i64, i1 immarg) #1

; Function Attrs: argmemonly nofree nounwind willreturn
declare void @llvm.memmove.p0i8.p0i8.i64(i8* nocapture writeonly, i8* nocapture readonly, i64, i1 immarg) #1

define internal void @testMemMove2() {
  %a = alloca i64*, i32 3, align 8
  %v = load i64*, i64** @ext, align 8
  store i64* %v, i64** %a, align 8
  %a.i8ptr = bitcast i64** %a to i8*
  %a.1 = getelementptr i64*, i64** %a, i64 1
  %a.1.i8ptr = bitcast i64** %a.1 to i8*
  call void @llvm.memcpy.p0i8.p0i8.i64(i8* %a.1.i8ptr, i8* %a.i8ptr, i64 8, i1 false)
  call void @llvm.memmove.p0i8.p0i8.i64(i8* %a.1.i8ptr, i8* %a.i8ptr, i64 16, i1 false)
  %a.2 = getelementptr i64*, i64** %a, i64 2
  %v.1 = load i64*, i64** %a.2, align 8
  store i64* %v.1, i64** @ext, align 8
  ret void
}

define void @runtime.initAll(i8* %context) unnamed_addr {
interpreted:
  call void @llvm.memset.p0i8.i64(i8* getelementptr inbounds ([5 x i8], [5 x i8]* @toSet.2, i32 0, i32 0), i8 67, i64 5, i1 true)
  %0 = load i64*, i64** @ext, align 8
  store i64* %0, i64** @ext, align 8
  ret void
}

attributes #0 = { argmemonly nofree nounwind willreturn writeonly }
attributes #1 = { argmemonly nofree nounwind willreturn }
