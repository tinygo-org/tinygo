target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@gep.global = global i8 0
@udiv.global = global i8 0
@udiv.result = local_unnamed_addr global i64 0

declare void @externalCall(ptr) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call fastcc void @gep.init(ptr undef)
  call fastcc void @udiv.init(ptr undef)
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
