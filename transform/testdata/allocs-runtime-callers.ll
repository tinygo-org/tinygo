target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

declare nonnull ptr @runtime.alloc(i32, ptr)

; Allocates in the runtime like append does.
define ptr @runtime.grow(i32 %n) {
  %buf = call ptr @runtime.alloc(i32 %n, ptr null)
  ret ptr %buf
}

define ptr @runtime.append(i32 %n) {
  %buf = call ptr @runtime.grow(i32 %n)
  ret ptr %buf
}

define ptr @main.first(i32 %n) {
  %buf = call ptr @runtime.append(i32 %n)
  ret ptr %buf
}

define ptr @main.second(i32 %n) {
  %buf = call ptr @runtime.append(i32 %n)
  ret ptr %buf
}
