target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@"runtime.lookupPanic$string" = constant [18 x i8] c"index out of range"

declare void @runtime.runtimePanic(ptr, i32)

declare void @runtime._panic(i32, ptr)

define void @runtime.lookupPanic() {
  call void @runtime.runtimePanic(ptr @"runtime.lookupPanic$string", i32 18)
  ret void
}

; This is equivalent to the following code:
;     func someFunc(x interface{}) {
;         panic(x)
;     }
define void @someFunc(i32 %typecode, ptr %value) {
  call void @runtime._panic(i32 %typecode, ptr %value)
  unreachable
}
