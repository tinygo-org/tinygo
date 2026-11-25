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
@nestedpanic.counter = global i32 0
@nestedpanic.result = global i32 0
@deferredpanic.counter = global i32 0
@deferredpanic.result = global i32 0

declare ptr @llvm.stacksave.p0()
declare void @runtime.setupDeferFrame(ptr, ptr, ptr)
declare void @runtime.destroyDeferFrame(ptr, ptr)
declare void @runtime._panic(ptr, ptr, ptr)
declare { ptr, ptr } @runtime._recover(i1, ptr)
declare void @external()

define void @runtime.initAll() {
  call void @main.init()
  ret void
}

define internal void @main.init() {
  %constant = call i32 @deferred(ptr @constant.counter, ptr @constantBody, ptr @callback)
  store i32 %constant, ptr @constant.result
  %external = call i32 @deferred(ptr @external.counter, ptr @externalBody, ptr @callback)
  store i32 %external, ptr @external.result
  %nested = call i32 @deferred(ptr @nested.counter, ptr @nestedBody, ptr @callback)
  store i32 %nested, ptr @nested.result
  %panic = call i32 @deferred(ptr @panic.counter, ptr @panicBody, ptr @callback)
  store i32 %panic, ptr @panic.result
  %nestedpanic = call i32 @deferred(ptr @nestedpanic.counter, ptr @nestedPanicBody, ptr @callback)
  store i32 %nestedpanic, ptr @nestedpanic.result
  %deferredpanic = call i32 @deferred(ptr @deferredpanic.counter, ptr @constantBody, ptr @panicCallback)
  store i32 %deferredpanic, ptr @deferredpanic.result
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

define internal void @nestedPanicBody() {
  call void @panicBody()
  ret void
}

define internal void @callback(ptr %counter) {
  %recovered = call { ptr, ptr } @runtime._recover(i1 false, ptr null)
  %type = extractvalue { ptr, ptr } %recovered, 0
  %value = extractvalue { ptr, ptr } %recovered, 1
  %nilType = icmp eq ptr %type, null
  %nilValue = icmp eq ptr %value, null
  %nil = and i1 %nilType, %nilValue
  %parent = call { ptr, ptr } @runtime._recover(i1 true, ptr null)
  %parentType = extractvalue { ptr, ptr } %parent, 0
  %parentValue = extractvalue { ptr, ptr } %parent, 1
  %nilParentType = icmp eq ptr %parentType, null
  %nilParentValue = icmp eq ptr %parentValue, null
  %nilParent = and i1 %nilParentType, %nilParentValue
  %bothNil = and i1 %nil, %nilParent
  %increment = select i1 %bothNil, i32 1, i32 100
  %old = load i32, ptr %counter
  %next = add i32 %old, %increment
  store i32 %next, ptr %counter
  ret void
}

define internal void @panicCallback(ptr %counter) {
  call void @callback(ptr %counter)
  call void @panicBody()
  ret void
}

define internal i32 @deferred(ptr %counter, ptr %bodyFn, ptr %callbackFn) {
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
  call void %callbackFn(ptr %arg)
  br label %defers

return:
  call void @runtime.destroyDeferFrame(ptr %frame, ptr undef)
  ret i32 42
}
