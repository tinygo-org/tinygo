target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7em-none-eabi"

%machine.UART = type { ptr }
%machine.RingBuffer = type { [128 x %"runtime/volatile.Register8"], %"runtime/volatile.Register8", %"runtime/volatile.Register8" }
%"runtime/volatile.Register8" = type { i8 }
%"runtime/interrupt.handle" = type { ptr, i32, %"runtime/interrupt.Interrupt" }
%"runtime/interrupt.Interrupt" = type { i32 }

@"runtime/interrupt.$interrupt2" = private unnamed_addr constant %"runtime/interrupt.handle" { ptr @machine.UART0, i32 ptrtoint (ptr @"(*machine.UART).handleInterrupt$bound" to i32), %"runtime/interrupt.Interrupt" { i32 2 } }
@machine.UART0 = internal global %machine.UART { ptr @"machine$alloc.335" }
@"machine$alloc.335" = internal global %machine.RingBuffer zeroinitializer

declare void @"runtime/interrupt.callHandlers"(i32, ptr) local_unnamed_addr

declare void @"device/arm.EnableIRQ"(i32, ptr nocapture readnone)

declare void @"device/arm.SetPriority"(i32, i32, ptr nocapture readnone)

declare void @"runtime/interrupt.use"(%"runtime/interrupt.Interrupt")

define void @runtime.initAll(ptr nocapture readnone) unnamed_addr {
entry:
  call void @"device/arm.SetPriority"(i32 ptrtoint (ptr @"runtime/interrupt.$interrupt2" to i32), i32 192, ptr undef)
  call void @"device/arm.EnableIRQ"(i32 ptrtoint (ptr @"runtime/interrupt.$interrupt2" to i32), ptr undef)
  call void @"runtime/interrupt.use"(%"runtime/interrupt.Interrupt" { i32 ptrtoint (ptr @"runtime/interrupt.$interrupt2" to i32) })
  ret void
}

define void @UARTE0_UART0_IRQHandler() {
  call void @"runtime/interrupt.callHandlers"(i32 2, ptr undef)
  ret void
}

define void @NFCT_IRQHandler() {
  call void @"runtime/interrupt.callHandlers"(i32 5, ptr undef)
  ret void
}

define internal void @interruptSWVector(i32 %num) {
entry:
  switch i32 %num, label %switch.done [
    i32 2, label %switch.body2
    i32 5, label %switch.body5
  ]

switch.body2:
  call void @"runtime/interrupt.callHandlers"(i32 2, ptr undef)
  ret void

switch.body5:
  call void @"runtime/interrupt.callHandlers"(i32 5, ptr undef)
  ret void

switch.done:
  ret void
}

define internal void @"(*machine.UART).handleInterrupt$bound"(i32, ptr nocapture %context) {
entry:
  call void @"(*machine.UART).handleInterrupt"(ptr %context, i32 %0, ptr undef)
  ret void
}

declare void @"(*machine.UART).handleInterrupt"(ptr nocapture, i32, ptr nocapture readnone)
