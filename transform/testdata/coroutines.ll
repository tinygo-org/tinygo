target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%"internal/task.state" = type { i8* }
%"internal/task.Task" = type { %"internal/task.Task", i8*, i32, %"internal/task.state" }

declare void @"internal/task.start"(i32, i8*, i32, i8*, i8*)
declare void @"internal/task.Pause"(i8*, i8*)

declare void @runtime.scheduler(i8*, i8*)

declare i8* @runtime.alloc(i32, i8*, i8*, i8*)
declare void @runtime.free(i8*, i8*, i8*)

declare %"internal/task.Task"* @"internal/task.Current"(i8*, i8*)

declare i8* @"(*internal/task.Task).setState"(%"internal/task.Task"*, i8*, i8*, i8*)
declare void @"(*internal/task.Task).setReturnPtr"(%"internal/task.Task"*, i8*, i8*, i8*)
declare i8* @"(*internal/task.Task).getReturnPtr"(%"internal/task.Task"*, i8*, i8*)
declare void @"(*internal/task.Task).returnTo"(%"internal/task.Task"*, i8*, i8*, i8*)
declare void @"(*internal/task.Task).returnCurrent"(%"internal/task.Task"*, i8*, i8*)
declare %"internal/task.Task"* @"internal/task.createTask"(i8*, i8*)

declare void @callMain(i8*, i8*)

; Test a simple sleep-like scenario.
declare void @enqueueTimer(%"internal/task.Task"*, i64, i8*, i8*)

define void @sleep(i64, i8*, i8* %parentHandle) {
entry:
  %2 = call %"internal/task.Task"* @"internal/task.Current"(i8* undef, i8* null)
  call void @enqueueTimer(%"internal/task.Task"* %2, i64 %0, i8* undef, i8* null)
  call void @"internal/task.Pause"(i8* undef, i8* null)
  ret void
}

; Test a delayed value return.
define i32 @delayedValue(i32, i64, i8*, i8* %parentHandle) {
entry:
  call void @sleep(i64 %1, i8* undef, i8* null)
  ret i32 %0
}

; Test a deadlocking async func.
define void @deadlock(i8*, i8* %parentHandle) {
entry:
  call void @"internal/task.Pause"(i8* undef, i8* null)
  unreachable
}

; Test a regular tail call.
define i32 @tail(i32, i64, i8*, i8* %parentHandle) {
entry:
  %3 = call i32 @delayedValue(i32 %0, i64 %1, i8* undef, i8* null)
  ret i32 %3
}

; Test a ditching tail call.
define void @ditchTail(i32, i64, i8*, i8* %parentHandle) {
entry:
  %3 = call i32 @delayedValue(i32 %0, i64 %1, i8* undef, i8* null)
  ret void
}

; Test a void tail call.
define void @voidTail(i32, i64, i8*, i8* %parentHandle) {
entry:
  call void @ditchTail(i32 %0, i64 %1, i8* undef, i8* null)
  ret void
}

; Test a tail call returning an alternate value.
define i32 @alternateTail(i32, i32, i64, i8*, i8* %parentHandle) {
entry:
  %4 = call i32 @delayedValue(i32 %1, i64 %2, i8* undef, i8* null)
  ret i32 %0
}

; Test a normal return from a coroutine.
; This must be turned into a coroutine.
define i1 @coroutine(i32, i64, i8*, i8* %parentHandle) {
entry:
  %3 = call i32 @delayedValue(i32 %0, i64 %1, i8* undef, i8* null)
  %4 = icmp eq i32 %3, 0
  ret i1 %4
}

; Normal function which should not be transformed.
define void @doNothing(i8*, i8* %parentHandle) {
entry:
  ret void
}

; Regression test: ensure that a tail call does not destroy the frame while it is still in use.
; Previously, the tail-call lowering transform would branch to the cleanup block after usePtr.
; This caused the lifetime of %a to be incorrectly reduced, and allowed the coroutine lowering transform to keep %a on the stack.
; After a suspend %a would be used, resulting in memory corruption.
define i8 @coroutineTailRegression(i8*, i8* %parentHandle) {
entry:
  %a = alloca i8
  store i8 5, i8* %a
  %val = call i8 @usePtr(i8* %a, i8* undef, i8* null)
  ret i8 %val
}

; Regression test: ensure that stack allocations alive during a suspend end up on the heap.
; This used to not be transformed to a coroutine, keeping %a on the stack.
; After a suspend %a would be used, resulting in memory corruption.
define i8 @allocaTailRegression(i8*, i8* %parentHandle) {
entry:
  %a = alloca i8
  call void @sleep(i64 1000000, i8* undef, i8* null)
  store i8 5, i8* %a
  %val = call i8 @usePtr(i8* %a, i8* undef, i8* null)
  ret i8 %val
}

; usePtr uses a pointer after a suspend.
define i8 @usePtr(i8*, i8*, i8* %parentHandle) {
entry:
  call void @sleep(i64 1000000, i8* undef, i8* null)
  %val = load i8, i8* %0
  ret i8 %val
}

; Goroutine that sleeps and does nothing.
; Should be a void tail call.
define void @sleepGoroutine(i8*, i8* %parentHandle) {
  call void @sleep(i64 1000000, i8* undef, i8* null)
  ret void
}

; Program main function.
define void @progMain(i8*, i8* %parentHandle) {
entry:
  ; Call a sync func in a goroutine.
  call void @"internal/task.start"(i32 ptrtoint (void (i8*, i8*)* @doNothing to i32), i8* undef, i32 undef, i8* undef, i8* null)
  ; Call an async func in a goroutine.
  call void @"internal/task.start"(i32 ptrtoint (void (i8*, i8*)* @sleepGoroutine to i32), i8* undef, i32 undef, i8* undef, i8* null)
  ; Sleep a bit.
  call void @sleep(i64 2000000, i8* undef, i8* null)
  ; Done.
  ret void
}

; Entrypoint of runtime.
define void @main() {
entry:
  call void @"internal/task.start"(i32 ptrtoint (void (i8*, i8*)* @progMain to i32), i8* undef, i32 undef, i8* undef, i8* null)
  call void @runtime.scheduler(i8* undef, i8* null)
  ret void
}
