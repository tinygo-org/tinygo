; ModuleID = 'go1.17.go'
source_filename = "go1.17.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-wasi"

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i8* @main.Add32(i8* %p, i32 %len, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = getelementptr i8, i8* %p, i32 %len
  ret i8* %0
}

; Function Attrs: nounwind
define hidden i8* @main.Add64(i8* %p, i64 %len, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = trunc i64 %len to i32
  %1 = getelementptr i8, i8* %p, i32 %0
  ret i8* %1
}

; Function Attrs: nounwind
define hidden [4 x i32]* @main.SliceToArray(i32* %s.data, i32 %s.len, i32 %s.cap, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = icmp ult i32 %s.len, 4
  br i1 %0, label %slicetoarray.throw, label %slicetoarray.next

slicetoarray.throw:                               ; preds = %entry
  call void @runtime.sliceToArrayPointerPanic(i8* undef, i8* null) #0
  unreachable

slicetoarray.next:                                ; preds = %entry
  %1 = bitcast i32* %s.data to [4 x i32]*
  ret [4 x i32]* %1
}

declare void @runtime.sliceToArrayPointerPanic(i8*, i8*)

; Function Attrs: nounwind
define hidden [4 x i32]* @main.SliceToArrayConst(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %makeslice = call i8* @runtime.alloc(i32 24, i8* null, i8* undef, i8* null) #0
  br i1 false, label %slicetoarray.throw, label %slicetoarray.next

slicetoarray.throw:                               ; preds = %entry
  unreachable

slicetoarray.next:                                ; preds = %entry
  %0 = bitcast i8* %makeslice to [4 x i32]*
  ret [4 x i32]* %0
}

; Function Attrs: nounwind
define hidden { i32*, i32, i32 } @main.SliceInt(i32* dereferenceable_or_null(4) %ptr, i32 %len, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = icmp ugt i32 %len, 1073741823
  %1 = icmp eq i32* %ptr, null
  %2 = icmp ne i32 %len, 0
  %3 = and i1 %1, %2
  %4 = or i1 %3, %0
  br i1 %4, label %unsafe.Slice.throw, label %unsafe.Slice.next

unsafe.Slice.throw:                               ; preds = %entry
  call void @runtime.unsafeSlicePanic(i8* undef, i8* null) #0
  unreachable

unsafe.Slice.next:                                ; preds = %entry
  %5 = insertvalue { i32*, i32, i32 } undef, i32* %ptr, 0
  %6 = insertvalue { i32*, i32, i32 } %5, i32 %len, 1
  %7 = insertvalue { i32*, i32, i32 } %6, i32 %len, 2
  ret { i32*, i32, i32 } %7
}

declare void @runtime.unsafeSlicePanic(i8*, i8*)

; Function Attrs: nounwind
define hidden { i8*, i32, i32 } @main.SliceUint16(i8* dereferenceable_or_null(1) %ptr, i16 %len, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = icmp eq i8* %ptr, null
  %1 = icmp ne i16 %len, 0
  %2 = and i1 %0, %1
  br i1 %2, label %unsafe.Slice.throw, label %unsafe.Slice.next

unsafe.Slice.throw:                               ; preds = %entry
  call void @runtime.unsafeSlicePanic(i8* undef, i8* null) #0
  unreachable

unsafe.Slice.next:                                ; preds = %entry
  %3 = zext i16 %len to i32
  %4 = insertvalue { i8*, i32, i32 } undef, i8* %ptr, 0
  %5 = insertvalue { i8*, i32, i32 } %4, i32 %3, 1
  %6 = insertvalue { i8*, i32, i32 } %5, i32 %3, 2
  ret { i8*, i32, i32 } %6
}

; Function Attrs: nounwind
define hidden { i32*, i32, i32 } @main.SliceUint64(i32* dereferenceable_or_null(4) %ptr, i64 %len, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = icmp ugt i64 %len, 1073741823
  %1 = icmp eq i32* %ptr, null
  %2 = icmp ne i64 %len, 0
  %3 = and i1 %1, %2
  %4 = or i1 %3, %0
  br i1 %4, label %unsafe.Slice.throw, label %unsafe.Slice.next

unsafe.Slice.throw:                               ; preds = %entry
  call void @runtime.unsafeSlicePanic(i8* undef, i8* null) #0
  unreachable

unsafe.Slice.next:                                ; preds = %entry
  %5 = trunc i64 %len to i32
  %6 = insertvalue { i32*, i32, i32 } undef, i32* %ptr, 0
  %7 = insertvalue { i32*, i32, i32 } %6, i32 %5, 1
  %8 = insertvalue { i32*, i32, i32 } %7, i32 %5, 2
  ret { i32*, i32, i32 } %8
}

; Function Attrs: nounwind
define hidden { i32*, i32, i32 } @main.SliceInt64(i32* dereferenceable_or_null(4) %ptr, i64 %len, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = icmp ugt i64 %len, 1073741823
  %1 = icmp eq i32* %ptr, null
  %2 = icmp ne i64 %len, 0
  %3 = and i1 %1, %2
  %4 = or i1 %3, %0
  br i1 %4, label %unsafe.Slice.throw, label %unsafe.Slice.next

unsafe.Slice.throw:                               ; preds = %entry
  call void @runtime.unsafeSlicePanic(i8* undef, i8* null) #0
  unreachable

unsafe.Slice.next:                                ; preds = %entry
  %5 = trunc i64 %len to i32
  %6 = insertvalue { i32*, i32, i32 } undef, i32* %ptr, 0
  %7 = insertvalue { i32*, i32, i32 } %6, i32 %5, 1
  %8 = insertvalue { i32*, i32, i32 } %7, i32 %5, 2
  ret { i32*, i32, i32 } %8
}

attributes #0 = { nounwind }
