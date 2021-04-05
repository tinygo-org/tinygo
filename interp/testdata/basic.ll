target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = internal global i64 0
@main.nonConst1 = global [4 x i64] zeroinitializer
@main.nonConst2 = global i64 0
@main.someArray = global [8 x {i16, i32}] zeroinitializer
@main.exportedValue = global [1 x i16*] [i16* @main.exposedValue1]
@main.exposedValue1 = global i16 0
@main.exposedValue2 = global i16 0

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

  ; Test that the following GEP works:
  ;     var someArray
  ;     modifyExternal(&someArray[3].field1)
  %gep3 = getelementptr [8 x {i16, i32}], [8 x {i16, i32}]* @main.someArray, i32 0, i32 3, i32 1
  call void @modifyExternal(i32* %gep3)

  ; Test that marking a value as external also marks all referenced values.
  call void @modifyExternal(i32* bitcast ([1 x i16*]* @main.exportedValue to i32*))
  store i16 5, i16* @main.exposedValue1

  ; Test that this even propagates through functions.
  call void @modifyExternal(i32* bitcast (void ()* @willModifyGlobal to i32*))
  store i16 7, i16* @main.exposedValue2

  ; Test switch statement.
  %switch1 = call i64 @testSwitch(i64 1) ; 1 returns 6
  %switch2 = call i64 @testSwitch(i64 9) ; 9 returns the default value -1
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

; This function will modify an external value. By passing this function as a
; function pointer to an external function, @main.exposedValue2 should be
; marked as external.
define void @willModifyGlobal() {
entry:
  store i16 8, i16* @main.exposedValue2
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
