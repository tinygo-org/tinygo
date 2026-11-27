target datalayout = "e-m:e-i16:32-i64:64-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.src = global [2 x i16] [i16 258, i16 772]
@main.dst = global [2 x i16] zeroinitializer
@main.read = global i8 0

define void @runtime.initAll() {
  call void @main.init(ptr undef)
  ret void
}

define internal void @main.init(ptr %context) {
  %src = getelementptr inbounds i8, ptr @main.src, i32 4
  %v = load i8, ptr %src
  store i8 %v, ptr @main.read
  %dst = getelementptr inbounds i8, ptr @main.dst, i32 4
  store i8 9, ptr %dst
  ret void
}
