target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@x = global i64 0
@ptr1 = global ptr @x
@ptr2 = global ptr @x
@split = global [2 x i64] zeroinitializer
@inFloat = global double 0.0
@inBytes = global [16 x i8] zeroinitializer
@copied = global [2 x i64] zeroinitializer
@element = global [2 x i64] zeroinitializer
@packed = global <{ i8, ptr }> zeroinitializer
@copiedElement = global [2 x ptr] zeroinitializer

declare void @llvm.memcpy.p0.p0.i64(ptr, ptr, i64, i1)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  ; A whole pointer across two i64 elements.
  call void @storePointer(ptr getelementptr inbounds (i8, ptr @split, i64 4))

  ; A whole pointer in a double.
  call void @storePointer(ptr @inFloat)

  ; A whole pointer in a byte array.
  call void @storePointer(ptr getelementptr inbounds (i8, ptr @inBytes, i64 8))

  ; Copy a whole pointer across two i64 elements.
  call void @llvm.memcpy.p0.p0.i64(ptr getelementptr inbounds (i8, ptr @copied, i64 4), ptr @ptr1, i64 8, i1 false)

  ; These pointers fill an element, so interp can write them in the initializer.
  call void @storePointer(ptr getelementptr inbounds (i8, ptr @element, i64 8))
  call void @storePointer(ptr getelementptr inbounds (i8, ptr @packed, i64 1))
  call void @llvm.memcpy.p0.p0.i64(ptr getelementptr inbounds (i8, ptr @copiedElement, i64 8), ptr @ptr2, i64 8, i1 false)
  ret void
}

define internal void @storePointer(ptr %dst) unnamed_addr {
entry:
  store i64 ptrtoint (ptr @x to i64), ptr %dst, align 8
  ret void
}
