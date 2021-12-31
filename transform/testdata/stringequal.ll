target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@zeroString = constant [0 x i8] zeroinitializer

declare i1 @runtime.stringEqual(i8*, i32, i8*, i32, i8*)

define i1 @main.stringCompareEqualConstantZero(i8* %s1.data, i32 %s1.len, i8* %context) {
entry:
  %0 = call i1 @runtime.stringEqual(i8* %s1.data, i32 %s1.len, i8* getelementptr inbounds ([0 x i8], [0 x i8]* @zeroString, i32 0, i32 0), i32 0, i8* undef)
  ret i1 %0
}

define i1 @main.stringCompareUnequalConstantZero(i8* %s1.data, i32 %s1.len, i8* %context) {
entry:
  %0 = call i1 @runtime.stringEqual(i8* %s1.data, i32 %s1.len, i8* getelementptr inbounds ([0 x i8], [0 x i8]* @zeroString, i32 0, i32 0), i32 0, i8* undef)
  %1 = xor i1 %0, true
  ret i1 %1
}
