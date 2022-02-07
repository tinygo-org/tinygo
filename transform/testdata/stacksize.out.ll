target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@"internal/task.stackSizes" = global [1 x i32] [i32 1024], section ".tinygo_stacksizes"
@llvm.used = appending global [2 x i8*] [i8* bitcast ([1 x i32]* @"internal/task.stackSizes" to i8*), i8* bitcast (void (i8*)* @"runtime.run$1$gowrapper" to i8*)]

declare i32 @"internal/task.getGoroutineStackSize"(i32, i8*, i8*)

declare void @"runtime.run$1$gowrapper"(i8*)

declare void @"internal/task.start"(i32, i8*, i32)

define void @Reset_Handler() {
entry:
  %stacksize1 = load i32, i32* getelementptr inbounds ([1 x i32], [1 x i32]* @"internal/task.stackSizes", i32 0, i32 0), align 4
  call void @"internal/task.start"(i32 ptrtoint (void (i8*)* @"runtime.run$1$gowrapper" to i32), i8* undef, i32 %stacksize1)
  ret void
}
