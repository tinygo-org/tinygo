target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@foo.knownAtRuntime = local_unnamed_addr global i64 0
@bar.knownAtRuntime = local_unnamed_addr global i64 0
@baz.someGlobal = external local_unnamed_addr global [3 x { i64, i32 }]
@baz.someInt = local_unnamed_addr global i32 0

declare void @externalCall(i64) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call fastcc void @baz.init(i8* undef, i8* undef)
  call fastcc void @foo.init(i8* undef, i8* undef)
  %val = load i64, i64* @foo.knownAtRuntime, align 8
  store i64 %val, i64* @bar.knownAtRuntime, align 8
  call void @externalCall(i64 3)
  ret void
}

define internal fastcc void @foo.init(i8* %context, i8* %parentHandle) unnamed_addr {
  store i64 5, i64* @foo.knownAtRuntime, align 8
  unreachable
}

define internal fastcc void @baz.init(i8* %context, i8* %parentHandle) unnamed_addr {
  unreachable
}
