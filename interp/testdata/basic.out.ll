target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.nonConst1 = local_unnamed_addr global [4 x i64] zeroinitializer
@main.nonConst2 = local_unnamed_addr global i64 0
@main.someArray = global [8 x { i16, i32 }] zeroinitializer
@main.exportedValue = global [1 x ptr] [ptr @main.exposedValue1]
@main.exportedConst = constant i64 42
@main.exposedValue1 = global i16 0
@main.exposedValue2 = local_unnamed_addr global i16 0
@main.insertedValue = local_unnamed_addr global { i8, i32, { float, { i64, i16 } } } zeroinitializer

declare void @runtime.printint64(i64) unnamed_addr

declare void @runtime.printnl() unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @runtime.printint64(i64 5)
  call void @runtime.printnl()
  %value1 = call i64 @someValue()
  store i64 %value1, ptr @main.nonConst1, align 8
  %value2 = load i64, ptr @main.nonConst1, align 8
  store i64 %value2, ptr @main.nonConst2, align 8
  call void @modifyExternal(ptr getelementptr inbounds (i8, ptr @main.someArray, i32 28))
  call void @modifyExternal(ptr @main.exportedValue)
  store i16 5, ptr @main.exposedValue1, align 2
  call void @readExternal(ptr @main.exportedConst)
  call void @runtime.printint64(i64 42)
  call void @modifyExternal(ptr @willModifyGlobal)
  store i16 7, ptr @main.exposedValue2, align 2
  call void @modifyExternal(ptr @hasInlineAsm)
  call void @runtime.printint64(i64 6)
  call void @runtime.printint64(i64 -1)
  %agg = call { i8, i32, { float, { i64, i16 } } } @nestedStruct()
  %elt.agg = extractvalue { i8, i32, { float, { i64, i16 } } } %agg, 2
  %elt.agg1 = extractvalue { float, { i64, i16 } } %elt.agg, 1
  %elt = extractvalue { i64, i16 } %elt.agg1, 0
  call void @runtime.printint64(i64 %elt)
  %agg2.agg0 = extractvalue { i8, i32, { float, { i64, i16 } } } %agg, 2
  %agg2.agg1 = extractvalue { float, { i64, i16 } } %agg2.agg0, 1
  %agg2.insertvalue2 = insertvalue { i64, i16 } %agg2.agg1, i64 5, 0
  %agg2.insertvalue1 = insertvalue { float, { i64, i16 } } %agg2.agg0, { i64, i16 } %agg2.insertvalue2, 1
  %agg2.insertvalue0 = insertvalue { i8, i32, { float, { i64, i16 } } } %agg, { float, { i64, i16 } } %agg2.insertvalue1, 2
  store { i8, i32, { float, { i64, i16 } } } %agg2.insertvalue0, ptr @main.insertedValue, align 8
  ret void
}

define void @main() unnamed_addr {
entry:
  call void @runtime.printint64(i64 3)
  call void @runtime.printnl()
  ret void
}

declare i64 @someValue() local_unnamed_addr

declare void @modifyExternal(ptr) local_unnamed_addr

declare void @readExternal(ptr) local_unnamed_addr

define void @willModifyGlobal() {
entry:
  store i16 8, ptr @main.exposedValue2, align 2
  ret void
}

define void @hasInlineAsm() {
entry:
  call void asm sideeffect "", ""()
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

declare { i8, i32, { float, { i64, i16 } } } @nestedStruct() local_unnamed_addr
