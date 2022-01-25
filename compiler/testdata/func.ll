; ModuleID = 'func.go'
source_filename = "func.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

declare void @runtime.trackPointer(i8* nocapture readonly, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.foo(i8* %callback.context, void ()* %callback.funcptr, i8* %context) unnamed_addr #0 {
entry:
  %0 = icmp eq void ()* %callback.funcptr, null
  br i1 %0, label %fpcall.throw, label %fpcall.next

fpcall.throw:                                     ; preds = %entry
  call void @runtime.nilPanic(i8* undef) #0
  unreachable

fpcall.next:                                      ; preds = %entry
  %1 = bitcast void ()* %callback.funcptr to void (i32, i8*)*
  call void %1(i32 3, i8* %callback.context) #0
  ret void
}

declare void @runtime.nilPanic(i8*)

; Function Attrs: nounwind
define hidden void @main.bar(i8* %context) unnamed_addr #0 {
entry:
  call void @main.foo(i8* undef, void ()* bitcast (void (i32, i8*)* @main.someFunc to void ()*), i8* undef)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.someFunc(i32 %arg0, i8* %context) unnamed_addr #0 {
entry:
  ret void
}

attributes #0 = { nounwind }
