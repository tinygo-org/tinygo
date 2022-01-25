target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7em-none-eabi"

%machine.UART = type { %machine.RingBuffer* }
%machine.RingBuffer = type { [128 x %"runtime/volatile.Register8"], %"runtime/volatile.Register8", %"runtime/volatile.Register8" }
%"runtime/volatile.Register8" = type { i8 }
%"runtime/interrupt.handle" = type { i8*, i32, %"runtime/interrupt.Interrupt" }
%"runtime/interrupt.Interrupt" = type { i32 }

@"runtime/interrupt.$interrupt2" = private unnamed_addr constant %"runtime/interrupt.handle" { i8* bitcast (%machine.UART* @machine.UART0 to i8*), i32 ptrtoint (void (i32, i8*)* @"(*machine.UART).handleInterrupt$bound" to i32), %"runtime/interrupt.Interrupt" { i32 2 } }
@machine.UART0 = internal global %machine.UART { %machine.RingBuffer* @"machine$alloc.335" }
@"machine$alloc.335" = internal global %machine.RingBuffer zeroinitializer

declare void @"runtime/interrupt.callHandlers"(i32, i8*) local_unnamed_addr

declare void @"device/arm.EnableIRQ"(i32, i8* nocapture readnone)

declare void @"device/arm.SetPriority"(i32, i32, i8* nocapture readnone)

declare void @"runtime/interrupt.use"(%"runtime/interrupt.Interrupt")

define void @runtime.initAll(i8* nocapture readnone) unnamed_addr {
entry:
  call void @"device/arm.SetPriority"(i32 ptrtoint (%"runtime/interrupt.handle"* @"runtime/interrupt.$interrupt2" to i32), i32 192, i8* undef)
  call void @"device/arm.EnableIRQ"(i32 ptrtoint (%"runtime/interrupt.handle"* @"runtime/interrupt.$interrupt2" to i32), i8* undef)
  call void @"runtime/interrupt.use"(%"runtime/interrupt.Interrupt" { i32 ptrtoint (%"runtime/interrupt.handle"* @"runtime/interrupt.$interrupt2" to i32) })
  ret void
}

define void @UARTE0_UART0_IRQHandler() {
  call void @"runtime/interrupt.callHandlers"(i32 2, i8* undef)
  ret void
}

define void @NFCT_IRQHandler() {
  call void @"runtime/interrupt.callHandlers"(i32 5, i8* undef)
  ret void
}

define internal void @interruptSWVector(i32 %num) {
entry:
  switch i32 %num, label %switch.done [
    i32 2, label %switch.body2
    i32 5, label %switch.body5
  ]

switch.body2:
  call void @"runtime/interrupt.callHandlers"(i32 2, i8* undef)
  ret void

switch.body5:
  call void @"runtime/interrupt.callHandlers"(i32 5, i8* undef)
  ret void

switch.done:
  ret void
}

define internal void @"(*machine.UART).handleInterrupt$bound"(i32, i8* nocapture %context) {
entry:
  %unpack.ptr = bitcast i8* %context to %machine.UART*
  call void @"(*machine.UART).handleInterrupt"(%machine.UART* %unpack.ptr, i32 %0, i8* undef)
  ret void
}

declare void @"(*machine.UART).handleInterrupt"(%machine.UART* nocapture, i32, i8* nocapture readnone)
