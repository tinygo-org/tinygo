target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@gep.global = global i8 0
@udiv.global = global i8 0
@udiv.result = local_unnamed_addr global i64 0
@add.global = global i8 0
@add.result = local_unnamed_addr global i64 0
@sub.global = global i8 0
@sub.result = local_unnamed_addr global i64 0
@xor.global = global i8 0
@xor.result = local_unnamed_addr global i64 0

declare void @externalCall(ptr) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call fastcc void @gep.init(ptr undef)
  call fastcc void @udiv.init(ptr undef)
  store i64 add (i64 ptrtoint (ptr @add.global to i64), i64 2), ptr @add.result, align 8
  store i64 sub (i64 ptrtoint (ptr @sub.global to i64), i64 2), ptr @sub.result, align 8
  store i64 xor (i64 ptrtoint (ptr @xor.global to i64), i64 2), ptr @xor.result, align 8
  ret void
}

define internal fastcc void @gep.init(ptr %context) unnamed_addr {
  %base = inttoptr i64 4096 to ptr
  %offset = ptrtoint ptr @gep.global to i64
  %addr = getelementptr i8, ptr %base, i64 %offset
  call void @externalCall(ptr %addr)
  ret void
}

define internal fastcc void @udiv.init(ptr %context) unnamed_addr {
  %n = ptrtoint ptr @udiv.global to i64
  %q = udiv i64 4096, %n
  store i64 %q, ptr @udiv.result, align 8
  ret void
}
