target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7em-none-eabi"

%machine.UART = type { %machine.RingBuffer* }
%machine.RingBuffer = type { [128 x %"runtime/volatile.Register8"], %"runtime/volatile.Register8", %"runtime/volatile.Register8" }
%"runtime/volatile.Register8" = type { i8 }

@machine.UART0 = internal global %machine.UART { %machine.RingBuffer* @"machine$alloc.335" }
@"machine$alloc.335" = internal global %machine.RingBuffer zeroinitializer
@"device/nrf.init$string.2" = internal unnamed_addr constant [23 x i8] c"UARTE0_UART0_IRQHandler"
@"device/nrf.init$string.3" = internal unnamed_addr constant [44 x i8] c"SPIM0_SPIS0_TWIM0_TWIS0_SPI0_TWI0_IRQHandler"

declare i32 @"runtime/interrupt.Register"(i32, i8*, i32, i8*, i8*) local_unnamed_addr

declare void @"device/arm.EnableIRQ"(i32, i8* nocapture readnone, i8* nocapture readnone)

declare void @"device/arm.SetPriority"(i32, i32, i8* nocapture readnone, i8* nocapture readnone)

define void @runtime.initAll(i8* nocapture readnone %0, i8* nocapture readnone %1) unnamed_addr {
entry:
  call void @"device/arm.SetPriority"(i32 2, i32 192, i8* undef, i8* undef)
  call void @"device/arm.EnableIRQ"(i32 2, i8* undef, i8* undef)
  ret void
}

define internal void @"(*machine.UART).handleInterrupt$bound"(i32 %0, i8* nocapture %context, i8* nocapture readnone %parentHandle) {
entry:
  %unpack.ptr = bitcast i8* %context to %machine.UART*
  call void @"(*machine.UART).handleInterrupt"(%machine.UART* %unpack.ptr, i32 %0, i8* undef, i8* undef)
  ret void
}

declare void @"(*machine.UART).handleInterrupt"(%machine.UART* nocapture, i32, i8* nocapture readnone, i8* nocapture readnone)

define void @UARTE0_UART0_IRQHandler() unnamed_addr section ".text.UARTE0_UART0_IRQHandler" {
entry:
  call void @"(*machine.UART).handleInterrupt$bound"(i32 2, i8* bitcast (%machine.UART* @machine.UART0 to i8*), i8* null)
  ret void
}
