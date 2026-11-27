target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@ext.ptr = external local_unnamed_addr global ptr
@ext.flag = external local_unnamed_addr global i1
@gep.result = local_unnamed_addr global ptr null
@gep.after = local_unnamed_addr global i32 1
@select.result = local_unnamed_addr global i32 0
@select.after = local_unnamed_addr global i32 1
@ext.idx = external local_unnamed_addr global i64
@buf = local_unnamed_addr global [4 x i32] zeroinitializer
@gep.global.result = local_unnamed_addr global i32 0
@a = local_unnamed_addr global i32 0
@b = local_unnamed_addr global i32 0
@select.global.result = local_unnamed_addr global i32 0

define void @runtime.initAll() unnamed_addr {
entry:
  %base = load ptr, ptr @ext.ptr, align 8
  %addr = getelementptr i8, ptr %base, i64 4
  store ptr %addr, ptr @gep.result, align 8
  %flag = load i1, ptr @ext.flag, align 1
  %value = select i1 %flag, i32 3, i32 4
  store i32 %value, ptr @select.result, align 4
  %idx = load i64, ptr @ext.idx, align 8
  %p = getelementptr [4 x i32], ptr @buf, i64 0, i64 %idx
  store i32 7, ptr %p, align 4
  %v = load i32, ptr @buf, align 4
  store i32 %v, ptr @gep.global.result, align 4
  %flag1 = load i1, ptr @ext.flag, align 1
  %p2 = select i1 %flag1, ptr @a, ptr @b
  store i32 7, ptr %p2, align 4
  %v3 = load i32, ptr @a, align 4
  store i32 %v3, ptr @select.global.result, align 4
  ret void
}
