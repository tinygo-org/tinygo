target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = internal global i64 0

declare void @runtime.printint64(i64) unnamed_addr

declare void @runtime.printnl() unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @runtime.init()
  call void @main.init()
  ret void
}

define void @main() unnamed_addr {
entry:
  %0 = load i64, i64* @main.v1
  call void @runtime.printint64(i64 %0)
  call void @runtime.printnl()
  ret void
}

define internal void @runtime.init() unnamed_addr {
entry:
  ret void
}

define internal void @main.init() unnamed_addr {
entry:
  store i64 3, i64* @main.v1
  call void @"main.init#1"()
  ret void
}

define internal void @"main.init#1"() unnamed_addr {
entry:
  call void @runtime.printint64(i64 5)
  call void @runtime.printnl()
  ret void
}
