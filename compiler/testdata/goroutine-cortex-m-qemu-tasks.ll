; ModuleID = 'goroutine.go'
source_filename = "goroutine.go"
target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "thumbv7m-unknown-unknown-eabi"

@"main$string" = internal unnamed_addr constant [4 x i8] c"test", align 1

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.regularFunctionGoroutine(ptr %context) unnamed_addr #1 {
entry:
  %stacksize = call i32 @"internal/task.getGoroutineStackSize"(i32 ptrtoint (ptr @"main.regularFunction$gowrapper" to i32), ptr undef) #9
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.regularFunction$gowrapper" to i32), ptr nonnull inttoptr (i32 5 to ptr), i32 %stacksize, ptr undef) #9
  ret void
}

declare void @main.regularFunction(i32, ptr) #2

; Function Attrs: nounwind
define linkonce_odr void @"main.regularFunction$gowrapper"(ptr %0) unnamed_addr #3 {
entry:
  %unpack.int = ptrtoint ptr %0 to i32
  call void @main.regularFunction(i32 %unpack.int, ptr undef) #9
  ret void
}

declare i32 @"internal/task.getGoroutineStackSize"(i32, ptr) #2

declare void @"internal/task.start"(i32, ptr, i32, ptr) #2

; Function Attrs: nounwind
define hidden void @main.inlineFunctionGoroutine(ptr %context) unnamed_addr #1 {
entry:
  %stacksize = call i32 @"internal/task.getGoroutineStackSize"(i32 ptrtoint (ptr @"main.inlineFunctionGoroutine$1$gowrapper" to i32), ptr undef) #9
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.inlineFunctionGoroutine$1$gowrapper" to i32), ptr nonnull inttoptr (i32 5 to ptr), i32 %stacksize, ptr undef) #9
  ret void
}

; Function Attrs: nounwind
define internal void @"main.inlineFunctionGoroutine$1"(i32 %x, ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define linkonce_odr void @"main.inlineFunctionGoroutine$1$gowrapper"(ptr %0) unnamed_addr #4 {
entry:
  %unpack.int = ptrtoint ptr %0 to i32
  call void @"main.inlineFunctionGoroutine$1"(i32 %unpack.int, ptr undef)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.closureFunctionGoroutine(ptr %context) unnamed_addr #1 {
entry:
  %n = call align 4 dereferenceable(4) ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  store i32 3, ptr %n, align 4
  %0 = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr null, ptr undef) #9
  store i32 5, ptr %0, align 4
  %1 = getelementptr inbounds { i32, ptr }, ptr %0, i32 0, i32 1
  store ptr %n, ptr %1, align 4
  %stacksize = call i32 @"internal/task.getGoroutineStackSize"(i32 ptrtoint (ptr @"main.closureFunctionGoroutine$1$gowrapper" to i32), ptr undef) #9
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.closureFunctionGoroutine$1$gowrapper" to i32), ptr nonnull %0, i32 %stacksize, ptr undef) #9
  %2 = load i32, ptr %n, align 4
  call void @runtime.printint32(i32 %2, ptr undef) #9
  ret void
}

; Function Attrs: nounwind
define internal void @"main.closureFunctionGoroutine$1"(i32 %x, ptr %context) unnamed_addr #1 {
entry:
  store i32 7, ptr %context, align 4
  ret void
}

; Function Attrs: nounwind
define linkonce_odr void @"main.closureFunctionGoroutine$1$gowrapper"(ptr %0) unnamed_addr #5 {
entry:
  %1 = load i32, ptr %0, align 4
  %2 = getelementptr inbounds { i32, ptr }, ptr %0, i32 0, i32 1
  %3 = load ptr, ptr %2, align 4
  call void @"main.closureFunctionGoroutine$1"(i32 %1, ptr %3)
  ret void
}

declare void @runtime.printint32(i32, ptr) #2

; Function Attrs: nounwind
define hidden void @main.funcGoroutine(ptr %fn.context, ptr %fn.funcptr, ptr %context) unnamed_addr #1 {
entry:
  %0 = call align 4 dereferenceable(12) ptr @runtime.alloc(i32 12, ptr null, ptr undef) #9
  store i32 5, ptr %0, align 4
  %1 = getelementptr inbounds { i32, ptr, ptr }, ptr %0, i32 0, i32 1
  store ptr %fn.context, ptr %1, align 4
  %2 = getelementptr inbounds { i32, ptr, ptr }, ptr %0, i32 0, i32 2
  store ptr %fn.funcptr, ptr %2, align 4
  %stacksize = call i32 @"internal/task.getGoroutineStackSize"(i32 ptrtoint (ptr @main.funcGoroutine.gowrapper to i32), ptr undef) #9
  call void @"internal/task.start"(i32 ptrtoint (ptr @main.funcGoroutine.gowrapper to i32), ptr nonnull %0, i32 %stacksize, ptr undef) #9
  ret void
}

; Function Attrs: nounwind
define linkonce_odr void @main.funcGoroutine.gowrapper(ptr %0) unnamed_addr #6 {
entry:
  %1 = load i32, ptr %0, align 4
  %2 = getelementptr inbounds { i32, ptr, ptr }, ptr %0, i32 0, i32 1
  %3 = load ptr, ptr %2, align 4
  %4 = getelementptr inbounds { i32, ptr, ptr }, ptr %0, i32 0, i32 2
  %5 = load ptr, ptr %4, align 4
  call void %5(i32 %1, ptr %3) #9
  ret void
}

; Function Attrs: nounwind
define hidden void @main.recoverBuiltinGoroutine(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.copyBuiltinGoroutine(ptr %dst.data, i32 %dst.len, i32 %dst.cap, ptr %src.data, i32 %src.len, i32 %src.cap, ptr %context) unnamed_addr #1 {
entry:
  %copy.n = call i32 @runtime.sliceCopy(ptr %dst.data, ptr %src.data, i32 %dst.len, i32 %src.len, i32 1, ptr undef) #9
  ret void
}

declare i32 @runtime.sliceCopy(ptr nocapture writeonly, ptr nocapture readonly, i32, i32, i32, ptr) #2

; Function Attrs: nounwind
define hidden void @main.closeBuiltinGoroutine(ptr dereferenceable_or_null(32) %ch, ptr %context) unnamed_addr #1 {
entry:
  call void @runtime.chanClose(ptr %ch, ptr undef) #9
  ret void
}

declare void @runtime.chanClose(ptr dereferenceable_or_null(32), ptr) #2

; Function Attrs: nounwind
define hidden void @main.startInterfaceMethod(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %0 = call align 4 dereferenceable(16) ptr @runtime.alloc(i32 16, ptr null, ptr undef) #9
  store ptr %itf.value, ptr %0, align 4
  %1 = getelementptr inbounds { ptr, ptr, i32, ptr }, ptr %0, i32 0, i32 1
  store ptr @"main$string", ptr %1, align 4
  %2 = getelementptr inbounds { ptr, ptr, i32, ptr }, ptr %0, i32 0, i32 2
  store i32 4, ptr %2, align 4
  %3 = getelementptr inbounds { ptr, ptr, i32, ptr }, ptr %0, i32 0, i32 3
  store ptr %itf.typecode, ptr %3, align 4
  %stacksize = call i32 @"internal/task.getGoroutineStackSize"(i32 ptrtoint (ptr @"interface:{Print:func:{basic:string}{}}.Print$invoke$gowrapper" to i32), ptr undef) #9
  call void @"internal/task.start"(i32 ptrtoint (ptr @"interface:{Print:func:{basic:string}{}}.Print$invoke$gowrapper" to i32), ptr nonnull %0, i32 %stacksize, ptr undef) #9
  ret void
}

declare void @"interface:{Print:func:{basic:string}{}}.Print$invoke"(ptr, ptr, i32, ptr, ptr) #7

; Function Attrs: nounwind
define linkonce_odr void @"interface:{Print:func:{basic:string}{}}.Print$invoke$gowrapper"(ptr %0) unnamed_addr #8 {
entry:
  %1 = load ptr, ptr %0, align 4
  %2 = getelementptr inbounds { ptr, ptr, i32, ptr }, ptr %0, i32 0, i32 1
  %3 = load ptr, ptr %2, align 4
  %4 = getelementptr inbounds { ptr, ptr, i32, ptr }, ptr %0, i32 0, i32 2
  %5 = load i32, ptr %4, align 4
  %6 = getelementptr inbounds { ptr, ptr, i32, ptr }, ptr %0, i32 0, i32 3
  %7 = load ptr, ptr %6, align 4
  call void @"interface:{Print:func:{basic:string}{}}.Print$invoke"(ptr %1, ptr %3, i32 %5, ptr %7, ptr undef) #9
  ret void
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #1 = { nounwind "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #2 = { "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #3 = { nounwind "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" "tinygo-gowrapper"="main.regularFunction" }
attributes #4 = { nounwind "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" "tinygo-gowrapper"="main.inlineFunctionGoroutine$1" }
attributes #5 = { nounwind "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" "tinygo-gowrapper"="main.closureFunctionGoroutine$1" }
attributes #6 = { nounwind "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" "tinygo-gowrapper" }
attributes #7 = { "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" "tinygo-invoke"="reflect/methods.Print(string)" "tinygo-methods"="reflect/methods.Print(string)" }
attributes #8 = { nounwind "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" "tinygo-gowrapper"="interface:{Print:func:{basic:string}{}}.Print$invoke" }
attributes #9 = { nounwind }
