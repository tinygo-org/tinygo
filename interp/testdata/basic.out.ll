; ModuleID = 'testdata/basic.ll'
source_filename = "testdata/basic.ll"
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = internal global i64 0
@main.nonConst1 = global [4 x i64] zeroinitializer
@main.nonConst2 = global i64 0
@main.someArray = global [8 x { i16, i32 }] zeroinitializer
@main.exportedValue = global [1 x i16*] [i16* @main.exposedValue1]
@main.exportedConst = constant i64 42
@main.exposedValue1 = global i16 0
@main.exposedValue2 = global i16 0
@main.insertedValue = global { i8, i32, { float, { i64, i16 } } } zeroinitializer

declare void @runtime.printint64(i64) unnamed_addr

declare void @runtime.printnl() unnamed_addr

define void @runtime.initAll(i8* %context) unnamed_addr {
interpreted:
  store i64 3, i64* @main.v1, align 8
  call void @runtime.printint64(i64 5)
  call void @runtime.printnl()
  %0 = call i64 @someValue()
  store i64 %0, i64* getelementptr inbounds ([4 x i64], [4 x i64]* @main.nonConst1, i64 0, i64 0), align 16
  store i64 %0, i64* @main.nonConst2, align 8
  call void @modifyExternal(i32* getelementptr inbounds ([8 x { i16, i32 }], [8 x { i16, i32 }]* @main.someArray, i64 0, i64 3, i32 1))
  call void @modifyExternal(i32* bitcast ([1 x i16*]* @main.exportedValue to i32*))
  store i16 5, i16* @main.exposedValue1, align 2
  call void @readExternal(i32* bitcast (i64* @main.exportedConst to i32*))
  call void @runtime.printint64(i64 42)
  call void @modifyExternal(i32* bitcast (void ()* @willModifyGlobal to i32*))
  store i16 7, i16* @main.exposedValue2, align 2
  call void @modifyExternal(i32* bitcast (void ()* @hasInlineAsm to i32*))
  call void @runtime.printint64(i64 6)
  call void @runtime.printint64(i64 -1)
  ret void
}

define void @main() unnamed_addr {
entry:
  %0 = load i64, i64* @main.v1, align 8
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
  store i64 3, i64* @main.v1, align 8
  call void @"main.init#1"()
  %value1 = call i64 @someValue()
  store i64 %value1, i64* getelementptr inbounds ([4 x i64], [4 x i64]* @main.nonConst1, i64 0, i64 0), align 16
  store i64 %value1, i64* @main.nonConst2, align 8
  call void @modifyExternal(i32* getelementptr inbounds ([8 x { i16, i32 }], [8 x { i16, i32 }]* @main.someArray, i64 0, i64 3, i32 1))
  call void @modifyExternal(i32* bitcast ([1 x i16*]* @main.exportedValue to i32*))
  store i16 5, i16* @main.exposedValue1, align 2
  call void @readExternal(i32* bitcast (i64* @main.exportedConst to i32*))
  call void @runtime.printint64(i64 42)
  call void @modifyExternal(i32* bitcast (void ()* @willModifyGlobal to i32*))
  store i16 7, i16* @main.exposedValue2, align 2
  call void @modifyExternal(i32* bitcast (void ()* @hasInlineAsm to i32*))
  %switch1 = call i64 @testSwitch(i64 1)
  %switch2 = call i64 @testSwitch(i64 9)
  call void @runtime.printint64(i64 %switch1)
  call void @runtime.printint64(i64 %switch2)
  ret void
}

define internal void @"main.init#1"() unnamed_addr {
entry:
  call void @runtime.printint64(i64 5)
  call void @runtime.printnl()
  ret void
}

declare i64 @someValue()

declare void @modifyExternal(i32*)

declare void @readExternal(i32*)

define void @willModifyGlobal() {
entry:
  store i16 8, i16* @main.exposedValue2, align 2
  ret void
}

define void @hasInlineAsm() {
entry:
  call void asm sideeffect "", ""() #0
  ret void
}

define i64 @testSwitch(i64 %val) {
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

declare { i8, i32, { float, { i64, i16 } } } @nestedStruct()

attributes #0 = { nounwind }
