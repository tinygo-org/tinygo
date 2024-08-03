; ModuleID = 'go1.20.go'
source_filename = "go1.20.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

; Function Attrs: nounwind
declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden ptr @main.unsafeSliceData(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s.data, ptr nonnull %stackalloc, ptr undef)
  ret ptr %s.data
}

; Function Attrs: nounwind
define hidden %runtime._string @main.unsafeString(ptr dereferenceable_or_null(1) %ptr, i16 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp slt i16 %len, 0
  %1 = icmp eq ptr %ptr, null
  %2 = icmp ne i16 %len, 0
  %3 = and i1 %1, %2
  %4 = or i1 %3, %0
  br i1 %4, label %unsafe.String.throw, label %unsafe.String.next

unsafe.String.next:                               ; preds = %entry
  %5 = zext i16 %len to i32
  %6 = insertvalue %runtime._string undef, ptr %ptr, 0
  %7 = insertvalue %runtime._string %6, i32 %5, 1
  call void @runtime.trackPointer(ptr %ptr, ptr nonnull %stackalloc, ptr undef)
  ret %runtime._string %7

unsafe.String.throw:                              ; preds = %entry
  call void @runtime.unsafeSlicePanic(ptr undef) #3
  unreachable
}

declare void @runtime.unsafeSlicePanic(ptr) #2

; Function Attrs: nounwind
define hidden ptr @main.unsafeStringData(ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s.data, ptr nonnull %stackalloc, ptr undef)
  ret ptr %s.data
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+exception-handling,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+exception-handling,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { "target-features"="+bulk-memory,+exception-handling,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #3 = { nounwind }
