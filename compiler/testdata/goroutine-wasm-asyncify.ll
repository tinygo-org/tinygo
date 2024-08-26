; ModuleID = 'goroutine.go'
source_filename = "goroutine.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

@"main$string" = internal unnamed_addr constant [4 x i8] c"test", align 1

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.regularFunctionGoroutine(ptr %context) unnamed_addr #2 {
entry:
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.regularFunction$gowrapper" to i32), ptr nonnull inttoptr (i32 5 to ptr), i32 65536, ptr undef) #9
  ret void
}

declare void @main.regularFunction(i32, ptr) #1

declare void @runtime.deadlock(ptr) #1

; Function Attrs: nounwind
define linkonce_odr void @"main.regularFunction$gowrapper"(ptr %0) unnamed_addr #3 {
entry:
  %unpack.int = ptrtoint ptr %0 to i32
  call void @main.regularFunction(i32 %unpack.int, ptr undef) #9
  call void @runtime.deadlock(ptr undef) #9
  unreachable
}

declare void @"internal/task.start"(i32, ptr, i32, ptr) #1

; Function Attrs: nounwind
define hidden void @main.inlineFunctionGoroutine(ptr %context) unnamed_addr #2 {
entry:
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.inlineFunctionGoroutine$1$gowrapper" to i32), ptr nonnull inttoptr (i32 5 to ptr), i32 65536, ptr undef) #9
  ret void
}

; Function Attrs: nounwind
define internal void @"main.inlineFunctionGoroutine$1"(i32 %x, ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define linkonce_odr void @"main.inlineFunctionGoroutine$1$gowrapper"(ptr %0) unnamed_addr #4 {
entry:
  %unpack.int = ptrtoint ptr %0 to i32
  call void @"main.inlineFunctionGoroutine$1"(i32 %unpack.int, ptr undef)
  call void @runtime.deadlock(ptr undef) #9
  unreachable
}

; Function Attrs: nounwind
define hidden void @main.closureFunctionGoroutine(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %n = call align 4 dereferenceable(4) ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %n, ptr nonnull %stackalloc, ptr undef) #9
  store i32 3, ptr %n, align 4
  call void @runtime.trackPointer(ptr nonnull %n, ptr nonnull %stackalloc, ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull @"main.closureFunctionGoroutine$1", ptr nonnull %stackalloc, ptr undef) #9
  %0 = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr null, ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #9
  store i32 5, ptr %0, align 4
  %1 = getelementptr inbounds i8, ptr %0, i32 4
  store ptr %n, ptr %1, align 4
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.closureFunctionGoroutine$1$gowrapper" to i32), ptr nonnull %0, i32 65536, ptr undef) #9
  %2 = load i32, ptr %n, align 4
  call void @runtime.printint32(i32 %2, ptr undef) #9
  ret void
}

; Function Attrs: nounwind
define internal void @"main.closureFunctionGoroutine$1"(i32 %x, ptr %context) unnamed_addr #2 {
entry:
  store i32 7, ptr %context, align 4
  ret void
}

; Function Attrs: nounwind
define linkonce_odr void @"main.closureFunctionGoroutine$1$gowrapper"(ptr %0) unnamed_addr #5 {
entry:
  %1 = load i32, ptr %0, align 4
  %2 = getelementptr inbounds i8, ptr %0, i32 4
  %3 = load ptr, ptr %2, align 4
  call void @"main.closureFunctionGoroutine$1"(i32 %1, ptr %3)
  call void @runtime.deadlock(ptr undef) #9
  unreachable
}

declare void @runtime.printint32(i32, ptr) #1

; Function Attrs: nounwind
define hidden void @main.funcGoroutine(ptr %fn.context, ptr %fn.funcptr, ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call align 4 dereferenceable(12) ptr @runtime.alloc(i32 12, ptr null, ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #9
  store i32 5, ptr %0, align 4
  %1 = getelementptr inbounds i8, ptr %0, i32 4
  store ptr %fn.context, ptr %1, align 4
  %2 = getelementptr inbounds i8, ptr %0, i32 8
  store ptr %fn.funcptr, ptr %2, align 4
  call void @"internal/task.start"(i32 ptrtoint (ptr @main.funcGoroutine.gowrapper to i32), ptr nonnull %0, i32 65536, ptr undef) #9
  ret void
}

; Function Attrs: nounwind
define linkonce_odr void @main.funcGoroutine.gowrapper(ptr %0) unnamed_addr #6 {
entry:
  %1 = load i32, ptr %0, align 4
  %2 = getelementptr inbounds i8, ptr %0, i32 4
  %3 = load ptr, ptr %2, align 4
  %4 = getelementptr inbounds i8, ptr %0, i32 8
  %5 = load ptr, ptr %4, align 4
  call void %5(i32 %1, ptr %3) #9
  call void @runtime.deadlock(ptr undef) #9
  unreachable
}

; Function Attrs: nounwind
define hidden void @main.recoverBuiltinGoroutine(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.copyBuiltinGoroutine(ptr %dst.data, i32 %dst.len, i32 %dst.cap, ptr %src.data, i32 %src.len, i32 %src.cap, ptr %context) unnamed_addr #2 {
entry:
  %copy.n = call i32 @runtime.sliceCopy(ptr %dst.data, ptr %src.data, i32 %dst.len, i32 %src.len, i32 1, ptr undef) #9
  ret void
}

declare i32 @runtime.sliceCopy(ptr nocapture writeonly, ptr nocapture readonly, i32, i32, i32, ptr) #1

; Function Attrs: nounwind
define hidden void @main.closeBuiltinGoroutine(ptr dereferenceable_or_null(32) %ch, ptr %context) unnamed_addr #2 {
entry:
  call void @runtime.chanClose(ptr %ch, ptr undef) #9
  ret void
}

declare void @runtime.chanClose(ptr dereferenceable_or_null(32), ptr) #1

; Function Attrs: nounwind
define hidden void @main.startInterfaceMethod(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call align 4 dereferenceable(16) ptr @runtime.alloc(i32 16, ptr null, ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #9
  store ptr %itf.value, ptr %0, align 4
  %1 = getelementptr inbounds i8, ptr %0, i32 4
  store ptr @"main$string", ptr %1, align 4
  %.repack1 = getelementptr inbounds i8, ptr %0, i32 8
  store i32 4, ptr %.repack1, align 4
  %2 = getelementptr inbounds i8, ptr %0, i32 12
  store ptr %itf.typecode, ptr %2, align 4
  call void @"internal/task.start"(i32 ptrtoint (ptr @"interface:{Print:func:{basic:string}{}}.Print$invoke$gowrapper" to i32), ptr nonnull %0, i32 65536, ptr undef) #9
  ret void
}

declare void @"interface:{Print:func:{basic:string}{}}.Print$invoke"(ptr, ptr, i32, ptr, ptr) #7

; Function Attrs: nounwind
define linkonce_odr void @"interface:{Print:func:{basic:string}{}}.Print$invoke$gowrapper"(ptr %0) unnamed_addr #8 {
entry:
  %1 = load ptr, ptr %0, align 4
  %2 = getelementptr inbounds i8, ptr %0, i32 4
  %3 = load ptr, ptr %2, align 4
  %4 = getelementptr inbounds i8, ptr %0, i32 8
  %5 = load i32, ptr %4, align 4
  %6 = getelementptr inbounds i8, ptr %0, i32 12
  %7 = load ptr, ptr %6, align 4
  call void @"interface:{Print:func:{basic:string}{}}.Print$invoke"(ptr %1, ptr %3, i32 %5, ptr %7, ptr undef) #9
  call void @runtime.deadlock(ptr undef) #9
  unreachable
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-gowrapper"="main.regularFunction" }
attributes #4 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-gowrapper"="main.inlineFunctionGoroutine$1" }
attributes #5 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-gowrapper"="main.closureFunctionGoroutine$1" }
attributes #6 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-gowrapper" }
attributes #7 = { "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-invoke"="reflect/methods.Print(string)" "tinygo-methods"="reflect/methods.Print(string)" }
attributes #8 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-gowrapper"="interface:{Print:func:{basic:string}{}}.Print$invoke" }
attributes #9 = { nounwind }
