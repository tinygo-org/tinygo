target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@ext.ptr = external global ptr
@ext.flag = external global i1
@gep.result = global ptr null
@gep.after = global i32 0
@select.result = global i32 0
@select.after = global i32 0
@ext.idx = external global i64
@buf = global [4 x i32] zeroinitializer
@gep.global.result = global i32 0
@a = global i32 0
@b = global i32 0
@select.global.result = global i32 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @gep.init(ptr undef)
  call void @select.init(ptr undef)
  call void @gep.global.init(ptr undef)
  call void @select.global.init(ptr undef)
  ret void
}

define internal void @gep.init(ptr %context) unnamed_addr {
  ; A GEP on a pointer only known at runtime is emitted without inbounds.
  %base = load ptr, ptr @ext.ptr
  %addr = getelementptr inbounds i8, ptr %base, i64 4
  store ptr %addr, ptr @gep.result
  store i32 1, ptr @gep.after
  ret void
}

define internal void @select.init(ptr %context) unnamed_addr {
  ; A select on a condition only known at runtime is emitted.
  %flag = load i1, ptr @ext.flag
  %value = select i1 %flag, i32 3, i32 4
  store i32 %value, ptr @select.result
  store i32 1, ptr @select.after
  ret void
}

define internal void @gep.global.init(ptr %context) unnamed_addr {
  ; A store through a runtime GEP of a global may change that global.
  %idx = load i64, ptr @ext.idx
  %p = getelementptr [4 x i32], ptr @buf, i64 0, i64 %idx
  store i32 7, ptr %p
  %v = load i32, ptr @buf
  store i32 %v, ptr @gep.global.result
  ret void
}

define internal void @select.global.init(ptr %context) unnamed_addr {
  ; A store through a runtime select of globals may change either global.
  %flag = load i1, ptr @ext.flag
  %p = select i1 %flag, ptr @a, ptr @b
  store i32 7, ptr %p
  %v = load i32, ptr @a
  store i32 %v, ptr @select.global.result
  ret void
}
