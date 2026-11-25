target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare void @externalCall(ptr)

@gep.global = global i8 0
@udiv.global = global i8 0
@udiv.result = global i64 0
@add.global = global i8 0
@add.result = global i64 0
@sub.global = global i8 0
@sub.result = global i64 0
@xor.global = global i8 0
@xor.result = global i64 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @gep.init(ptr undef)
  call void @udiv.init(ptr undef)
  call void @add.init(ptr undef)
  call void @sub.init(ptr undef)
  call void @xor.init(ptr undef)
  ret void
}

define internal void @gep.init(ptr %context) unnamed_addr {
  ; An integer derived from a pointer can't be used as a GEP index at compile
  ; time, even when nothing interpretable follows it.
  %base = inttoptr i64 4096 to ptr
  %offset = ptrtoint ptr @gep.global to i64
  %addr = getelementptr i8, ptr %base, i64 %offset
  call void @externalCall(ptr %addr)
  ret void
}

define internal void @udiv.init(ptr %context) unnamed_addr {
  ; Dividing by an integer derived from a pointer must not divide by zero.
  %n = ptrtoint ptr @udiv.global to i64
  %q = udiv i64 4096, %n
  store i64 %q, ptr @udiv.result
  ret void
}

define internal void @add.init(ptr %context) unnamed_addr {
  ; Constant add on a pointer-derived integer is left for runtime.
  store i64 add (i64 ptrtoint (ptr @add.global to i64), i64 2), ptr @add.result
  ret void
}

define internal void @sub.init(ptr %context) unnamed_addr {
  ; Constant sub on a pointer-derived integer is left for runtime.
  store i64 sub (i64 ptrtoint (ptr @sub.global to i64), i64 2), ptr @sub.result
  ret void
}

define internal void @xor.init(ptr %context) unnamed_addr {
  ; Constant xor on a pointer-derived integer is left for runtime.
  store i64 xor (i64 ptrtoint (ptr @xor.global to i64), i64 2), ptr @xor.result
  ret void
}
