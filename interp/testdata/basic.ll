target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = internal global i64 0
@main.nonConst1 = global [4 x i64] zeroinitializer
@main.nonConst2 = global i64 0
@main.someArray = global [8 x {i16, i32}] zeroinitializer
@main.exportedValue = global [1 x ptr] [ptr @main.exposedValue1]
@main.exportedConst = constant i64 42
@main.exposedValue1 = global i16 0
@main.exposedValue2 = global i16 0
@main.insertedValue = global {i8, i32, {float, {i64, i16}}} zeroinitializer

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
  %0 = load i64, ptr @main.v1
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
  store i64 3, ptr @main.v1
  call void @"main.init#1"()

  ; test the following pattern:
  ;     func someValue() int // extern function
  ;     var nonConst1 = [4]int{someValue(), 0, 0, 0}
  %value1 = call i64 @someValue()
  %gep1 = getelementptr [4 x i64], ptr @main.nonConst1, i32 0, i32 0
  store i64 %value1, ptr %gep1

  ; Test that the global really is marked dirty:
  ;     var nonConst2 = nonConst1[0]
  %gep2 = getelementptr [4 x i64], ptr @main.nonConst1, i32 0, i32 0
  %value2 = load i64, ptr %gep2
  store i64 %value2, ptr @main.nonConst2

  ; Test that the following GEP works:
  ;     var someArray
  ;     modifyExternal(&someArray[3].field1)
  %gep3 = getelementptr [8 x {i16, i32}], ptr @main.someArray, i32 0, i32 3, i32 1
  call void @modifyExternal(ptr %gep3)

  ; Test that marking a value as external also marks all referenced values.
  call void @modifyExternal(ptr @main.exportedValue)
  store i16 5, ptr @main.exposedValue1

  ; Test that marking a constant as external still allows loading from it.
  call void @readExternal(ptr @main.exportedConst)
  %constLoad = load i64, ptr @main.exportedConst
  call void @runtime.printint64(i64 %constLoad)

  ; Test that this even propagates through functions.
  call void @modifyExternal(ptr @willModifyGlobal)
  store i16 7, ptr @main.exposedValue2

  ; Test that inline assembly is ignored.
  call void @modifyExternal(ptr @hasInlineAsm)

  ; Test switch statement.
  %switch1 = call i64 @testSwitch(i64 1) ; 1 returns 6
  %switch2 = call i64 @testSwitch(i64 9) ; 9 returns the default value -1
  call void @runtime.printint64(i64 %switch1)
  call void @runtime.printint64(i64 %switch2)

  ; Test extractvalue/insertvalue with multiple operands.
  %agg = call {i8, i32, {float, {i64, i16}}} @nestedStruct()
  %elt = extractvalue {i8, i32, {float, {i64, i16}}} %agg, 2, 1, 0
  call void @runtime.printint64(i64 %elt)
  %agg2 = insertvalue {i8, i32, {float, {i64, i16}}} %agg, i64 5, 2, 1, 0
  store {i8, i32, {float, {i64, i16}}} %agg2, ptr @main.insertedValue

  ret void
}

define internal void @"main.init#1"() unnamed_addr {
entry:
  call void @runtime.printint64(i64 5)
  call void @runtime.printnl()
  ret void
}

declare i64 @someValue()

declare void @modifyExternal(ptr)

declare void @readExternal(ptr)

; This function will modify an external value. By passing this function as a
; function pointer to an external function, @main.exposedValue2 should be
; marked as external.
define void @willModifyGlobal() {
entry:
  store i16 8, ptr @main.exposedValue2
  ret void
}

; Inline assembly should be ignored in the interp package. While it is possible
; to modify other globals that way, usually that's not the case and there is no
; real way to check.
define void @hasInlineAsm() {
entry:
  call void asm sideeffect "", ""()
  ret void
}

define i64 @testSwitch(i64 %val) {
entry:
  ; Test switch statement.
  switch i64 %val, label %otherwise [ i64 0, label %zero
                                      i64 1, label %one
                                      i64 2, label %two ]
zero:
  ret i64 5

one:
  ret i64 6

two:
  ret i64 7

otherwise:
  ret i64 -1
}

declare {i8, i32, {float, {i64, i16}}} @nestedStruct()
