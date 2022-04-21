; ModuleID = 'testdata/side.ll'
source_filename = "testdata/side.ll"
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@attrs.var1 = internal global i32 1
@attrs.var2 = global i32 0
@attrs.selfRef = internal global i8* bitcast (i8** @attrs.selfRef to i8*)
@atomicNoEscape.cond = internal global i32 1
@task.done = internal global i32 0
@task.bgData = internal global i64 1
@somehw.intCond = internal global i32 0
@somehw.intData = internal global i64 0

define void @main() local_unnamed_addr {
entry:
  call void @runtime.initAll(i8* undef)
  call void @main.main(i8* undef)
  ret void
}

define internal void @runtime.initAll(i8* %context) unnamed_addr {
interpreted:
  store i32 2, i32* @attrs.var2, align 4
  call void @attrs.nop()
  store i32 2, i32* @attrs.var2, align 4
  call void @attrs.readOnlyNop()
  store i32 2, i32* @attrs.var2, align 4
  call void @attrs.nopWithArg(i32* @attrs.var1)
  store i32 2, i32* @attrs.var2, align 4
  call void @attrs.nop()
  store i32 2, i32* @attrs.var2, align 4
  call void @attrs.readArg(i32* @attrs.var1)
  store i32 1, i32* @attrs.var1, align 4
  store i32 2, i32* @attrs.var2, align 4
  call void @attrs.modifyArg(i32* @attrs.var1)
  %0 = load i32, i32* @attrs.var1, align 4
  store i32 %0, i32* @attrs.var1, align 4
  store i32 2, i32* @attrs.var2, align 4
  call void @attrs.captureNoModifyArg(i32* @attrs.var1)
  %1 = load i32, i32* @attrs.var2, align 4
  store i32 %0, i32* @attrs.var1, align 4
  call void @attrs.nop()
  call void @attrs.nopWithArg(i32* bitcast (i8** @attrs.selfRef to i32*))
  call void @attrs.captureNoModifyArg(i32* bitcast (i8** @attrs.selfRef to i32*))
  %2 = load i8*, i8** @attrs.selfRef, align 8
  call void @start(i8* bitcast (void (i8*)* @task.bgStore to i8*), i8* undef)
  call void @spinWait(i32* @task.done)
  %3 = load i64, i64* @task.bgData, align 8
  call void @print64(i64 %3)
  store i32 0, i32* @somehw.intCond, align 4
  store i64 0, i64* @somehw.intData, align 8
  call void @somehw.enableInterrupt()
  call void @spinWait(i32* @somehw.intCond)
  %4 = load i64, i64* @somehw.intData, align 8
  call void @print64(i64 %4)
  ret void
}

define internal void @main.main(i8* %context) unnamed_addr {
entry:
  ret void
}

define internal void @attrs.init() {
  store i32 1, i32* @attrs.var1, align 4
  store i32 2, i32* @attrs.var2, align 4
  call void @attrs.nop()
  %var1.0 = load i32, i32* @attrs.var1, align 4
  %var2.0 = load i32, i32* @attrs.var2, align 4
  store i32 %var1.0, i32* @attrs.var1, align 4
  store i32 %var2.0, i32* @attrs.var2, align 4
  call void @attrs.readOnlyNop()
  %var1.1 = load i32, i32* @attrs.var1, align 4
  %var2.1 = load i32, i32* @attrs.var2, align 4
  store i32 %var1.1, i32* @attrs.var1, align 4
  store i32 %var2.1, i32* @attrs.var2, align 4
  call void @attrs.nopWithArg(i32* @attrs.var1)
  %var1.2 = load i32, i32* @attrs.var1, align 4
  %var2.2 = load i32, i32* @attrs.var2, align 4
  store i32 %var1.2, i32* @attrs.var1, align 4
  store i32 %var2.2, i32* @attrs.var2, align 4
  call void @attrs.nop()
  %var1.3 = load i32, i32* @attrs.var1, align 4
  %var2.3 = load i32, i32* @attrs.var2, align 4
  store i32 %var1.3, i32* @attrs.var1, align 4
  store i32 %var2.3, i32* @attrs.var2, align 4
  call void @attrs.readArg(i32* @attrs.var1)
  %var1.4 = load i32, i32* @attrs.var1, align 4
  %var2.4 = load i32, i32* @attrs.var2, align 4
  store i32 %var1.4, i32* @attrs.var1, align 4
  store i32 %var2.4, i32* @attrs.var2, align 4
  call void @attrs.modifyArg(i32* @attrs.var1)
  %var1.5 = load i32, i32* @attrs.var1, align 4
  %var2.5 = load i32, i32* @attrs.var2, align 4
  store i32 %var1.5, i32* @attrs.var1, align 4
  store i32 %var2.5, i32* @attrs.var2, align 4
  call void @attrs.captureNoModifyArg(i32* @attrs.var1)
  %var1.6 = load i32, i32* @attrs.var1, align 4
  %var2.6 = load i32, i32* @attrs.var2, align 4
  store i32 %var1.6, i32* @attrs.var1, align 4
  call void @attrs.nop()
  %var1.7 = load i32, i32* @attrs.var1, align 4
  call void @attrs.nopWithArg(i32* bitcast (i8** @attrs.selfRef to i32*))
  %self = load i8*, i8** @attrs.selfRef, align 8
  store i8* %self, i8** @attrs.selfRef, align 8
  call void @attrs.captureNoModifyArg(i32* bitcast (i8** @attrs.selfRef to i32*))
  %unknown = load i8*, i8** @attrs.selfRef, align 8
  ret void
}

; Function Attrs: nosync readnone
declare void @attrs.nop() #0

; Function Attrs: nosync readonly
declare void @attrs.readOnlyNop() #1

; Function Attrs: nosync readnone
declare void @attrs.nopWithArg(i32* nocapture) #0

; Function Attrs: argmemonly nosync
declare void @attrs.readArg(i32* nocapture readonly) #2

; Function Attrs: argmemonly nosync
declare void @attrs.modifyArg(i32* nocapture) #2

declare void @attrs.captureNoModifyArg(i32* readonly)

define internal void @atomicNoEscape.init() {
  call void @spinNotify(i32* @atomicNoEscape.cond)
  call void @spinWait(i32* @atomicNoEscape.cond)
  ret void
}

define internal void @task.init() {
  store i64 1, i64* @task.bgData, align 8
  call void @start(i8* bitcast (void (i8*)* @task.bgStore to i8*), i8* undef)
  call void @spinWait(i32* @task.done)
  %data = load i64, i64* @task.bgData, align 8
  call void @print64(i64 %data)
  ret void
}

define internal void @task.bgStore(i8* %context) {
  store i64 7, i64* @task.bgData, align 8
  call void @spinNotify(i32* @task.done)
  ret void
}

declare void @start(i8*, i8*)

define internal void @somehw.init() unnamed_addr {
  store i32 0, i32* @somehw.intCond, align 4
  store i64 0, i64* @somehw.intData, align 8
  call void @somehw.enableInterrupt()
  %data = call i64 @somehw.waitForInterrupt()
  call void @print64(i64 %data)
  ret void
}

declare void @somehw.enableInterrupt()

define void @handleInterrupt(i64 %data) {
  store i64 %data, i64* @somehw.intData, align 8
  call void @spinNotify(i32* @somehw.intCond)
  ret void
}

define internal i64 @somehw.waitForInterrupt() {
  call void @spinWait(i32* @somehw.intCond)
  %data = load i64, i64* @somehw.intData, align 8
  ret i64 %data
}

define internal void @spinWait(i32* %on) {
entry:
  br label %wait

wait:                                             ; preds = %wait, %entry
  %load = load atomic i32, i32* %on seq_cst, align 4
  %ok = trunc i32 %load to i1
  br i1 %ok, label %exit, label %wait

exit:                                             ; preds = %wait
  ret void
}

define internal void @spinNotify(i32* %on) {
  store atomic i32 1, i32* %on seq_cst, align 4
  ret void
}

; Function Attrs: nosync readnone
declare void @print32(i32) #0

; Function Attrs: nosync readnone
declare void @print64(i64) #0

; Function Attrs: nosync readnone
declare void @printPtr(i8*) #0

attributes #0 = { nosync readnone }
attributes #1 = { nosync readonly }
attributes #2 = { argmemonly nosync }
