target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.nonConst1 = local_unnamed_addr global [4 x i64] zeroinitializer
@main.nonConst2 = local_unnamed_addr global i64 0
@main.someArray = global [8 x { i16, i32 }] zeroinitializer
@main.exportedValue = global [1 x i16*] [i16* @main.exposedValue1]
@main.exposedValue1 = global i16 0
@main.exposedValue2 = local_unnamed_addr global i16 0

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
  call void @modifyExternal(i32* getelementptr inbounds ([8 x { i16, i32 }], [8 x { i16, i32 }]* @main.someArray, i32 0, i32 3, i32 1))
  call void @modifyExternal(i32* bitcast ([1 x i16*]* @main.exportedValue to i32*))
  store i16 5, i16* @main.exposedValue1
  call void @modifyExternal(i32* bitcast (void ()* @willModifyGlobal to i32*))
  store i16 7, i16* @main.exposedValue2
  call void @runtime.printint64(i64 6)
  call void @runtime.printint64(i64 -1)
  ret void
}

define void @main() unnamed_addr {
entry:
  call void @runtime.printint64(i64 3)
  call void @runtime.printnl()
  ret void
}

declare i64 @someValue() local_unnamed_addr

declare void @modifyExternal(i32*) local_unnamed_addr

define void @willModifyGlobal() {
entry:
  store i16 8, i16* @main.exposedValue2
  ret void
}

define i64 @testSwitch(i64 %val) local_unnamed_addr {
entry:
  switch i64 %val, label %otherwise [
    i64 0, label %zero
    i64 1, label %one
    i64 2, label %two
  ]

zero:                                             ; preds = %entry
  ret i64 5

one:                                              ; preds = %entry
  ret i64 6

two:                                              ; preds = %entry
  ret i64 7

otherwise:                                        ; preds = %entry
  ret i64 -1
}
