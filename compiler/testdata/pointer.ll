; ModuleID = 'pointer.go'
source_filename = "pointer.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-wasi"

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden [0 x i32] @main.pointerDerefZero([0 x i32]* %x, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret [0 x i32] zeroinitializer
}

; Function Attrs: nounwind
define hidden i32* @main.pointerCastFromUnsafe(i8* %x, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = bitcast i8* %x to i32*
  ret i32* %0
}

; Function Attrs: nounwind
define hidden i8* @main.pointerCastToUnsafe(i32* dereferenceable_or_null(4) %x, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = bitcast i32* %x to i8*
  ret i8* %0
}

; Function Attrs: nounwind
define hidden i8* @main.pointerCastToUnsafeNoop(i8* dereferenceable_or_null(1) %x, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret i8* %x
}

; Function Attrs: nounwind
define hidden i8* @main.pointerUnsafeGEPFixedOffset(i8* dereferenceable_or_null(1) %ptr, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = getelementptr inbounds i8, i8* %ptr, i32 10
  ret i8* %0
}

; Function Attrs: nounwind
define hidden i8* @main.pointerUnsafeGEPByteOffset(i8* dereferenceable_or_null(1) %ptr, i32 %offset, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = getelementptr inbounds i8, i8* %ptr, i32 %offset
  ret i8* %0
}

; Function Attrs: nounwind
define hidden i32* @main.pointerUnsafeGEPIntOffset(i32* dereferenceable_or_null(4) %ptr, i32 %offset, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = getelementptr i32, i32* %ptr, i32 %offset
  ret i32* %0
}

attributes #0 = { nounwind }
