target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@device.CLOCKS = global ptr inttoptr (i64 1073774592 to ptr)
@machine.clocks = global ptr null
@machine.Serial = global ptr null
@board.uart = global i32 42
@main.observed = global ptr null

define void @runtime.socInit0(ptr %context) unnamed_addr {
entry:
  %clocks = load ptr, ptr @machine.clocks
  store volatile i32 1, ptr %clocks
  ret void
}

define void @runtime.hwInit0(ptr %context) unnamed_addr {
entry:
  call void @runtime.socInit0(ptr undef)
  ret void
}

define void @runtime.boardInit1(ptr %context) unnamed_addr {
entry:
  store ptr @board.uart, ptr @machine.Serial
  ret void
}

define void @runtime.hwInit1(ptr %context) unnamed_addr {
entry:
  call void @runtime.boardInit1(ptr undef)
  ret void
}

define void @runtime.initAll() unnamed_addr {
entry:
  call void @machine.init(ptr undef)
  call void @main.init(ptr undef)
  ret void
}

define internal void @machine.init(ptr %context) unnamed_addr {
entry:
  %clocks = load ptr, ptr @device.CLOCKS
  store ptr %clocks, ptr @machine.clocks
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  %serial = load ptr, ptr @machine.Serial
  store ptr %serial, ptr @main.observed
  ret void
}
