target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

define internal i32 @main.add(i32 %x, i32 %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = add i32 %x, %y
  ret i32 %0
}

define internal void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}
