target datalayout = "e-p:32:32"
target triple = "wasm32-unknown-unknown"

@runtime.unwindPendingSignal = internal global i1 false
@value = global i32 0

define internal void @safe() {
entry:
  store i32 1, ptr @value
  ret void
}

define internal void @panics() {
entry:
  store i1 true, ptr @runtime.unwindPendingSignal
  ret void
}

declare void @external()

define i1 @checkSafe() {
entry:
  call void @safe()
  %unwind = call i1 @runtime.unwindPending()
  ret i1 %unwind
}

define i1 @checkPanic() {
entry:
  call void @panics()
  %unwind = call i1 @runtime.unwindPending()
  ret i1 %unwind
}

define i1 @checkExternal() {
entry:
  call void @external()
  %unwind = call i1 @runtime.unwindPending()
  ret i1 %unwind
}

define i1 @checkIndirect(ptr %fn) {
entry:
  call void %fn()
  %unwind = call i1 @runtime.unwindPending()
  ret i1 %unwind
}

define i1 @runtime.unwindPending() {
entry:
  %unwind = load i1, ptr @runtime.unwindPendingSignal
  ret i1 %unwind
}
