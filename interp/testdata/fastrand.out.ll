target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@runtime.xorshift32State = local_unnamed_addr global i32 1
@runtime.xorshift64State = local_unnamed_addr global i64 1
@foo.rand32 = local_unnamed_addr global i32 99009
@foo.rand64 = local_unnamed_addr global i64 5180492295206395165

define void @runtime.initAll() unnamed_addr {
entry:
  call fastcc void @external.init(ptr undef)
  ret void
}

define internal fastcc void @external.init(ptr %context) unnamed_addr {
  unreachable
}
