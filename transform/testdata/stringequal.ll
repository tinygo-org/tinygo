target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@zeroString = constant [0 x i8] zeroinitializer

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr)

define i1 @main.stringCompareEqualConstantZero(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @zeroString, i32 0, ptr undef)
  ret i1 %0
}

define i1 @main.stringCompareUnequalConstantZero(ptr %s1.data, i32 %s1.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr @zeroString, i32 0, ptr undef)
  %1 = xor i1 %0, true
  ret i1 %1
}
