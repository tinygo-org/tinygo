target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

%frame = type { ptr, ptr, i32, i32, ptr, ptr, ptr }
%record = type { ptr, ptr }

@constant.counter = local_unnamed_addr global i32 41
@constant.result = local_unnamed_addr global i32 42
@external.counter = global i32 0
@external.result = local_unnamed_addr global i32 0
@nested.counter = global i32 0
@nested.result = local_unnamed_addr global i32 0
@panic.counter = global i32 0
@panic.result = local_unnamed_addr global i32 0

declare ptr @llvm.stacksave.p0() #0

declare void @runtime.setupDeferFrame(ptr, ptr, ptr) local_unnamed_addr

declare void @runtime.destroyDeferFrame(ptr, ptr) local_unnamed_addr

declare void @runtime._panic(ptr, ptr, ptr) local_unnamed_addr

declare void @external() local_unnamed_addr

define void @runtime.initAll() local_unnamed_addr {
  %external = call fastcc i32 @deferred(ptr @external.counter, ptr @externalBody)
  store i32 %external, ptr @external.result, align 4
  %nested = call fastcc i32 @deferred(ptr @nested.counter, ptr @nestedBody)
  store i32 %nested, ptr @nested.result, align 4
  %panic = call fastcc i32 @deferred(ptr @panic.counter, ptr @panicBody)
  store i32 %panic, ptr @panic.result, align 4
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

define internal fastcc void @callback(ptr %counter) unnamed_addr {
  %old = load i32, ptr %counter, align 4
  %next = add i32 %old, 1
  store i32 %next, ptr %counter, align 4
  ret void
}

define internal fastcc i32 @deferred(ptr %counter, ptr %bodyFn) unnamed_addr {
entry:
  %frame = alloca %frame, align 8
  %head = getelementptr %frame, ptr %frame, i32 0, i32 6
  %sp = call ptr @llvm.stacksave.p0()
  call void @runtime.setupDeferFrame(ptr %frame, ptr %sp, ptr undef)
  %record = alloca %record, align 8
  %prev = load ptr, ptr %head, align 8
  store ptr %prev, ptr %record, align 8
  %argument = getelementptr %record, ptr %record, i32 0, i32 1
  store ptr %counter, ptr %argument, align 8
  store ptr %record, ptr %head, align 8
  %checkpoint = call i32 asm "", "=r,r"(ptr %frame) #1
  %normal = icmp eq i32 %checkpoint, 0
  br i1 %normal, label %body, label %defers

body:                                             ; preds = %entry
  store i32 40, ptr %counter, align 4
  call void %bodyFn()
  br label %defers

defers:                                           ; preds = %callback, %body, %entry
  %current = load ptr, ptr %head, align 8
  %done = icmp eq ptr %current, null
  br i1 %done, label %return, label %callback

callback:                                         ; preds = %defers
  %next = load ptr, ptr %current, align 8
  store ptr %next, ptr %head, align 8
  %argptr = getelementptr %record, ptr %current, i32 0, i32 1
  %arg = load ptr, ptr %argptr, align 8
  call fastcc void @callback(ptr %arg)
  br label %defers

return:                                           ; preds = %defers
  call void @runtime.destroyDeferFrame(ptr %frame, ptr undef)
  ret i32 42
}

attributes #0 = { nocallback nofree nosync nounwind willreturn }
attributes #1 = { returns_twice }
