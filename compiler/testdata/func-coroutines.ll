; ModuleID = 'func.go'
source_filename = "func.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.funcValueWithSignature = type { i32, i8* }

@"reflect/types.funcid:func:{basic:int}{}" = external constant i8
@"main.someFunc$withSignature" = linkonce_odr constant %runtime.funcValueWithSignature { i32 ptrtoint (void (i32, i8*, i8*)* @main.someFunc to i32), i8* @"reflect/types.funcid:func:{basic:int}{}" }

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*, i8*)

declare void @runtime.trackPointer(i8* nocapture readonly, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.foo(i8* %callback.context, i32 %callback.funcptr, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = call i32 @runtime.getFuncPtr(i8* %callback.context, i32 %callback.funcptr, i8* nonnull @"reflect/types.funcid:func:{basic:int}{}", i8* undef, i8* null) #0
  %1 = icmp eq i32 %0, 0
  br i1 %1, label %fpcall.throw, label %fpcall.next

fpcall.throw:                                     ; preds = %entry
  call void @runtime.nilPanic(i8* undef, i8* null) #0
  unreachable

fpcall.next:                                      ; preds = %entry
  %2 = inttoptr i32 %0 to void (i32, i8*, i8*)*
  call void %2(i32 3, i8* %callback.context, i8* undef) #0
  ret void
}

declare i32 @runtime.getFuncPtr(i8*, i32, i8* dereferenceable_or_null(1), i8*, i8*)

declare void @runtime.nilPanic(i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.bar(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  call void @main.foo(i8* undef, i32 ptrtoint (%runtime.funcValueWithSignature* @"main.someFunc$withSignature" to i32), i8* undef, i8* undef)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.someFunc(i32 %arg0, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

attributes #0 = { nounwind }
