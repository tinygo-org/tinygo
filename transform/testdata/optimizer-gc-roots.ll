target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-unknown-wasm"

@enabled = internal global i32 0
@queue = internal global ptr null

declare ptr @runtime.alloc(i32, ptr)
declare void @runtime.free(ptr)
declare void @runtime.nilPanic()
declare i32 @runtime.gcGlobalRootCount()
declare ptr @runtime.gcGlobalRoot(i32)
declare i32 @runtime.gcGlobalRootSize(i32)

define {i32, ptr, i32} @roots(i32 %index) {
entry:
  %count = call i32 @runtime.gcGlobalRootCount()
  %root = call ptr @runtime.gcGlobalRoot(i32 %index)
  %size = call i32 @runtime.gcGlobalRootSize(i32 %index)
  %result0 = insertvalue {i32, ptr, i32} undef, i32 %count, 0
  %result1 = insertvalue {i32, ptr, i32} %result0, ptr %root, 1
  %result2 = insertvalue {i32, ptr, i32} %result1, i32 %size, 2
  ret {i32, ptr, i32} %result2
}

define void @start(ptr %channel) {
entry:
  %slot = alloca ptr
  store ptr null, ptr %slot
  call void @newTimer(ptr %slot)
  ret void
}

define internal void @newTimer(ptr %slot) {
entry:
  %channel = load ptr, ptr %slot
  %isChannel = icmp ne ptr %channel, null
  br i1 %isChannel, label %enable, label %done

enable:
  store i32 1, ptr @enabled
  br label %done

done:
  ret void
}

define void @receive(ptr %timer) {
entry:
  %flag = call i32 @loadFlag(ptr @enabled)
  %disabled = icmp eq i32 %flag, 0
  br i1 %disabled, label %done, label %enqueue

enqueue:
  store ptr %timer, ptr @queue
  br label %done

done:
  ret void
}

define internal i32 @loadFlag(ptr %address) {
entry:
  %flag = load i32, ptr %address
  ret i32 %flag
}

define i32 @readFlag(ptr %address) {
entry:
  %flag = call i32 @loadFlag(ptr %address)
  ret i32 %flag
}

define ptr @poll() {
entry:
  %timer = load ptr, ptr @queue
  ret ptr %timer
}
