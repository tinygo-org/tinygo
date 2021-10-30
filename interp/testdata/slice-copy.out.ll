target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@"main$alloc.1" = internal unnamed_addr constant [6 x i8] c"\05\00{\00\00\04", align 8

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
  %int16SliceDst.val = load i16, i16* bitcast ([6 x i8]* @"main$alloc.1" to i16*), align 2
  call void @runtime.printint16(i16 %int16SliceDst.val)
  ret void
}
