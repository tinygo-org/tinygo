target datalayout = "e-P1-p:16:8-i8:8-i16:8-i32:8-i64:8-f32:8-f64:8-n8-a:8"
target triple = "avr-unknown-unknown"

%runtime.typecodeID = type { %runtime.typecodeID*, i16 }
%runtime.funcValueWithSignature = type { i16, %runtime.typecodeID* }
%machine.UART = type { %machine.RingBuffer* }
%machine.RingBuffer = type { [128 x %"runtime/volatile.Register8"], %"runtime/volatile.Register8", %"runtime/volatile.Register8" }
%"runtime/volatile.Register8" = type { i8 }
%"runtime/interrupt.Interrupt" = type { i32 }

@"reflect/types.type:func:{named:runtime/interrupt.Interrupt}{}" = external constant %runtime.typecodeID
@"(machine.UART).Configure$1$withSignature" = internal constant %runtime.funcValueWithSignature { i16 ptrtoint (void (i32, i8*, i8*) addrspace(1)* @"(machine.UART).Configure$1" to i16), %runtime.typecodeID* @"reflect/types.type:func:{named:runtime/interrupt.Interrupt}{}" }
@machine.UART0 = internal global %machine.UART zeroinitializer
@"device/avr.init$string.18" = internal unnamed_addr constant [17 x i8] c"__vector_USART_RX"

declare void @"(machine.UART).Configure$1"(i32, i8*, i8*) unnamed_addr addrspace(1)

declare i32 @"runtime/interrupt.Register"(i32, i8*, i16, i8*, i8*) addrspace(1)

declare void @"runtime/interrupt.use"(%"runtime/interrupt.Interrupt") addrspace(1)

define void @"(machine.UART).Configure"(%machine.RingBuffer* %0, i32 %1, i8 %2, i8 %3, i8* %context, i8* %parentHandle) unnamed_addr addrspace(1) {
  ret void
}

define void @"device/avr.init"(i8* %context, i8* %parentHandle) unnamed_addr addrspace(1) {
entry:
  ret void
}

define avr_signalcc  void @__vector_USART_RX() unnamed_addr addrspace(1) section ".text.__vector_USART_RX" {
entry:
  call addrspace(1) void @"(machine.UART).Configure$1"(i32 18, i8* undef, i8* null)
  ret void
}
