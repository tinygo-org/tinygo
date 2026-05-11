target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@overlap.buf = global [4 x i8] c"\01\02\03\04"
@alias.src = global [4 x i8] c"\05\06\07\08"
@alias.dst = global [2 x i8] zeroinitializer
@reload.buf = global [4 x i8] c"\01\02\03\04"
@reload.out = global [2 x i8] zeroinitializer

define void @runtime.initAll() unnamed_addr {
entry:
  call void @overlap.init(ptr undef)
  call void @alias.init(ptr undef)
  call void @reload.init(ptr undef)
  ret void
}

define internal void @overlap.init(ptr %context) unnamed_addr {
entry:
  %tail = getelementptr [4 x i8], ptr @overlap.buf, i32 0, i32 3
  store i8 9, ptr %tail
  %val = load i16, ptr @overlap.buf
  %dst = getelementptr [4 x i8], ptr @overlap.buf, i32 0, i32 1
  store i16 %val, ptr %dst
  ret void
}

define internal void @alias.init(ptr %context) unnamed_addr {
entry:
  %src = getelementptr [4 x i8], ptr @alias.src, i32 0, i32 1
  %val = load i16, ptr %src
  store i16 %val, ptr @alias.dst
  store i8 9, ptr @alias.dst
  ret void
}

define internal void @reload.init(ptr %context) unnamed_addr {
entry:
  ; First store makes reload.buf writable in the current memory view.
  %tail = getelementptr [4 x i8], ptr @reload.buf, i32 0, i32 3
  store i8 9, ptr %tail
  ; Partial load whose result may share the underlying buffer.
  %val = load i16, ptr @reload.buf
  ; Subsequent in-place partial store; this must not corrupt %val.
  store i8 99, ptr @reload.buf
  ; Write the originally-loaded value to a separate global.
  store i16 %val, ptr @reload.out
  ret void
}
