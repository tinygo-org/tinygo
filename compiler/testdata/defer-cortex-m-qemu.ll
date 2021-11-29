; ModuleID = 'defer.go'
source_filename = "defer.go"
target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "thumbv7m-unknown-unknown-eabi"

%runtime._defer = type { i32, %runtime._defer* }

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context) unnamed_addr #0 {
entry:
  ret void
}

declare void @main.external(i8*)

; Function Attrs: nounwind
define hidden void @main.deferSimple(i8* %context) unnamed_addr #0 {
entry:
  %defer.alloca = alloca { i32, %runtime._defer* }, align 4
  %deferPtr = alloca %runtime._defer*, align 4
  store %runtime._defer* null, %runtime._defer** %deferPtr, align 4
  %defer.alloca.repack = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca, i32 0, i32 0
  store i32 0, i32* %defer.alloca.repack, align 4
  %defer.alloca.repack1 = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca, i32 0, i32 1
  store %runtime._defer* null, %runtime._defer** %defer.alloca.repack1, align 4
  %0 = bitcast %runtime._defer** %deferPtr to { i32, %runtime._defer* }**
  store { i32, %runtime._defer* }* %defer.alloca, { i32, %runtime._defer* }** %0, align 4
  call void @main.external(i8* undef) #0
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %rundefers.callback0, %entry
  %1 = load %runtime._defer*, %runtime._defer** %deferPtr, align 4
  %stackIsNil = icmp eq %runtime._defer* %1, null
  br i1 %stackIsNil, label %rundefers.end, label %rundefers.loop

rundefers.loop:                                   ; preds = %rundefers.loophead
  %stack.next.gep = getelementptr inbounds %runtime._defer, %runtime._defer* %1, i32 0, i32 1
  %stack.next = load %runtime._defer*, %runtime._defer** %stack.next.gep, align 4
  store %runtime._defer* %stack.next, %runtime._defer** %deferPtr, align 4
  %callback.gep = getelementptr inbounds %runtime._defer, %runtime._defer* %1, i32 0, i32 0
  %callback = load i32, i32* %callback.gep, align 4
  switch i32 %callback, label %rundefers.default [
    i32 0, label %rundefers.callback0
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  call void @"main.deferSimple$1"(i8* undef)
  br label %rundefers.loophead

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.end:                                    ; preds = %rundefers.loophead
  ret void

recover:                                          ; No predecessors!
  ret void
}

; Function Attrs: nounwind
define hidden void @"main.deferSimple$1"(i8* %context) unnamed_addr #0 {
entry:
  call void @runtime.printint32(i32 3, i8* undef) #0
  ret void
}

declare void @runtime.printint32(i32, i8*)

; Function Attrs: nounwind
define hidden void @main.deferMultiple(i8* %context) unnamed_addr #0 {
entry:
  %defer.alloca2 = alloca { i32, %runtime._defer* }, align 4
  %defer.alloca = alloca { i32, %runtime._defer* }, align 4
  %deferPtr = alloca %runtime._defer*, align 4
  store %runtime._defer* null, %runtime._defer** %deferPtr, align 4
  %defer.alloca.repack = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca, i32 0, i32 0
  store i32 0, i32* %defer.alloca.repack, align 4
  %defer.alloca.repack5 = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca, i32 0, i32 1
  store %runtime._defer* null, %runtime._defer** %defer.alloca.repack5, align 4
  %0 = bitcast %runtime._defer** %deferPtr to { i32, %runtime._defer* }**
  store { i32, %runtime._defer* }* %defer.alloca, { i32, %runtime._defer* }** %0, align 4
  %defer.alloca2.repack = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca2, i32 0, i32 0
  store i32 1, i32* %defer.alloca2.repack, align 4
  %defer.alloca2.repack6 = getelementptr inbounds { i32, %runtime._defer* }, { i32, %runtime._defer* }* %defer.alloca2, i32 0, i32 1
  %1 = bitcast %runtime._defer** %defer.alloca2.repack6 to { i32, %runtime._defer* }**
  store { i32, %runtime._defer* }* %defer.alloca, { i32, %runtime._defer* }** %1, align 4
  %2 = bitcast %runtime._defer** %deferPtr to { i32, %runtime._defer* }**
  store { i32, %runtime._defer* }* %defer.alloca2, { i32, %runtime._defer* }** %2, align 4
  call void @main.external(i8* undef) #0
  br label %rundefers.loophead

rundefers.loophead:                               ; preds = %rundefers.callback1, %rundefers.callback0, %entry
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
    i32 1, label %rundefers.callback1
  ]

rundefers.callback0:                              ; preds = %rundefers.loop
  call void @"main.deferMultiple$1"(i8* undef)
  br label %rundefers.loophead

rundefers.callback1:                              ; preds = %rundefers.loop
  call void @"main.deferMultiple$2"(i8* undef)
  br label %rundefers.loophead

rundefers.default:                                ; preds = %rundefers.loop
  unreachable

rundefers.end:                                    ; preds = %rundefers.loophead
  ret void

recover:                                          ; No predecessors!
  ret void
}

; Function Attrs: nounwind
define hidden void @"main.deferMultiple$1"(i8* %context) unnamed_addr #0 {
entry:
  call void @runtime.printint32(i32 3, i8* undef) #0
  ret void
}

; Function Attrs: nounwind
define hidden void @"main.deferMultiple$2"(i8* %context) unnamed_addr #0 {
entry:
  call void @runtime.printint32(i32 5, i8* undef) #0
  ret void
}

attributes #0 = { nounwind }
