; ModuleID = 'goroutine.go'
source_filename = "goroutine.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }

@"main$string" = internal unnamed_addr constant [4 x i8] c"test", align 1

declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.regularFunctionGoroutine(ptr %context) unnamed_addr #1 {
entry:
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.regularFunction$gowrapper" to i32), ptr nonnull inttoptr (i32 5 to ptr), i32 16384, ptr undef) #8
  ret void
}

declare void @main.regularFunction(i32, ptr) #0

declare void @runtime.deadlock(ptr) #0

; Function Attrs: nounwind
define linkonce_odr void @"main.regularFunction$gowrapper"(ptr %0) unnamed_addr #2 {
entry:
  %unpack.int = ptrtoint ptr %0 to i32
  call void @main.regularFunction(i32 %unpack.int, ptr undef) #8
  call void @runtime.deadlock(ptr undef) #8
  unreachable
}

declare void @"internal/task.start"(i32, ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden void @main.inlineFunctionGoroutine(ptr %context) unnamed_addr #1 {
entry:
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.inlineFunctionGoroutine$1$gowrapper" to i32), ptr nonnull inttoptr (i32 5 to ptr), i32 16384, ptr undef) #8
  ret void
}

; Function Attrs: nounwind
define internal void @"main.inlineFunctionGoroutine$1"(i32 %x, ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define linkonce_odr void @"main.inlineFunctionGoroutine$1$gowrapper"(ptr %0) unnamed_addr #3 {
entry:
  %unpack.int = ptrtoint ptr %0 to i32
  call void @"main.inlineFunctionGoroutine$1"(i32 %unpack.int, ptr undef)
  call void @runtime.deadlock(ptr undef) #8
  unreachable
}

; Function Attrs: nounwind
define hidden void @main.closureFunctionGoroutine(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %n = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull %n, ptr nonnull %stackalloc, ptr undef) #8
  store i32 3, ptr %n, align 4
  call void @runtime.trackPointer(ptr nonnull %n, ptr nonnull %stackalloc, ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull @"main.closureFunctionGoroutine$1", ptr nonnull %stackalloc, ptr undef) #8
  %0 = call ptr @runtime.alloc(i32 8, ptr null, ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #8
  store i32 5, ptr %0, align 4
  %1 = getelementptr inbounds { i32, ptr }, ptr %0, i32 0, i32 1
  store ptr %n, ptr %1, align 4
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.closureFunctionGoroutine$1$gowrapper" to i32), ptr nonnull %0, i32 16384, ptr undef) #8
  %2 = load i32, ptr %n, align 4
  call void @runtime.printint32(i32 %2, ptr undef) #8
  ret void
}

; Function Attrs: nounwind
define internal void @"main.closureFunctionGoroutine$1"(i32 %x, ptr %context) unnamed_addr #1 {
entry:
  store i32 7, ptr %context, align 4
  ret void
}

; Function Attrs: nounwind
define linkonce_odr void @"main.closureFunctionGoroutine$1$gowrapper"(ptr %0) unnamed_addr #4 {
entry:
  %1 = load i32, ptr %0, align 4
  %2 = getelementptr inbounds { i32, ptr }, ptr %0, i32 0, i32 1
  %3 = load ptr, ptr %2, align 4
  call void @"main.closureFunctionGoroutine$1"(i32 %1, ptr %3)
  call void @runtime.deadlock(ptr undef) #8
  unreachable
}

declare void @runtime.printint32(i32, ptr) #0

; Function Attrs: nounwind
define hidden void @main.funcGoroutine(ptr %fn.context, ptr %fn.funcptr, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.alloc(i32 12, ptr null, ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #8
  store i32 5, ptr %0, align 4
  %1 = getelementptr inbounds { i32, ptr, ptr }, ptr %0, i32 0, i32 1
  store ptr %fn.context, ptr %1, align 4
  %2 = getelementptr inbounds { i32, ptr, ptr }, ptr %0, i32 0, i32 2
  store ptr %fn.funcptr, ptr %2, align 4
  call void @"internal/task.start"(i32 ptrtoint (ptr @main.funcGoroutine.gowrapper to i32), ptr nonnull %0, i32 16384, ptr undef) #8
  ret void
}

; Function Attrs: nounwind
define linkonce_odr void @main.funcGoroutine.gowrapper(ptr %0) unnamed_addr #5 {
entry:
  %1 = load i32, ptr %0, align 4
  %2 = getelementptr inbounds { i32, ptr, ptr }, ptr %0, i32 0, i32 1
  %3 = load ptr, ptr %2, align 4
  %4 = getelementptr inbounds { i32, ptr, ptr }, ptr %0, i32 0, i32 2
  %5 = load ptr, ptr %4, align 4
  call void %5(i32 %1, ptr %3) #8
  call void @runtime.deadlock(ptr undef) #8
  unreachable
}

; Function Attrs: nounwind
define hidden void @main.recoverBuiltinGoroutine(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.copyBuiltinGoroutine(ptr %dst.data, i32 %dst.len, i32 %dst.cap, ptr %src.data, i32 %src.len, i32 %src.cap, ptr %context) unnamed_addr #1 {
entry:
  %copy.n = call i32 @runtime.sliceCopy(ptr %dst.data, ptr %src.data, i32 %dst.len, i32 %src.len, i32 1, ptr undef) #8
  ret void
}

declare i32 @runtime.sliceCopy(ptr nocapture writeonly, ptr nocapture readonly, i32, i32, i32, ptr) #0

; Function Attrs: nounwind
define hidden void @main.closeBuiltinGoroutine(ptr dereferenceable_or_null(32) %ch, ptr %context) unnamed_addr #1 {
entry:
  call void @runtime.chanClose(ptr %ch, ptr undef) #8
  ret void
}

declare void @runtime.chanClose(ptr dereferenceable_or_null(32), ptr) #0

; Function Attrs: nounwind
define hidden void @main.startInterfaceMethod(i32 %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.alloc(i32 16, ptr null, ptr undef) #8
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #8
  store ptr %itf.value, ptr %0, align 4
  %1 = getelementptr inbounds { ptr, %runtime._string, i32 }, ptr %0, i32 0, i32 1
  store ptr @"main$string", ptr %1, align 4
  %.repack1 = getelementptr inbounds { ptr, %runtime._string, i32 }, ptr %0, i32 0, i32 1, i32 1
  store i32 4, ptr %.repack1, align 4
  %2 = getelementptr inbounds { ptr, %runtime._string, i32 }, ptr %0, i32 0, i32 2
  store i32 %itf.typecode, ptr %2, align 4
  call void @"internal/task.start"(i32 ptrtoint (ptr @"interface:{Print:func:{basic:string}{}}.Print$invoke$gowrapper" to i32), ptr nonnull %0, i32 16384, ptr undef) #8
  ret void
}

declare void @"interface:{Print:func:{basic:string}{}}.Print$invoke"(ptr, ptr, i32, i32, ptr) #6

; Function Attrs: nounwind
define linkonce_odr void @"interface:{Print:func:{basic:string}{}}.Print$invoke$gowrapper"(ptr %0) unnamed_addr #7 {
entry:
  %1 = load ptr, ptr %0, align 4
  %2 = getelementptr inbounds { ptr, ptr, i32, i32 }, ptr %0, i32 0, i32 1
  %3 = load ptr, ptr %2, align 4
  %4 = getelementptr inbounds { ptr, ptr, i32, i32 }, ptr %0, i32 0, i32 2
  %5 = load i32, ptr %4, align 4
  %6 = getelementptr inbounds { ptr, ptr, i32, i32 }, ptr %0, i32 0, i32 3
  %7 = load i32, ptr %6, align 4
  call void @"interface:{Print:func:{basic:string}{}}.Print$invoke"(ptr %1, ptr %3, i32 %5, i32 %7, ptr undef) #8
  call void @runtime.deadlock(ptr undef) #8
  unreachable
}

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-gowrapper"="main.regularFunction" }
attributes #3 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-gowrapper"="main.inlineFunctionGoroutine$1" }
attributes #4 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-gowrapper"="main.closureFunctionGoroutine$1" }
attributes #5 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-gowrapper" }
attributes #6 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-invoke"="reflect/methods.Print(string)" "tinygo-methods"="reflect/methods.Print(string)" }
attributes #7 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-gowrapper"="interface:{Print:func:{basic:string}{}}.Print$invoke" }
attributes #8 = { nounwind }
