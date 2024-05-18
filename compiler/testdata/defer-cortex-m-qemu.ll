; ModuleID = 'defer.go'
source_filename = "defer.go"
target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "thumbv7m-unknown-unknown-eabi"

%runtime.deferFrame = type { ptr, ptr, [0 x ptr], ptr, i1, %runtime._interface }
%runtime._interface = type { ptr, ptr }
%runtime._defer = type { i32, ptr }

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

declare void @main.external(ptr) #2

; Function Attrs: nounwind
define hidden void @main.deferSimple(ptr %context) unnamed_addr #1 {
entry:
  %defer.alloca = alloca { i32, ptr }, align 4
  %deferPtr = alloca ptr, align 4
  store ptr null, ptr %deferPtr, align 4
  %deferframe.buf = alloca %runtime.deferFrame, align 4
  %0 = call ptr @llvm.stacksave.p0()
  call void @runtime.setupDeferFrame(ptr nonnull %deferframe.buf, ptr %0, ptr undef) #4
  store i32 0, ptr %defer.alloca, align 4
  %defer.alloca.repack15 = getelementptr inbounds { i32, ptr }, ptr %defer.alloca, i32 0, i32 1
  store ptr null, ptr %defer.alloca.repack15, align 4
  store ptr %defer.alloca, ptr %deferPtr, align 4
  %setjmp = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(ptr nonnull %deferframe.buf) #5
  %setjmp.result = icmp eq i32 %setjmp, 0
  br i1 %setjmp.result, label %1, label %lpad

1:                                                ; preds = %entry
  call void @main.external(ptr undef) #4
  br label %rundefers.block

rundefers.after:                                  ; preds = %rundefers.end
  call void @runtime.destroyDeferFrame(ptr nonnull %deferframe.buf, ptr undef) #4
  ret void

rundefers.block:                                  ; preds = %1
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %3, %rundefers.block
  %2 = load ptr, ptr %deferPtr, align 4
  %stackIsNil = icmp eq ptr %2, null
  br i1 %stackIsNil, label %rundefers.end, label %rundefers.loop

rundefers.loop:                                   ; preds = %rundefers.loophead
  %stack.next.gep = getelementptr inbounds %runtime._defer, ptr %2, i32 0, i32 1
  %stack.next = load ptr, ptr %stack.next.gep, align 4
  store ptr %stack.next, ptr %deferPtr, align 4
  %callback = load i32, ptr %2, align 4
  switch i32 %callback, label %rundefers.default [
    i32 0, label %rundefers.callback0
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  %setjmp1 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(ptr nonnull %deferframe.buf) #5
  %setjmp.result2 = icmp eq i32 %setjmp1, 0
  br i1 %setjmp.result2, label %3, label %lpad

3:                                                ; preds = %rundefers.callback0
  call void @"main.deferSimple$1"(ptr undef)
  br label %rundefers.loophead

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.end:                                    ; preds = %rundefers.loophead
  br label %rundefers.after

recover:                                          ; preds = %rundefers.end3
  call void @runtime.destroyDeferFrame(ptr nonnull %deferframe.buf, ptr undef) #4
  ret void

lpad:                                             ; preds = %rundefers.callback012, %rundefers.callback0, %entry
  br label %rundefers.loophead6

rundefers.loophead6:                              ; preds = %5, %lpad
  %4 = load ptr, ptr %deferPtr, align 4
  %stackIsNil7 = icmp eq ptr %4, null
  br i1 %stackIsNil7, label %rundefers.end3, label %rundefers.loop5

rundefers.loop5:                                  ; preds = %rundefers.loophead6
  %stack.next.gep8 = getelementptr inbounds %runtime._defer, ptr %4, i32 0, i32 1
  %stack.next9 = load ptr, ptr %stack.next.gep8, align 4
  store ptr %stack.next9, ptr %deferPtr, align 4
  %callback11 = load i32, ptr %4, align 4
  switch i32 %callback11, label %rundefers.default4 [
    i32 0, label %rundefers.callback012
  ]

rundefers.callback012:                            ; preds = %rundefers.loop5
  %setjmp13 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(ptr nonnull %deferframe.buf) #5
  %setjmp.result14 = icmp eq i32 %setjmp13, 0
  br i1 %setjmp.result14, label %5, label %lpad

5:                                                ; preds = %rundefers.callback012
  call void @"main.deferSimple$1"(ptr undef)
  br label %rundefers.loophead6

rundefers.default4:                               ; preds = %rundefers.loop5
  unreachable

rundefers.end3:                                   ; preds = %rundefers.loophead6
  br label %recover
}

; Function Attrs: nocallback nofree nosync nounwind willreturn
declare ptr @llvm.stacksave.p0() #3

declare void @runtime.setupDeferFrame(ptr dereferenceable_or_null(24), ptr, ptr) #2

declare void @runtime.destroyDeferFrame(ptr dereferenceable_or_null(24), ptr) #2

; Function Attrs: nounwind
define internal void @"main.deferSimple$1"(ptr %context) unnamed_addr #1 {
entry:
  call void @runtime.printint32(i32 3, ptr undef) #4
  ret void
}

declare void @runtime.printint32(i32, ptr) #2

; Function Attrs: nounwind
define hidden void @main.deferMultiple(ptr %context) unnamed_addr #1 {
entry:
  %defer.alloca2 = alloca { i32, ptr }, align 4
  %defer.alloca = alloca { i32, ptr }, align 4
  %deferPtr = alloca ptr, align 4
  store ptr null, ptr %deferPtr, align 4
  %deferframe.buf = alloca %runtime.deferFrame, align 4
  %0 = call ptr @llvm.stacksave.p0()
  call void @runtime.setupDeferFrame(ptr nonnull %deferframe.buf, ptr %0, ptr undef) #4
  store i32 0, ptr %defer.alloca, align 4
  %defer.alloca.repack22 = getelementptr inbounds { i32, ptr }, ptr %defer.alloca, i32 0, i32 1
  store ptr null, ptr %defer.alloca.repack22, align 4
  store ptr %defer.alloca, ptr %deferPtr, align 4
  store i32 1, ptr %defer.alloca2, align 4
  %defer.alloca2.repack23 = getelementptr inbounds { i32, ptr }, ptr %defer.alloca2, i32 0, i32 1
  store ptr %defer.alloca, ptr %defer.alloca2.repack23, align 4
  store ptr %defer.alloca2, ptr %deferPtr, align 4
  %setjmp = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(ptr nonnull %deferframe.buf) #5
  %setjmp.result = icmp eq i32 %setjmp, 0
  br i1 %setjmp.result, label %1, label %lpad

1:                                                ; preds = %entry
  call void @main.external(ptr undef) #4
  br label %rundefers.block

rundefers.after:                                  ; preds = %rundefers.end
  call void @runtime.destroyDeferFrame(ptr nonnull %deferframe.buf, ptr undef) #4
  ret void

rundefers.block:                                  ; preds = %1
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %4, %3, %rundefers.block
  %2 = load ptr, ptr %deferPtr, align 4
  %stackIsNil = icmp eq ptr %2, null
  br i1 %stackIsNil, label %rundefers.end, label %rundefers.loop

rundefers.loop:                                   ; preds = %rundefers.loophead
  %stack.next.gep = getelementptr inbounds %runtime._defer, ptr %2, i32 0, i32 1
  %stack.next = load ptr, ptr %stack.next.gep, align 4
  store ptr %stack.next, ptr %deferPtr, align 4
  %callback = load i32, ptr %2, align 4
  switch i32 %callback, label %rundefers.default [
    i32 0, label %rundefers.callback0
    i32 1, label %rundefers.callback1
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  %setjmp3 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(ptr nonnull %deferframe.buf) #5
  %setjmp.result4 = icmp eq i32 %setjmp3, 0
  br i1 %setjmp.result4, label %3, label %lpad

3:                                                ; preds = %rundefers.callback0
  call void @"main.deferMultiple$1"(ptr undef)
  br label %rundefers.loophead

rundefers.callback1:                              ; preds = %rundefers.loop
  %setjmp5 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(ptr nonnull %deferframe.buf) #5
  %setjmp.result6 = icmp eq i32 %setjmp5, 0
  br i1 %setjmp.result6, label %4, label %lpad

4:                                                ; preds = %rundefers.callback1
  call void @"main.deferMultiple$2"(ptr undef)
  br label %rundefers.loophead

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.end:                                    ; preds = %rundefers.loophead
  br label %rundefers.after

recover:                                          ; preds = %rundefers.end7
  call void @runtime.destroyDeferFrame(ptr nonnull %deferframe.buf, ptr undef) #4
  ret void

lpad:                                             ; preds = %rundefers.callback119, %rundefers.callback016, %rundefers.callback1, %rundefers.callback0, %entry
  br label %rundefers.loophead10

rundefers.loophead10:                             ; preds = %7, %6, %lpad
  %5 = load ptr, ptr %deferPtr, align 4
  %stackIsNil11 = icmp eq ptr %5, null
  br i1 %stackIsNil11, label %rundefers.end7, label %rundefers.loop9

rundefers.loop9:                                  ; preds = %rundefers.loophead10
  %stack.next.gep12 = getelementptr inbounds %runtime._defer, ptr %5, i32 0, i32 1
  %stack.next13 = load ptr, ptr %stack.next.gep12, align 4
  store ptr %stack.next13, ptr %deferPtr, align 4
  %callback15 = load i32, ptr %5, align 4
  switch i32 %callback15, label %rundefers.default8 [
    i32 0, label %rundefers.callback016
    i32 1, label %rundefers.callback119
  ]

rundefers.callback016:                            ; preds = %rundefers.loop9
  %setjmp17 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(ptr nonnull %deferframe.buf) #5
  %setjmp.result18 = icmp eq i32 %setjmp17, 0
  br i1 %setjmp.result18, label %6, label %lpad

6:                                                ; preds = %rundefers.callback016
  call void @"main.deferMultiple$1"(ptr undef)
  br label %rundefers.loophead10

rundefers.callback119:                            ; preds = %rundefers.loop9
  %setjmp20 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(ptr nonnull %deferframe.buf) #5
  %setjmp.result21 = icmp eq i32 %setjmp20, 0
  br i1 %setjmp.result21, label %7, label %lpad

7:                                                ; preds = %rundefers.callback119
  call void @"main.deferMultiple$2"(ptr undef)
  br label %rundefers.loophead10

rundefers.default8:                               ; preds = %rundefers.loop9
  unreachable

rundefers.end7:                                   ; preds = %rundefers.loophead10
  br label %recover
}

; Function Attrs: nounwind
define internal void @"main.deferMultiple$1"(ptr %context) unnamed_addr #1 {
entry:
  call void @runtime.printint32(i32 3, ptr undef) #4
  ret void
}

; Function Attrs: nounwind
define internal void @"main.deferMultiple$2"(ptr %context) unnamed_addr #1 {
entry:
  call void @runtime.printint32(i32 5, ptr undef) #4
  ret void
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #1 = { nounwind "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #2 = { "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #3 = { nocallback nofree nosync nounwind willreturn }
attributes #4 = { nounwind }
attributes #5 = { nounwind returns_twice }
