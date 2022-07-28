; ModuleID = 'defer.go'
source_filename = "defer.go"
target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "thumbv7m-unknown-unknown-eabi"

%runtime._defer = type { i32, %runtime._defer* }
%runtime.deferFrame = type { i8*, i8*, [0 x i8*], %runtime.deferFrame*, i1, %runtime._interface }
%runtime._interface = type { i32, i8* }

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*) #0

; Function Attrs: nounwind
define hidden void @main.init(i8* %context) unnamed_addr #1 {
entry:
  ret void
}

declare void @main.external(i8*) #0

; Function Attrs: nounwind
define hidden void @main.deferSimple(i8* %context) unnamed_addr #1 {
entry:
  %defer.alloca = alloca { i32, %runtime._defer* }, align 4
  %deferPtr = alloca %runtime._defer*, align 4
  store %runtime._defer* null, %runtime._defer** %deferPtr, align 4
  %deferframe.buf = alloca %runtime.deferFrame, align 4
  %0 = call i8* @llvm.stacksave()
  call void @runtime.setupDeferFrame(%runtime.deferFrame* nonnull %deferframe.buf, i8* %0, i8* undef) #3
  %defer.alloca.repack = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca, i32 0, i32 0
  store i32 0, i32* %defer.alloca.repack, align 4
  %defer.alloca.repack16 = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca, i32 0, i32 1
  store %runtime._defer* null, %runtime._defer** %defer.alloca.repack16, align 4
  %1 = bitcast %runtime._defer** %deferPtr to { i32, %runtime._defer* }**
  store { i32, %runtime._defer* }* %defer.alloca, { i32, %runtime._defer* }** %1, align 4
  %setjmp = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(%runtime.deferFrame* nonnull %deferframe.buf) #4
  %setjmp.result = icmp eq i32 %setjmp, 0
  br i1 %setjmp.result, label %2, label %lpad

2:                                                ; preds = %entry
  call void @main.external(i8* undef) #3
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %4, %2
  %3 = load %runtime._defer*, %runtime._defer** %deferPtr, align 4
  %stackIsNil = icmp eq %runtime._defer* %3, null
  br i1 %stackIsNil, label %rundefers.end, label %rundefers.loop

rundefers.loop:                                   ; preds = %rundefers.loophead
  %stack.next.gep = getelementptr inbounds %runtime._defer, %runtime._defer* %3, i32 0, i32 1
  %stack.next = load %runtime._defer*, %runtime._defer** %stack.next.gep, align 4
  store %runtime._defer* %stack.next, %runtime._defer** %deferPtr, align 4
  %callback.gep = getelementptr inbounds %runtime._defer, %runtime._defer* %3, i32 0, i32 0
  %callback = load i32, i32* %callback.gep, align 4
  switch i32 %callback, label %rundefers.default [
    i32 0, label %rundefers.callback0
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  %setjmp1 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(%runtime.deferFrame* nonnull %deferframe.buf) #4
  %setjmp.result2 = icmp eq i32 %setjmp1, 0
  br i1 %setjmp.result2, label %4, label %lpad

4:                                                ; preds = %rundefers.callback0
  call void @"main.deferSimple$1"(i8* undef)
  br label %rundefers.loophead

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.end:                                    ; preds = %rundefers.loophead
  call void @runtime.destroyDeferFrame(%runtime.deferFrame* nonnull %deferframe.buf, i8* undef) #3
  ret void

recover:                                          ; preds = %rundefers.end3
  call void @runtime.destroyDeferFrame(%runtime.deferFrame* nonnull %deferframe.buf, i8* undef) #3
  ret void

lpad:                                             ; preds = %rundefers.callback012, %rundefers.callback0, %entry
  br label %rundefers.loophead6

rundefers.loophead6:                              ; preds = %6, %lpad
  %5 = load %runtime._defer*, %runtime._defer** %deferPtr, align 4
  %stackIsNil7 = icmp eq %runtime._defer* %5, null
  br i1 %stackIsNil7, label %rundefers.end3, label %rundefers.loop5

rundefers.loop5:                                  ; preds = %rundefers.loophead6
  %stack.next.gep8 = getelementptr inbounds %runtime._defer, %runtime._defer* %5, i32 0, i32 1
  %stack.next9 = load %runtime._defer*, %runtime._defer** %stack.next.gep8, align 4
  store %runtime._defer* %stack.next9, %runtime._defer** %deferPtr, align 4
  %callback.gep10 = getelementptr inbounds %runtime._defer, %runtime._defer* %5, i32 0, i32 0
  %callback11 = load i32, i32* %callback.gep10, align 4
  switch i32 %callback11, label %rundefers.default4 [
    i32 0, label %rundefers.callback012
  ]

rundefers.callback012:                            ; preds = %rundefers.loop5
  %setjmp14 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(%runtime.deferFrame* nonnull %deferframe.buf) #4
  %setjmp.result15 = icmp eq i32 %setjmp14, 0
  br i1 %setjmp.result15, label %6, label %lpad

6:                                                ; preds = %rundefers.callback012
  call void @"main.deferSimple$1"(i8* undef)
  br label %rundefers.loophead6

rundefers.default4:                               ; preds = %rundefers.loop5
  unreachable

rundefers.end3:                                   ; preds = %rundefers.loophead6
  br label %recover
}

; Function Attrs: nofree nosync nounwind willreturn
declare i8* @llvm.stacksave() #2

declare void @runtime.setupDeferFrame(%runtime.deferFrame* dereferenceable_or_null(24), i8*, i8*) #0

; Function Attrs: nounwind
define internal void @"main.deferSimple$1"(i8* %context) unnamed_addr #1 {
entry:
  call void @runtime.printint32(i32 3, i8* undef) #3
  ret void
}

declare void @runtime.destroyDeferFrame(%runtime.deferFrame* dereferenceable_or_null(24), i8*) #0

declare void @runtime.printint32(i32, i8*) #0

; Function Attrs: nounwind
define hidden void @main.deferMultiple(i8* %context) unnamed_addr #1 {
entry:
  %defer.alloca2 = alloca { i32, %runtime._defer* }, align 4
  %defer.alloca = alloca { i32, %runtime._defer* }, align 4
  %deferPtr = alloca %runtime._defer*, align 4
  store %runtime._defer* null, %runtime._defer** %deferPtr, align 4
  %deferframe.buf = alloca %runtime.deferFrame, align 4
  %0 = call i8* @llvm.stacksave()
  call void @runtime.setupDeferFrame(%runtime.deferFrame* nonnull %deferframe.buf, i8* %0, i8* undef) #3
  %defer.alloca.repack = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca, i32 0, i32 0
  store i32 0, i32* %defer.alloca.repack, align 4
  %defer.alloca.repack26 = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca, i32 0, i32 1
  store %runtime._defer* null, %runtime._defer** %defer.alloca.repack26, align 4
  %1 = bitcast %runtime._defer** %deferPtr to { i32, %runtime._defer* }**
  store { i32, %runtime._defer* }* %defer.alloca, { i32, %runtime._defer* }** %1, align 4
  %defer.alloca2.repack = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca2, i32 0, i32 0
  store i32 1, i32* %defer.alloca2.repack, align 4
  %defer.alloca2.repack27 = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca2, i32 0, i32 1
  %2 = bitcast %runtime._defer** %defer.alloca2.repack27 to { i32, %runtime._defer* }**
  store { i32, %runtime._defer* }* %defer.alloca, { i32, %runtime._defer* }** %2, align 4
  %3 = bitcast %runtime._defer** %deferPtr to { i32, %runtime._defer* }**
  store { i32, %runtime._defer* }* %defer.alloca2, { i32, %runtime._defer* }** %3, align 4
  %setjmp = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(%runtime.deferFrame* nonnull %deferframe.buf) #4
  %setjmp.result = icmp eq i32 %setjmp, 0
  br i1 %setjmp.result, label %4, label %lpad

4:                                                ; preds = %entry
  call void @main.external(i8* undef) #3
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %7, %6, %4
  %5 = load %runtime._defer*, %runtime._defer** %deferPtr, align 4
  %stackIsNil = icmp eq %runtime._defer* %5, null
  br i1 %stackIsNil, label %rundefers.end, label %rundefers.loop

rundefers.loop:                                   ; preds = %rundefers.loophead
  %stack.next.gep = getelementptr inbounds %runtime._defer, %runtime._defer* %5, i32 0, i32 1
  %stack.next = load %runtime._defer*, %runtime._defer** %stack.next.gep, align 4
  store %runtime._defer* %stack.next, %runtime._defer** %deferPtr, align 4
  %callback.gep = getelementptr inbounds %runtime._defer, %runtime._defer* %5, i32 0, i32 0
  %callback = load i32, i32* %callback.gep, align 4
  switch i32 %callback, label %rundefers.default [
    i32 0, label %rundefers.callback0
    i32 1, label %rundefers.callback1
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  %setjmp4 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(%runtime.deferFrame* nonnull %deferframe.buf) #4
  %setjmp.result5 = icmp eq i32 %setjmp4, 0
  br i1 %setjmp.result5, label %6, label %lpad

6:                                                ; preds = %rundefers.callback0
  call void @"main.deferMultiple$1"(i8* undef)
  br label %rundefers.loophead

rundefers.callback1:                              ; preds = %rundefers.loop
  %setjmp7 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(%runtime.deferFrame* nonnull %deferframe.buf) #4
  %setjmp.result8 = icmp eq i32 %setjmp7, 0
  br i1 %setjmp.result8, label %7, label %lpad

7:                                                ; preds = %rundefers.callback1
  call void @"main.deferMultiple$2"(i8* undef)
  br label %rundefers.loophead

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.end:                                    ; preds = %rundefers.loophead
  call void @runtime.destroyDeferFrame(%runtime.deferFrame* nonnull %deferframe.buf, i8* undef) #3
  ret void

recover:                                          ; preds = %rundefers.end9
  call void @runtime.destroyDeferFrame(%runtime.deferFrame* nonnull %deferframe.buf, i8* undef) #3
  ret void

lpad:                                             ; preds = %rundefers.callback122, %rundefers.callback018, %rundefers.callback1, %rundefers.callback0, %entry
  br label %rundefers.loophead12

rundefers.loophead12:                             ; preds = %10, %9, %lpad
  %8 = load %runtime._defer*, %runtime._defer** %deferPtr, align 4
  %stackIsNil13 = icmp eq %runtime._defer* %8, null
  br i1 %stackIsNil13, label %rundefers.end9, label %rundefers.loop11

rundefers.loop11:                                 ; preds = %rundefers.loophead12
  %stack.next.gep14 = getelementptr inbounds %runtime._defer, %runtime._defer* %8, i32 0, i32 1
  %stack.next15 = load %runtime._defer*, %runtime._defer** %stack.next.gep14, align 4
  store %runtime._defer* %stack.next15, %runtime._defer** %deferPtr, align 4
  %callback.gep16 = getelementptr inbounds %runtime._defer, %runtime._defer* %8, i32 0, i32 0
  %callback17 = load i32, i32* %callback.gep16, align 4
  switch i32 %callback17, label %rundefers.default10 [
    i32 0, label %rundefers.callback018
    i32 1, label %rundefers.callback122
  ]

rundefers.callback018:                            ; preds = %rundefers.loop11
  %setjmp20 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(%runtime.deferFrame* nonnull %deferframe.buf) #4
  %setjmp.result21 = icmp eq i32 %setjmp20, 0
  br i1 %setjmp.result21, label %9, label %lpad

9:                                                ; preds = %rundefers.callback018
  call void @"main.deferMultiple$1"(i8* undef)
  br label %rundefers.loophead12

rundefers.callback122:                            ; preds = %rundefers.loop11
  %setjmp24 = call i32 asm "\0Amovs r0, #0\0Amov r2, pc\0Astr r2, [r1, #4]", "={r0},{r1},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r11},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"(%runtime.deferFrame* nonnull %deferframe.buf) #4
  %setjmp.result25 = icmp eq i32 %setjmp24, 0
  br i1 %setjmp.result25, label %10, label %lpad

10:                                               ; preds = %rundefers.callback122
  call void @"main.deferMultiple$2"(i8* undef)
  br label %rundefers.loophead12

rundefers.default10:                              ; preds = %rundefers.loop11
  unreachable

rundefers.end9:                                   ; preds = %rundefers.loophead12
  br label %recover
}

; Function Attrs: nounwind
define internal void @"main.deferMultiple$1"(i8* %context) unnamed_addr #1 {
entry:
  call void @runtime.printint32(i32 3, i8* undef) #3
  ret void
}

; Function Attrs: nounwind
define internal void @"main.deferMultiple$2"(i8* %context) unnamed_addr #1 {
entry:
  call void @runtime.printint32(i32 5, i8* undef) #3
  ret void
}

attributes #0 = { "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #1 = { nounwind "target-features"="+armv7-m,+hwdiv,+soft-float,+strict-align,+thumb-mode,-aes,-bf16,-cdecp0,-cdecp1,-cdecp2,-cdecp3,-cdecp4,-cdecp5,-cdecp6,-cdecp7,-crc,-crypto,-d32,-dotprod,-dsp,-fp-armv8,-fp-armv8d16,-fp-armv8d16sp,-fp-armv8sp,-fp16,-fp16fml,-fp64,-fpregs,-fullfp16,-hwdiv-arm,-i8mm,-lob,-mve,-mve.fp,-neon,-pacbti,-ras,-sb,-sha2,-vfp2,-vfp2sp,-vfp3,-vfp3d16,-vfp3d16sp,-vfp3sp,-vfp4,-vfp4d16,-vfp4d16sp,-vfp4sp" }
attributes #2 = { nofree nosync nounwind willreturn }
attributes #3 = { nounwind }
attributes #4 = { nounwind returns_twice }
