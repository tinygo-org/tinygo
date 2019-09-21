target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

declare void @runtime.printuint8(i8) local_unnamed_addr

declare void @runtime.printint16(i16) local_unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}

define void @main() unnamed_addr {
entry:
  call void @runtime.printuint8(i8 3)
  call void @runtime.printuint8(i8 3)
  call void @runtime.printint16(i16 5)
  call void @runtime.printint16(i16 5)
  ret void
}
