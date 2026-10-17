target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.a = global [256 x i8] zeroinitializer, align 256
@main.b = global [256 x i8] zeroinitializer, align 256
@main.pa = global ptr @main.a
@main.pb = global ptr @main.b
@main.pa1 = global ptr @main.a
@main.pb1 = global ptr @main.b
@main.pa2 = global ptr @main.a
@main.pa3 = global ptr @main.a
@main.pointersEqual = global i1 true
@main.bytesEqual = global i1 true
@main.byteIsZero = global i1 true
@main.sameBytesDiffer = global i1 true

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) unnamed_addr {
entry:
  ; Compare whole pointers at compile time.
  %pa = load ptr, ptr @main.pa
  %pb = load ptr, ptr @main.pb
  %pointersEqual = icmp eq ptr %pa, %pb
  store i1 %pointersEqual, ptr @main.pointersEqual

  ; Compare bytes of pointers only at runtime.
  call void @main.compareBytes(ptr undef)
  call void @main.compareByteWithZero(ptr undef)
  call void @main.compareBytesOfSamePointer(ptr undef)
  ret void
}

define internal void @main.compareBytes(ptr %context) unnamed_addr {
entry:
  %av = load i8, ptr @main.pa1
  %bv = load i8, ptr @main.pb1
  %equal = icmp eq i8 %av, %bv
  store i1 %equal, ptr @main.bytesEqual
  ret void
}

define internal void @main.compareByteWithZero(ptr %context) unnamed_addr {
entry:
  %av = load i8, ptr @main.pa2
  %isZero = icmp eq i8 %av, 0
  store i1 %isZero, ptr @main.byteIsZero
  ret void
}

define internal void @main.compareBytesOfSamePointer(ptr %context) unnamed_addr {
entry:
  %b0 = load i8, ptr @main.pa3
  %b1 = load i8, ptr getelementptr inbounds (i8, ptr @main.pa3, i64 1)
  %notEqual = icmp ne i8 %b0, %b1
  store i1 %notEqual, ptr @main.sameBytesDiffer
  ret void
}
