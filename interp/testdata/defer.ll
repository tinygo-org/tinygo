target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

%frame = type { ptr, ptr, i32, i32, ptr, ptr, ptr }
%record = type { ptr, ptr }

@constant.counter = global i32 0
@constant.result = global i32 0
@external.counter = global i32 0
@external.result = global i32 0
@nested.counter = global i32 0
@nested.result = global i32 0
@panic.counter = global i32 0
@panic.result = global i32 0

declare ptr @llvm.stacksave.p0()
declare void @runtime.setupDeferFrame(ptr, ptr, ptr)
declare void @runtime.destroyDeferFrame(ptr, ptr)
declare void @runtime._panic(ptr, ptr, ptr)
declare void @external()

define void @runtime.initAll() {
  call void @main.init()
  ret void
}

define internal void @main.init() {
  %constant = call i32 @deferred(ptr @constant.counter, ptr @constantBody)
  store i32 %constant, ptr @constant.result
  %external = call i32 @deferred(ptr @external.counter, ptr @externalBody)
  store i32 %external, ptr @external.result
  %nested = call i32 @deferred(ptr @nested.counter, ptr @nestedBody)
  store i32 %nested, ptr @nested.result
  %panic = call i32 @deferred(ptr @panic.counter, ptr @panicBody)
  store i32 %panic, ptr @panic.result
  ret void
}

define internal void @constantBody() {
  ret void
}

define internal void @externalBody() {
  call void @external()
  ret void
}

define internal void @nestedBody() {
  call void @externalBody()
  ret void
}

define internal void @panicBody() {
  call void @runtime._panic(ptr undef, ptr undef, ptr undef)
  unreachable
}

define internal void @callback(ptr %counter) {
  %old = load i32, ptr %counter
  %next = add i32 %old, 1
  store i32 %next, ptr %counter
  ret void
}

define internal i32 @deferred(ptr %counter, ptr %bodyFn) {
entry:
  %frame = alloca %frame
  %head = getelementptr %frame, ptr %frame, i32 0, i32 6
  %sp = call ptr @llvm.stacksave.p0()
  call void @runtime.setupDeferFrame(ptr %frame, ptr %sp, ptr undef)
  %record = alloca %record
  %prev = load ptr, ptr %head
  store ptr %prev, ptr %record
  %argument = getelementptr %record, ptr %record, i32 0, i32 1
  store ptr %counter, ptr %argument
  store ptr %record, ptr %head
  %checkpoint = call i32 asm "", "=r,r"(ptr %frame) returns_twice
  %normal = icmp eq i32 %checkpoint, 0
  br i1 %normal, label %body, label %defers

body:
  store i32 40, ptr %counter
  call void %bodyFn()
  br label %defers

defers:
  %current = load ptr, ptr %head
  %done = icmp eq ptr %current, null
  br i1 %done, label %return, label %callback

callback:
  %next = load ptr, ptr %current
  store ptr %next, ptr %head
  %argptr = getelementptr %record, ptr %current, i32 0, i32 1
  %arg = load ptr, ptr %argptr
  call void @callback(ptr %arg)
  br label %defers

return:
  call void @runtime.destroyDeferFrame(ptr %frame, ptr undef)
  ret i32 42
}
