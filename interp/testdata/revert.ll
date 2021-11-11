target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare void @externalCall(i64)

@foo.knownAtRuntime = global i64 0
@bar.knownAtRuntime = global i64 0
@baz.someGlobal = external global [3 x {i64, i32}]
@baz.someInt = global i32 0

define void @runtime.initAll() unnamed_addr {
entry:
  call void @baz.init(i8* undef, i8* undef)
  call void @foo.init(i8* undef, i8* undef)
  call void @bar.init(i8* undef, i8* undef)
  call void @main.init(i8* undef, i8* undef)
  ret void
}

define internal void @foo.init(i8* %context, i8* %parentHandle) unnamed_addr {
  store i64 5, i64* @foo.knownAtRuntime
  unreachable ; this triggers a revert of @foo.init.
}

define internal void @bar.init(i8* %context, i8* %parentHandle) unnamed_addr {
  %val = load i64, i64* @foo.knownAtRuntime
  store i64 %val, i64* @bar.knownAtRuntime
  ret void
}

define internal void @baz.init(i8* %context, i8* %parentHandle) unnamed_addr {
  ; Test extractvalue/insertvalue with more than one index.
  %val = load [3 x {i64, i32}], [3 x {i64, i32}]* @baz.someGlobal
  %part = extractvalue [3 x {i64, i32}] %val, 0, 1
  %val2 = insertvalue [3 x {i64, i32}] %val, i32 5, 2, 1
  unreachable ; trigger revert
}

define internal void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  call void @externalCall(i64 3)
  ret void
}
