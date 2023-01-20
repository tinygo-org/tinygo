; ModuleID = 'func.go'
source_filename = "func.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.foo(ptr %callback.context, ptr %callback.funcptr, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %callback.funcptr, null
  br i1 %0, label %fpcall.throw, label %fpcall.next

fpcall.next:                                      ; preds = %entry
  call void %callback.funcptr(i32 3, ptr %callback.context) #2
  ret void

fpcall.throw:                                     ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  unreachable
}

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define hidden void @main.bar(ptr %context) unnamed_addr #1 {
entry:
  call void @main.foo(ptr undef, ptr nonnull @main.someFunc, ptr undef)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.someFunc(i32 %arg0, ptr %context) unnamed_addr #1 {
entry:
  ret void
}

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind }
