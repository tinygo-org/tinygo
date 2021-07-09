target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%"internal/task.Task" = type { %"internal/task.Task", i8*, i32, %"internal/task.state" }
%"internal/task.state" = type { i8* }

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

declare void @enqueueTimer(%"internal/task.Task"*, i64, i8*, i8*)

define void @sleep(i64 %0, i8* %1, i8* %parentHandle) {
entry:
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  %task.current1 = bitcast i8* %parentHandle to %"internal/task.Task"*
  call void @enqueueTimer(%"internal/task.Task"* %task.current1, i64 %0, i8* undef, i8* null)
  ret void
}

define i32 @delayedValue(i32 %0, i64 %1, i8* %2, i8* %parentHandle) {
entry:
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  %ret.ptr = call i8* @"(*internal/task.Task).getReturnPtr"(%"internal/task.Task"* %task.current, i8* undef, i8* undef)
  %ret.ptr.bitcast = bitcast i8* %ret.ptr to i32*
  store i32 %0, i32* %ret.ptr.bitcast, align 4
  call void @sleep(i64 %1, i8* undef, i8* %parentHandle)
  ret i32 undef
}

define void @deadlock(i8* %0, i8* %parentHandle) {
entry:
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  ret void
}

define i32 @tail(i32 %0, i64 %1, i8* %2, i8* %parentHandle) {
entry:
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  %3 = call i32 @delayedValue(i32 %0, i64 %1, i8* undef, i8* %parentHandle)
  ret i32 undef
}

define void @ditchTail(i32 %0, i64 %1, i8* %2, i8* %parentHandle) {
entry:
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  %ret.ditch = call i8* @runtime.alloc(i32 4, i8* null, i8* undef, i8* undef)
  call void @"(*internal/task.Task).setReturnPtr"(%"internal/task.Task"* %task.current, i8* %ret.ditch, i8* undef, i8* undef)
  %3 = call i32 @delayedValue(i32 %0, i64 %1, i8* undef, i8* %parentHandle)
  ret void
}

define void @voidTail(i32 %0, i64 %1, i8* %2, i8* %parentHandle) {
entry:
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  call void @ditchTail(i32 %0, i64 %1, i8* undef, i8* %parentHandle)
  ret void
}

define i32 @alternateTail(i32 %0, i32 %1, i64 %2, i8* %3, i8* %parentHandle) {
entry:
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  %ret.ptr = call i8* @"(*internal/task.Task).getReturnPtr"(%"internal/task.Task"* %task.current, i8* undef, i8* undef)
  %ret.ptr.bitcast = bitcast i8* %ret.ptr to i32*
  store i32 %0, i32* %ret.ptr.bitcast, align 4
  %ret.alternate = call i8* @runtime.alloc(i32 4, i8* null, i8* undef, i8* undef)
  call void @"(*internal/task.Task).setReturnPtr"(%"internal/task.Task"* %task.current, i8* %ret.alternate, i8* undef, i8* undef)
  %4 = call i32 @delayedValue(i32 %1, i64 %2, i8* undef, i8* %parentHandle)
  ret i32 undef
}

define i1 @coroutine(i32 %0, i64 %1, i8* %2, i8* %parentHandle) {
entry:
  %call.return = alloca i32, align 4
  %coro.id = call token @llvm.coro.id(i32 0, i8* null, i8* null, i8* null)
  %coro.size = call i32 @llvm.coro.size.i32()
  %coro.alloc = call i8* @runtime.alloc(i32 %coro.size, i8* null, i8* undef, i8* undef)
  %coro.state = call i8* @llvm.coro.begin(token %coro.id, i8* %coro.alloc)
  %task.current2 = bitcast i8* %parentHandle to %"internal/task.Task"*
  %task.state.parent = call i8* @"(*internal/task.Task).setState"(%"internal/task.Task"* %task.current2, i8* %coro.state, i8* undef, i8* undef)
  %task.retPtr = call i8* @"(*internal/task.Task).getReturnPtr"(%"internal/task.Task"* %task.current2, i8* undef, i8* undef)
  %task.retPtr.bitcast = bitcast i8* %task.retPtr to i1*
  %call.return.bitcast = bitcast i32* %call.return to i8*
  call void @llvm.lifetime.start.p0i8(i64 4, i8* %call.return.bitcast)
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  %call.return.bitcast1 = bitcast i32* %call.return to i8*
  call void @"(*internal/task.Task).setReturnPtr"(%"internal/task.Task"* %task.current, i8* %call.return.bitcast1, i8* undef, i8* undef)
  %3 = call i32 @delayedValue(i32 %0, i64 %1, i8* undef, i8* %parentHandle)
  %coro.save = call token @llvm.coro.save(i8* %coro.state)
  %call.suspend = call i8 @llvm.coro.suspend(token %coro.save, i1 false)
  switch i8 %call.suspend, label %suspend [
    i8 0, label %wakeup
    i8 1, label %cleanup
  ]

wakeup:                                           ; preds = %entry
  %4 = load i32, i32* %call.return, align 4
  call void @llvm.lifetime.end.p0i8(i64 4, i8* %call.return.bitcast)
  %5 = icmp eq i32 %4, 0
  store i1 %5, i1* %task.retPtr.bitcast, align 1
  call void @"(*internal/task.Task).returnTo"(%"internal/task.Task"* %task.current2, i8* %task.state.parent, i8* undef, i8* undef)
  br label %cleanup

suspend:                                          ; preds = %entry, %cleanup
  %unused = call i1 @llvm.coro.end(i8* %coro.state, i1 false)
  ret i1 undef

cleanup:                                          ; preds = %entry, %wakeup
  %coro.memFree = call i8* @llvm.coro.free(token %coro.id, i8* %coro.state)
  call void @runtime.free(i8* %coro.memFree, i8* undef, i8* undef)
  br label %suspend
}

define void @doNothing(i8* %0, i8* %parentHandle) {
entry:
  ret void
}

define i8 @coroutineTailRegression(i8* %0, i8* %parentHandle) {
entry:
  %a = alloca i8, align 1
  %coro.id = call token @llvm.coro.id(i32 0, i8* null, i8* null, i8* null)
  %coro.size = call i32 @llvm.coro.size.i32()
  %coro.alloc = call i8* @runtime.alloc(i32 %coro.size, i8* null, i8* undef, i8* undef)
  %coro.state = call i8* @llvm.coro.begin(token %coro.id, i8* %coro.alloc)
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  %task.state.parent = call i8* @"(*internal/task.Task).setState"(%"internal/task.Task"* %task.current, i8* %coro.state, i8* undef, i8* undef)
  %task.retPtr = call i8* @"(*internal/task.Task).getReturnPtr"(%"internal/task.Task"* %task.current, i8* undef, i8* undef)
  store i8 5, i8* %a, align 1
  %coro.state.restore = call i8* @"(*internal/task.Task).setState"(%"internal/task.Task"* %task.current, i8* %task.state.parent, i8* undef, i8* undef)
  call void @"(*internal/task.Task).setReturnPtr"(%"internal/task.Task"* %task.current, i8* %task.retPtr, i8* undef, i8* undef)
  %val = call i8 @usePtr(i8* %a, i8* undef, i8* %parentHandle)
  br label %post.tail

suspend:                                          ; preds = %post.tail, %cleanup
  %unused = call i1 @llvm.coro.end(i8* %coro.state, i1 false)
  ret i8 undef

cleanup:                                          ; preds = %post.tail
  %coro.memFree = call i8* @llvm.coro.free(token %coro.id, i8* %coro.state)
  call void @runtime.free(i8* %coro.memFree, i8* undef, i8* undef)
  br label %suspend

post.tail:                                        ; preds = %entry
  %coro.save = call token @llvm.coro.save(i8* %coro.state)
  %call.suspend = call i8 @llvm.coro.suspend(token %coro.save, i1 false)
  switch i8 %call.suspend, label %suspend [
    i8 0, label %unreachable
    i8 1, label %cleanup
  ]

unreachable:                                      ; preds = %post.tail
  unreachable
}

define i8 @allocaTailRegression(i8* %0, i8* %parentHandle) {
entry:
  %a = alloca i8, align 1
  %coro.id = call token @llvm.coro.id(i32 0, i8* null, i8* null, i8* null)
  %coro.size = call i32 @llvm.coro.size.i32()
  %coro.alloc = call i8* @runtime.alloc(i32 %coro.size, i8* null, i8* undef, i8* undef)
  %coro.state = call i8* @llvm.coro.begin(token %coro.id, i8* %coro.alloc)
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  %task.state.parent = call i8* @"(*internal/task.Task).setState"(%"internal/task.Task"* %task.current, i8* %coro.state, i8* undef, i8* undef)
  %task.retPtr = call i8* @"(*internal/task.Task).getReturnPtr"(%"internal/task.Task"* %task.current, i8* undef, i8* undef)
  call void @sleep(i64 1000000, i8* undef, i8* %parentHandle)
  %coro.save1 = call token @llvm.coro.save(i8* %coro.state)
  %call.suspend2 = call i8 @llvm.coro.suspend(token %coro.save1, i1 false)
  switch i8 %call.suspend2, label %suspend [
    i8 0, label %wakeup
    i8 1, label %cleanup
  ]

wakeup:                                           ; preds = %entry
  store i8 5, i8* %a, align 1
  %1 = call i8* @"(*internal/task.Task).setState"(%"internal/task.Task"* %task.current, i8* %task.state.parent, i8* undef, i8* undef)
  call void @"(*internal/task.Task).setReturnPtr"(%"internal/task.Task"* %task.current, i8* %task.retPtr, i8* undef, i8* undef)
  %2 = call i8 @usePtr(i8* %a, i8* undef, i8* %parentHandle)
  br label %post.tail

suspend:                                          ; preds = %entry, %post.tail, %cleanup
  %unused = call i1 @llvm.coro.end(i8* %coro.state, i1 false)
  ret i8 undef

cleanup:                                          ; preds = %entry, %post.tail
  %coro.memFree = call i8* @llvm.coro.free(token %coro.id, i8* %coro.state)
  call void @runtime.free(i8* %coro.memFree, i8* undef, i8* undef)
  br label %suspend

post.tail:                                        ; preds = %wakeup
  %coro.save = call token @llvm.coro.save(i8* %coro.state)
  %call.suspend = call i8 @llvm.coro.suspend(token %coro.save, i1 false)
  switch i8 %call.suspend, label %suspend [
    i8 0, label %unreachable
    i8 1, label %cleanup
  ]

unreachable:                                      ; preds = %post.tail
  unreachable
}

define i8 @usePtr(i8* %0, i8* %1, i8* %parentHandle) {
entry:
  %coro.id = call token @llvm.coro.id(i32 0, i8* null, i8* null, i8* null)
  %coro.size = call i32 @llvm.coro.size.i32()
  %coro.alloc = call i8* @runtime.alloc(i32 %coro.size, i8* null, i8* undef, i8* undef)
  %coro.state = call i8* @llvm.coro.begin(token %coro.id, i8* %coro.alloc)
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  %task.state.parent = call i8* @"(*internal/task.Task).setState"(%"internal/task.Task"* %task.current, i8* %coro.state, i8* undef, i8* undef)
  %task.retPtr = call i8* @"(*internal/task.Task).getReturnPtr"(%"internal/task.Task"* %task.current, i8* undef, i8* undef)
  call void @sleep(i64 1000000, i8* undef, i8* %parentHandle)
  %coro.save = call token @llvm.coro.save(i8* %coro.state)
  %call.suspend = call i8 @llvm.coro.suspend(token %coro.save, i1 false)
  switch i8 %call.suspend, label %suspend [
    i8 0, label %wakeup
    i8 1, label %cleanup
  ]

wakeup:                                           ; preds = %entry
  %2 = load i8, i8* %0, align 1
  store i8 %2, i8* %task.retPtr, align 1
  call void @"(*internal/task.Task).returnTo"(%"internal/task.Task"* %task.current, i8* %task.state.parent, i8* undef, i8* undef)
  br label %cleanup

suspend:                                          ; preds = %entry, %cleanup
  %unused = call i1 @llvm.coro.end(i8* %coro.state, i1 false)
  ret i8 undef

cleanup:                                          ; preds = %entry, %wakeup
  %coro.memFree = call i8* @llvm.coro.free(token %coro.id, i8* %coro.state)
  call void @runtime.free(i8* %coro.memFree, i8* undef, i8* undef)
  br label %suspend
}

define void @sleepGoroutine(i8* %0, i8* %parentHandle) {
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  call void @sleep(i64 1000000, i8* undef, i8* %parentHandle)
  ret void
}

define void @progMain(i8* %0, i8* %parentHandle) {
entry:
  %task.current = bitcast i8* %parentHandle to %"internal/task.Task"*
  call void @doNothing(i8* undef, i8* undef)
  %start.task = call %"internal/task.Task"* @"internal/task.createTask"(i8* undef, i8* undef)
  %start.task.bitcast = bitcast %"internal/task.Task"* %start.task to i8*
  call void @sleepGoroutine(i8* undef, i8* %start.task.bitcast)
  call void @sleep(i64 2000000, i8* undef, i8* %parentHandle)
  ret void
}

define void @main() {
entry:
  %start.task = call %"internal/task.Task"* @"internal/task.createTask"(i8* undef, i8* undef)
  %start.task.bitcast = bitcast %"internal/task.Task"* %start.task to i8*
  call void @progMain(i8* undef, i8* %start.task.bitcast)
  call void @runtime.scheduler(i8* undef, i8* null)
  ret void
}

; Function Attrs: argmemonly nounwind readonly
declare token @llvm.coro.id(i32, i8* readnone, i8* nocapture readonly, i8*) #0

; Function Attrs: nounwind readnone
declare i32 @llvm.coro.size.i32() #1

; Function Attrs: nounwind
declare i8* @llvm.coro.begin(token, i8* writeonly) #2

; Function Attrs: nounwind
declare i8 @llvm.coro.suspend(token, i1) #2

; Function Attrs: nounwind
declare i1 @llvm.coro.end(i8*, i1) #2

; Function Attrs: argmemonly nounwind readonly
declare i8* @llvm.coro.free(token, i8* nocapture readonly) #0

; Function Attrs: nounwind
declare token @llvm.coro.save(i8*) #2

; Function Attrs: argmemonly nounwind willreturn
declare void @llvm.lifetime.start.p0i8(i64 immarg, i8* nocapture) #3

; Function Attrs: argmemonly nounwind willreturn
declare void @llvm.lifetime.end.p0i8(i64 immarg, i8* nocapture) #3

attributes #0 = { argmemonly nounwind readonly }
attributes #1 = { nounwind readnone }
attributes #2 = { nounwind }
attributes #3 = { argmemonly nounwind willreturn }
