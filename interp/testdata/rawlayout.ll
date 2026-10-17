target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.x = global i64 0
@main.pointerFree = global ptr null
@main.misaligned = global ptr null
@main.self = global ptr null
@main.straddle = global ptr null
@main.aligned = global ptr null

declare ptr @runtime.alloc(i64, ptr) unnamed_addr

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
entry:
  ; Pointer stored as uintptr in an object without pointers.
  %pointerFree = call ptr @runtime.alloc(i64 16, ptr inttoptr (i64 3 to ptr))
  %pointerFree.1 = getelementptr inbounds i8, ptr %pointerFree, i64 8
  store i64 ptrtoint (ptr @main.x to i64), ptr %pointerFree.1
  store ptr %pointerFree, ptr @main.pointerFree

  ; The same, but at an offset that is not pointer aligned.
  %misaligned = call ptr @runtime.alloc(i64 16, ptr inttoptr (i64 3 to ptr))
  %misaligned.1 = getelementptr inbounds i8, ptr %misaligned, i64 4
  store i64 ptrtoint (ptr @main.x to i64), ptr %misaligned.1, align 4
  store ptr %misaligned, ptr @main.misaligned

  ; Object without pointers that holds a pointer to itself.
  %self = call ptr @runtime.alloc(i64 16, ptr inttoptr (i64 3 to ptr))
  %self.1 = getelementptr inbounds i8, ptr %self, i64 8
  %self.int = ptrtoint ptr %self to i64
  store i64 %self.int, ptr %self.1
  store ptr %self, ptr @main.self

  ; Object with layout { ptr, i64, i64 } and a pointer across both integers.
  %straddle = call ptr @runtime.alloc(i64 24, ptr inttoptr (i64 135 to ptr))
  store ptr @main.x, ptr %straddle
  %straddle.1 = getelementptr inbounds i8, ptr %straddle, i64 12
  store i64 ptrtoint (ptr @main.x to i64), ptr %straddle.1, align 4
  store ptr %straddle, ptr @main.straddle

  ; Pointer in an integer of the layout, which keeps the layout type.
  %aligned = call ptr @runtime.alloc(i64 24, ptr inttoptr (i64 135 to ptr))
  %aligned.1 = getelementptr inbounds i8, ptr %aligned, i64 8
  store i64 ptrtoint (ptr @main.x to i64), ptr %aligned.1
  store ptr %aligned, ptr @main.aligned
  ret void
}
