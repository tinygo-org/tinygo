target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@"runtime.lookupPanic$string" = constant [18 x i8] c"index out of range"

declare void @runtime.runtimePanic(i8*, i32)

declare void @runtime._panic(i32, i8*)

define void @runtime.lookupPanic() {
  call void @llvm.trap()
  call void @runtime.runtimePanic(i8* getelementptr inbounds ([18 x i8], [18 x i8]* @"runtime.lookupPanic$string", i64 0, i64 0), i32 18)
  ret void
}

define void @someFunc(i32 %typecode, i8* %value) {
  call void @llvm.trap()
  call void @runtime._panic(i32 %typecode, i8* %value)
  unreachable
}

; Function Attrs: cold noreturn nounwind
declare void @llvm.trap() #0

attributes #0 = { cold noreturn nounwind }
