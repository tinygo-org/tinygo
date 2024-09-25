; ModuleID = 'defer.go'
source_filename = "defer.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.deferFrameWasmEH = type { ptr, i1, %runtime._interface }
%runtime._interface = type { ptr, ptr }
%runtime._defer = type { i32, ptr }

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

; Function Attrs: nounwind
declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

declare void @main.external(ptr) #2

define hidden void @main.deferSimple(ptr %context) unnamed_addr #2 personality ptr @__gxx_wasm_personality_v0 {
entry:
  %defer.alloca = alloca { i32, ptr }, align 8
  %deferPtr = alloca ptr, align 4
  store ptr null, ptr %deferPtr, align 4
  %deferframe.buf = alloca %runtime.deferFrameWasmEH, align 8
  call void @runtime.setupDeferFrameWasmEH(ptr nonnull %deferframe.buf, ptr undef)
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull %defer.alloca, ptr nonnull %stackalloc, ptr undef)
  store i32 0, ptr %defer.alloca, align 4
  %defer.alloca.repack13 = getelementptr inbounds { i32, ptr }, ptr %defer.alloca, i32 0, i32 1
  store ptr null, ptr %defer.alloca.repack13, align 4
  store ptr %defer.alloca, ptr %deferPtr, align 4
  invoke void @main.external(ptr undef)
          to label %invoke.cont unwind label %catch.dispatch

invoke.cont:                                      ; preds = %entry
  br label %rundefers.block

rundefers.after:                                  ; preds = %rundefers.end
  call void @runtime.destroyDeferFrameWasmEH(ptr nonnull %deferframe.buf, ptr undef)
  ret void

rundefers.block:                                  ; preds = %invoke.cont
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %rundefers.catch.start, %rundefers.callback0, %rundefers.block
  %0 = load ptr, ptr %deferPtr, align 4
  %stackIsNil = icmp eq ptr %0, null
  br i1 %stackIsNil, label %rundefers.end, label %rundefers.loop

rundefers.loop:                                   ; preds = %rundefers.loophead
  %stack.next.gep = getelementptr inbounds %runtime._defer, ptr %0, i32 0, i32 1
  %stack.next = load ptr, ptr %stack.next.gep, align 4
  store ptr %stack.next, ptr %deferPtr, align 4
  %callback = load i32, ptr %0, align 4
  switch i32 %callback, label %rundefers.default [
    i32 0, label %rundefers.callback0
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  invoke void @"main.deferSimple$1"(ptr undef)
          to label %rundefers.loophead unwind label %rundefers.catch.dispatch

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.catch.dispatch:                         ; preds = %rundefers.callback0
  %1 = catchswitch within none [label %rundefers.catch.start] unwind to caller

rundefers.catch.start:                            ; preds = %rundefers.catch.dispatch
  %2 = catchpad within %1 [ptr null]
  catchret from %2 to label %rundefers.loophead

rundefers.end:                                    ; preds = %rundefers.loophead
  br label %rundefers.after

recover:                                          ; preds = %rundefers.end1
  call void @runtime.destroyDeferFrameWasmEH(ptr nonnull %deferframe.buf, ptr undef)
  ret void

catch.dispatch:                                   ; preds = %entry
  %3 = catchswitch within none [label %catch.start] unwind to caller

catch.start:                                      ; preds = %catch.dispatch
  %4 = catchpad within %3 [ptr null]
  catchret from %4 to label %rundefers

rundefers:                                        ; preds = %catch.start
  br label %rundefers.loophead4

rundefers.loophead4:                              ; preds = %rundefers.catch.start6, %rundefers.callback012, %rundefers
  %5 = load ptr, ptr %deferPtr, align 4
  %stackIsNil7 = icmp eq ptr %5, null
  br i1 %stackIsNil7, label %rundefers.end1, label %rundefers.loop3

rundefers.loop3:                                  ; preds = %rundefers.loophead4
  %stack.next.gep8 = getelementptr inbounds %runtime._defer, ptr %5, i32 0, i32 1
  %stack.next9 = load ptr, ptr %stack.next.gep8, align 4
  store ptr %stack.next9, ptr %deferPtr, align 4
  %callback11 = load i32, ptr %5, align 4
  switch i32 %callback11, label %rundefers.default2 [
    i32 0, label %rundefers.callback012
  ]

rundefers.callback012:                            ; preds = %rundefers.loop3
  invoke void @"main.deferSimple$1"(ptr undef)
          to label %rundefers.loophead4 unwind label %rundefers.catch.dispatch5

rundefers.default2:                               ; preds = %rundefers.loop3
  unreachable

rundefers.catch.dispatch5:                        ; preds = %rundefers.callback012
  %6 = catchswitch within none [label %rundefers.catch.start6] unwind to caller

rundefers.catch.start6:                           ; preds = %rundefers.catch.dispatch5
  %7 = catchpad within %6 [ptr null]
  catchret from %7 to label %rundefers.loophead4

rundefers.end1:                                   ; preds = %rundefers.loophead4
  br label %recover
}

declare i32 @__gxx_wasm_personality_v0(...)

declare void @runtime.setupDeferFrameWasmEH(ptr dereferenceable_or_null(16), ptr) #2

declare void @runtime.destroyDeferFrameWasmEH(ptr dereferenceable_or_null(16), ptr) #2

define internal void @"main.deferSimple$1"(ptr %context) unnamed_addr #2 {
entry:
  call void @runtime.printint32(i32 3, ptr undef)
  ret void
}

declare void @runtime.printint32(i32, ptr) #2

define hidden void @main.deferMultiple(ptr %context) unnamed_addr #2 personality ptr @__gxx_wasm_personality_v0 {
entry:
  %defer.alloca2 = alloca { i32, ptr }, align 8
  %defer.alloca = alloca { i32, ptr }, align 8
  %deferPtr = alloca ptr, align 4
  store ptr null, ptr %deferPtr, align 4
  %deferframe.buf = alloca %runtime.deferFrameWasmEH, align 8
  call void @runtime.setupDeferFrameWasmEH(ptr nonnull %deferframe.buf, ptr undef)
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull %defer.alloca, ptr nonnull %stackalloc, ptr undef)
  store i32 0, ptr %defer.alloca, align 4
  %defer.alloca.repack16 = getelementptr inbounds { i32, ptr }, ptr %defer.alloca, i32 0, i32 1
  store ptr null, ptr %defer.alloca.repack16, align 4
  store ptr %defer.alloca, ptr %deferPtr, align 4
  call void @runtime.trackPointer(ptr nonnull %defer.alloca2, ptr nonnull %stackalloc, ptr undef)
  store i32 1, ptr %defer.alloca2, align 4
  %defer.alloca2.repack17 = getelementptr inbounds { i32, ptr }, ptr %defer.alloca2, i32 0, i32 1
  store ptr %defer.alloca, ptr %defer.alloca2.repack17, align 4
  store ptr %defer.alloca2, ptr %deferPtr, align 4
  invoke void @main.external(ptr undef)
          to label %invoke.cont unwind label %catch.dispatch

invoke.cont:                                      ; preds = %entry
  br label %rundefers.block

rundefers.after:                                  ; preds = %rundefers.end
  call void @runtime.destroyDeferFrameWasmEH(ptr nonnull %deferframe.buf, ptr undef)
  ret void

rundefers.block:                                  ; preds = %invoke.cont
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %rundefers.catch.start, %rundefers.callback1, %rundefers.callback0, %rundefers.block
  %0 = load ptr, ptr %deferPtr, align 4
  %stackIsNil = icmp eq ptr %0, null
  br i1 %stackIsNil, label %rundefers.end, label %rundefers.loop

rundefers.loop:                                   ; preds = %rundefers.loophead
  %stack.next.gep = getelementptr inbounds %runtime._defer, ptr %0, i32 0, i32 1
  %stack.next = load ptr, ptr %stack.next.gep, align 4
  store ptr %stack.next, ptr %deferPtr, align 4
  %callback = load i32, ptr %0, align 4
  switch i32 %callback, label %rundefers.default [
    i32 0, label %rundefers.callback0
    i32 1, label %rundefers.callback1
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  invoke void @"main.deferMultiple$1"(ptr undef)
          to label %rundefers.loophead unwind label %rundefers.catch.dispatch

rundefers.callback1:                              ; preds = %rundefers.loop
  invoke void @"main.deferMultiple$2"(ptr undef)
          to label %rundefers.loophead unwind label %rundefers.catch.dispatch

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.catch.dispatch:                         ; preds = %rundefers.callback1, %rundefers.callback0
  %1 = catchswitch within none [label %rundefers.catch.start] unwind to caller

rundefers.catch.start:                            ; preds = %rundefers.catch.dispatch
  %2 = catchpad within %1 [ptr null]
  catchret from %2 to label %rundefers.loophead

rundefers.end:                                    ; preds = %rundefers.loophead
  br label %rundefers.after

recover:                                          ; preds = %rundefers.end3
  call void @runtime.destroyDeferFrameWasmEH(ptr nonnull %deferframe.buf, ptr undef)
  ret void

catch.dispatch:                                   ; preds = %entry
  %3 = catchswitch within none [label %catch.start] unwind to caller

catch.start:                                      ; preds = %catch.dispatch
  %4 = catchpad within %3 [ptr null]
  catchret from %4 to label %rundefers

rundefers:                                        ; preds = %catch.start
  br label %rundefers.loophead6

rundefers.loophead6:                              ; preds = %rundefers.catch.start8, %rundefers.callback115, %rundefers.callback014, %rundefers
  %5 = load ptr, ptr %deferPtr, align 4
  %stackIsNil9 = icmp eq ptr %5, null
  br i1 %stackIsNil9, label %rundefers.end3, label %rundefers.loop5

rundefers.loop5:                                  ; preds = %rundefers.loophead6
  %stack.next.gep10 = getelementptr inbounds %runtime._defer, ptr %5, i32 0, i32 1
  %stack.next11 = load ptr, ptr %stack.next.gep10, align 4
  store ptr %stack.next11, ptr %deferPtr, align 4
  %callback13 = load i32, ptr %5, align 4
  switch i32 %callback13, label %rundefers.default4 [
    i32 0, label %rundefers.callback014
    i32 1, label %rundefers.callback115
  ]

rundefers.callback014:                            ; preds = %rundefers.loop5
  invoke void @"main.deferMultiple$1"(ptr undef)
          to label %rundefers.loophead6 unwind label %rundefers.catch.dispatch7

rundefers.callback115:                            ; preds = %rundefers.loop5
  invoke void @"main.deferMultiple$2"(ptr undef)
          to label %rundefers.loophead6 unwind label %rundefers.catch.dispatch7

rundefers.default4:                               ; preds = %rundefers.loop5
  unreachable

rundefers.catch.dispatch7:                        ; preds = %rundefers.callback115, %rundefers.callback014
  %6 = catchswitch within none [label %rundefers.catch.start8] unwind to caller

rundefers.catch.start8:                           ; preds = %rundefers.catch.dispatch7
  %7 = catchpad within %6 [ptr null]
  catchret from %7 to label %rundefers.loophead6

rundefers.end3:                                   ; preds = %rundefers.loophead6
  br label %recover
}

define internal void @"main.deferMultiple$1"(ptr %context) unnamed_addr #2 {
entry:
  call void @runtime.printint32(i32 3, ptr undef)
  ret void
}

define internal void @"main.deferMultiple$2"(ptr %context) unnamed_addr #2 {
entry:
  call void @runtime.printint32(i32 5, ptr undef)
  ret void
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+exception-handling,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+exception-handling,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { "target-features"="+bulk-memory,+exception-handling,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
