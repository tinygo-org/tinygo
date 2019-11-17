target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.nonConst1 = local_unnamed_addr global [4 x i64] zeroinitializer
@main.nonConst2 = local_unnamed_addr global i64 0

declare void @runtime.printint64(i64) unnamed_addr

declare void @runtime.printnl() unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @runtime.printint64(i64 5)
  call void @runtime.printnl()
  %value1 = call i64 @someValue()
  store i64 %value1, i64* getelementptr inbounds ([4 x i64], [4 x i64]* @main.nonConst1, i32 0, i32 0)
  %value2 = load i64, i64* getelementptr inbounds ([4 x i64], [4 x i64]* @main.nonConst1, i32 0, i32 0)
  store i64 %value2, i64* @main.nonConst2
  ret void
}

define void @main() unnamed_addr {
entry:
  call void @runtime.printint64(i64 3)
  call void @runtime.printnl()
  ret void
}

declare i64 @someValue() local_unnamed_addr
