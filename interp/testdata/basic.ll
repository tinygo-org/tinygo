target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = internal global i64 0
@main.nonConst1 = global [4 x i64] zeroinitializer
@main.nonConst2 = global i64 0

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

  ; test the following pattern:
  ;     func someValue() int // extern function
  ;     var nonConst1 = [4]int{someValue(), 0, 0, 0}
  %value1 = call i64 @someValue()
  %gep1 = getelementptr [4 x i64], [4 x i64]* @main.nonConst1, i32 0, i32 0
  store i64 %value1, i64* %gep1

  ; Test that the global really is marked dirty:
  ;     var nonConst2 = nonConst1[0]
  %gep2 = getelementptr [4 x i64], [4 x i64]* @main.nonConst1, i32 0, i32 0
  %value2 = load i64, i64* %gep2
  store i64 %value2, i64* @main.nonConst2

  ret void
}

define internal void @"main.init#1"() unnamed_addr {
entry:
  call void @runtime.printint64(i64 5)
  call void @runtime.printnl()
  ret void
}

declare i64 @someValue()
