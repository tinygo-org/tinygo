target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare void @runtime.printint64(i64) unnamed_addr

declare void @runtime.printnl() unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @runtime.printint64(i64 5)
  call void @runtime.printnl()
  ret void
}

define void @main() unnamed_addr {
entry:
  call void @runtime.printint64(i64 3)
  call void @runtime.printnl()
  ret void
}
