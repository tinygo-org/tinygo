target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@device.CLOCKS = local_unnamed_addr global ptr inttoptr (i64 1073774592 to ptr)
@machine.clocks = local_unnamed_addr global ptr inttoptr (i64 1073774592 to ptr)
@machine.Serial = local_unnamed_addr global ptr null
@board.uart = global i32 42
@main.observed = local_unnamed_addr global ptr null

define void @runtime.socInit0(ptr %context) unnamed_addr {
entry:
  %clocks = load ptr, ptr @machine.clocks, align 8
  store volatile i32 1, ptr %clocks, align 4
  ret void
}

define void @runtime.hwInit0(ptr %context) unnamed_addr {
entry:
  call void @runtime.socInit0(ptr undef)
  ret void
}

define void @runtime.boardInit1(ptr %context) unnamed_addr {
entry:
  store ptr @board.uart, ptr @machine.Serial, align 8
  ret void
}

define void @runtime.hwInit1(ptr %context) unnamed_addr {
entry:
  call void @runtime.boardInit1(ptr undef)
  ret void
}

define void @runtime.initAll() unnamed_addr {
entry:
  %serial = load ptr, ptr @machine.Serial, align 8
  store ptr %serial, ptr @main.observed, align 8
  ret void
}
