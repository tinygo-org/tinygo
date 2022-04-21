; ModuleID = 'testdata/alloca.ll'
source_filename = "testdata/alloca.ll"
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@extern = external global i8

define void @runtime.initAll(i8* %context) unnamed_addr {
interpreted:
  %0 = alloca [4 x i8], align 1
  %1 = load i8, i8* @extern, align 1
  %2 = sext i8 %1 to i64
  %3 = bitcast [4 x i8]* %0 to i8*
  call void @llvm.lifetime.start.p0i8(i64 4, i8* %3)
  %4 = bitcast [4 x i8]* %0 to i8*
  store i8 1, i8* %4, align 1
  %5 = bitcast [4 x i8]* %0 to i8*
  %6 = getelementptr i8, i8* %5, i64 1
  store i8 2, i8* %6, align 1
  %7 = bitcast [4 x i8]* %0 to i8*
  %8 = getelementptr i8, i8* %7, i64 2
  store i8 0, i8* %8, align 1
  %9 = bitcast [4 x i8]* %0 to i8*
  %10 = getelementptr i8, i8* %9, i64 3
  store i8 4, i8* %10, align 1
  %11 = bitcast [4 x i8]* %0 to i8*
  %12 = getelementptr i8, i8* %11, i64 %2
  %13 = load i8, i8* %12, align 1
  store i8 %13, i8* @extern, align 1
  call void @llvm.lifetime.end.p0i8(i64 4, i8* %4)
  ret void
}

define i8 @lookup(i8 %x) {
  %tbl = alloca i8, i32 4, align 1
  call void @llvm.memset.p0i8.i64(i8* %tbl, i8 0, i64 4, i1 false)
  store i8 1, i8* %tbl, align 1
  %tbl.1 = getelementptr i8, i8* %tbl, i8 1
  store i8 2, i8* %tbl.1, align 1
  %tbl.3 = getelementptr i8, i8* %tbl, i8 3
  store i8 4, i8* %tbl.3, align 1
  %elem = getelementptr i8, i8* %tbl, i8 %x
  %v = load i8, i8* %elem, align 1
  ret i8 %v
}

; Function Attrs: argmemonly nofree nounwind willreturn writeonly
declare void @llvm.memset.p0i8.i64(i8* nocapture writeonly, i8, i64, i1 immarg) #0

; Function Attrs: argmemonly nofree nosync nounwind willreturn
declare void @llvm.lifetime.start.p0i8(i64 immarg, i8* nocapture) #1

; Function Attrs: argmemonly nofree nosync nounwind willreturn
declare void @llvm.lifetime.end.p0i8(i64 immarg, i8* nocapture) #1

attributes #0 = { argmemonly nofree nounwind willreturn writeonly }
attributes #1 = { argmemonly nofree nosync nounwind willreturn }
