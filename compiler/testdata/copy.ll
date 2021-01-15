; ModuleID = 'copy.go'
source_filename = "copy.go"
target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

define internal void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define internal i32 @main.copy64(i64* %dst.data, i32 %dst.len, i32 %dst.cap, i64* %src.data, i32 %src.len, i32 %src.cap, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %copy.dstPtr = bitcast i64* %dst.data to i8*
  %copy.srcPtr = bitcast i64* %src.data to i8*
  %copy.len.cmp = icmp ult i32 %dst.len, %src.len
  %copy.len = select i1 %copy.len.cmp, i32 %dst.len, i32 %src.len
  %copy.len.bytes = shl i32 %copy.len, 3
  call void @llvm.memmove.p0i8.p0i8.i32(i8* align 4 %copy.dstPtr, i8* align 4 %copy.srcPtr, i32 %copy.len.bytes, i1 false)
  ret i32 %copy.len
}

define internal void @main.arrCopy([4 x i32]* dereferenceable_or_null(16) %dst, [4 x i32]* dereferenceable_or_null(16) %src, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %copy.dstPtr = bitcast [4 x i32]* %dst to i8*
  %copy.srcPtr = bitcast [4 x i32]* %src to i8*
  call void @llvm.memmove.p0i8.p0i8.i32(i8* nonnull align 4 dereferenceable(16) %copy.dstPtr, i8* nonnull align 4 dereferenceable(16) %copy.srcPtr, i32 16, i1 false)
  ret void
}

; Function Attrs: argmemonly nounwind willreturn
declare void @llvm.memmove.p0i8.p0i8.i32(i8* nocapture, i8* nocapture readonly, i32, i1 immarg) #0

attributes #0 = { argmemonly nounwind willreturn }
