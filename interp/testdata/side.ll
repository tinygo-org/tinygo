target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

define void @main() local_unnamed_addr {
entry:
  call void @runtime.initAll(i8* undef)
  call void @main.main(i8* undef)
  ret void
}

define internal void @runtime.initAll(i8* %context) unnamed_addr {
entry:
  call void @attrs.init()
  call void @atomicNoEscape.init()
  call void @task.init()
  call void @somehw.init()
  ret void
}

define internal void @main.main(i8* %context) unnamed_addr {
entry:
  ret void
}


@attrs.var1 = internal global i32 0
@attrs.var2 = global i32 0
@attrs.selfRef = internal global i8* bitcast (i8** @attrs.selfRef to i8*)

define internal void @attrs.init() {
  ; Test that a function that accesses no memory has no impact on any unescaped variables.
  ; An escaped variable should also be flushed (in case the function longjmps out of init) but not invalidated.
  ; There is currently no attribute strong enough to skip this flush (nounwind does not exclude longjmp).
  store i32 1, i32* @attrs.var1
  store i32 2, i32* @attrs.var2
  call void @attrs.nop()
  %var1.0 = load i32, i32* @attrs.var1
  %var2.0 = load i32, i32* @attrs.var2

  ; Test that a function that only reads external memory does the same.
  store i32 %var1.0, i32* @attrs.var1
  store i32 %var2.0, i32* @attrs.var2
  call void @attrs.readOnlyNop()
  %var1.1 = load i32, i32* @attrs.var1
  %var2.1 = load i32, i32* @attrs.var2

  ; Test that an unused pointer argument does not result in any other behavior.
  store i32 %var1.1, i32* @attrs.var1
  store i32 %var2.1, i32* @attrs.var2
  call void @attrs.nopWithArg(i32* @attrs.var1)
  %var1.2 = load i32, i32* @attrs.var1
  %var2.2 = load i32, i32* @attrs.var2
  store i32 %var1.2, i32* @attrs.var1
  store i32 %var2.2, i32* @attrs.var2
  call void @attrs.nop()
  %var1.3 = load i32, i32* @attrs.var1
  %var2.3 = load i32, i32* @attrs.var2

  ; Test that a function can read an argument without escaping it or accessing other memory.
  ; The behavior with escaped memory is the same as the previous calls.
  store i32 %var1.3, i32* @attrs.var1
  store i32 %var2.3, i32* @attrs.var2
  call void @attrs.readArg(i32* @attrs.var1)
  %var1.4 = load i32, i32* @attrs.var1
  %var2.4 = load i32, i32* @attrs.var2

  ; Test that a function can modify an argument without escaping it or accessing other memory.
  store i32 %var1.4, i32* @attrs.var1
  store i32 %var2.4, i32* @attrs.var2
  call void @attrs.modifyArg(i32* @attrs.var1)
  %var1.5 = load i32, i32* @attrs.var1
  %var2.5 = load i32, i32* @attrs.var2

  ; Test that a capture without modification does not immediately invalidate the captured value.
  store i32 %var1.5, i32* @attrs.var1
  store i32 %var2.5, i32* @attrs.var2
  call void @attrs.captureNoModifyArg(i32* @attrs.var1)
  %var1.6 = load i32, i32* @attrs.var1
  %var2.6 = load i32, i32* @attrs.var2

  ; Verify that the argument was escaped.
  store i32 %var1.6, i32* @attrs.var1
  call void @attrs.nop()
  %var1.7 = load i32, i32* @attrs.var1

  ; Test that @attrs.nopWithArg behaves the same even if the target object contains a pointer.
  call void @attrs.nopWithArg(i32* bitcast (i8** @attrs.selfRef to i32*))
  %self = load i8*, i8** @attrs.selfRef
  store i8* %self, i8** @attrs.selfRef

  ; Test that a capture without direct modification will still invalidate a self-referencing capture.
  ; This is necessary because the referenced value can still be accessed through a reference in the observed memory.
  call void @attrs.captureNoModifyArg(i32* bitcast (i8** @attrs.selfRef to i32*))
  %unknown = load i8*, i8** @attrs.selfRef

  ret void
}

declare void @attrs.nop() readnone nosync
declare void @attrs.readOnlyNop() readonly nosync
declare void @attrs.nopWithArg(i32* nocapture) readnone nosync
declare void @attrs.readArg(i32* nocapture readonly) argmemonly nosync
declare void @attrs.modifyArg(i32* nocapture) argmemonly nosync
declare void @attrs.captureNoModifyArg(i32* readonly)

@atomicNoEscape.cond = internal global i32 0

define internal void @atomicNoEscape.init() {
  ; Non-volatile atomic ops on unescaped memory can be run as regular memory ops.
  call void @spinNotify(i32* @atomicNoEscape.cond)
  call void @spinWait(i32* @atomicNoEscape.cond)
  ret void
}

@task.done = internal global i32 0
@task.bgData = internal global i64 0

define internal void @task.init() {
  ; @task.bgData is not yet escaped, so this store should be initializing.
  store i64 1, i64* @task.bgData

  ; Run a function asynchronously to set bgData.
  call void @start(i8* bitcast (void (i8*)* @task.bgStore to i8*), i8* undef)
  call void @spinWait(i32* @task.done)

  ; Load the result (not known).
  %data = load i64, i64* @task.bgData
  call void @print64(i64 %data)
  ret void
}

define internal void @task.bgStore(i8* %context) {
  store i64 7, i64* @task.bgData
  call void @spinNotify(i32* @task.done)
  ret void
}

declare void @start(i8*, i8*)


@somehw.intCond = internal global i32 0
@somehw.intData = internal global i64 0

define internal void @somehw.init() unnamed_addr {
  ; These stores should be processed immediately, and should be visible.
  ; However, they will be flushed by the next call to enableInterrupt and do not initialize the sources.
  store i32 0, i32* @somehw.intCond
  store i64 0, i64* @somehw.intData
  call void @somehw.enableInterrupt()

  ; The wait loop should revert, and the syncronization should force @intData to be loaded at runtime.
  %data = call i64 @somehw.waitForInterrupt()
  call void @print64(i64 %data)

  ret void
}

declare void @somehw.enableInterrupt()

define void @handleInterrupt(i64 %data) {
  store i64 %data, i64* @somehw.intData
  call void @spinNotify(i32* @somehw.intCond)
  ret void
}

define internal i64 @somehw.waitForInterrupt() {
  call void @spinWait(i32* @somehw.intCond)
  %data = load i64, i64* @somehw.intData
  ret i64 %data
}

define internal void @spinWait(i32* %on) {
entry:
  br label %wait
wait:
  %load = load atomic i32, i32* %on seq_cst, align 4
  %ok = trunc i32 %load to i1
  br i1 %ok, label %exit, label %wait
exit:
  ret void
}

define internal void @spinNotify(i32* %on) {
  store atomic i32 1, i32* %on seq_cst, align 4
  ret void
}

declare void @print32(i32) readnone nosync
declare void @print64(i64) readnone nosync
declare void @printPtr(i8*) readnone nosync
