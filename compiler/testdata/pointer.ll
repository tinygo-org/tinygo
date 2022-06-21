; ModuleID = 'pointer.go'
source_filename = "pointer.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*) #0

declare void @runtime.trackPointer(i8* nocapture readonly, i8*) #0

; Function Attrs: nounwind
define hidden void @main.init(i8* %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden [0 x i32] @main.pointerDerefZero([0 x i32]* %x, i8* %context) unnamed_addr #1 {
entry:
  ret [0 x i32] zeroinitializer
}

; Function Attrs: nounwind
define hidden i32* @main.pointerCastFromUnsafe(i8* %x, i8* %context) unnamed_addr #1 {
entry:
  %0 = bitcast i8* %x to i32*
  call void @runtime.trackPointer(i8* %x, i8* undef) #2
  ret i32* %0
}

; Function Attrs: nounwind
define hidden i8* @main.pointerCastToUnsafe(i32* dereferenceable_or_null(4) %x, i8* %context) unnamed_addr #1 {
entry:
  %0 = bitcast i32* %x to i8*
  call void @runtime.trackPointer(i8* %0, i8* undef) #2
  ret i8* %0
}

; Function Attrs: nounwind
define hidden i8* @main.pointerCastToUnsafeNoop(i8* dereferenceable_or_null(1) %x, i8* %context) unnamed_addr #1 {
entry:
  call void @runtime.trackPointer(i8* %x, i8* undef) #2
  ret i8* %x
}

; Function Attrs: nounwind
define hidden i8* @main.pointerUnsafeGEPFixedOffset(i8* dereferenceable_or_null(1) %ptr, i8* %context) unnamed_addr #1 {
entry:
  call void @runtime.trackPointer(i8* %ptr, i8* undef) #2
  %0 = getelementptr inbounds i8, i8* %ptr, i32 10
  call void @runtime.trackPointer(i8* nonnull %0, i8* undef) #2
  call void @runtime.trackPointer(i8* nonnull %0, i8* undef) #2
  ret i8* %0
}

; Function Attrs: nounwind
define hidden i8* @main.pointerUnsafeGEPByteOffset(i8* dereferenceable_or_null(1) %ptr, i32 %offset, i8* %context) unnamed_addr #1 {
entry:
  call void @runtime.trackPointer(i8* %ptr, i8* undef) #2
  %0 = getelementptr inbounds i8, i8* %ptr, i32 %offset
  call void @runtime.trackPointer(i8* %0, i8* undef) #2
  call void @runtime.trackPointer(i8* %0, i8* undef) #2
  ret i8* %0
}

; Function Attrs: nounwind
define hidden i32* @main.pointerUnsafeGEPIntOffset(i32* dereferenceable_or_null(4) %ptr, i32 %offset, i8* %context) unnamed_addr #1 {
entry:
  %0 = bitcast i32* %ptr to i8*
  call void @runtime.trackPointer(i8* %0, i8* undef) #2
  %1 = getelementptr i32, i32* %ptr, i32 %offset
  %2 = bitcast i32* %1 to i8*
  call void @runtime.trackPointer(i8* %2, i8* undef) #2
  %3 = bitcast i32* %1 to i8*
  call void @runtime.trackPointer(i8* %3, i8* undef) #2
  ret i32* %1
}

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind }
