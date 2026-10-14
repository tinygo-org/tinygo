target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%inner = type { ptr, i32 }
%outer = type { i32, %inner }

declare ptr @runtime.alloc(i32, ptr)

define %outer @wrap(ptr %p) {
  %inner = insertvalue %inner { ptr null, i32 7 }, ptr %p, 0
  %outer = insertvalue %outer { i32 3, %inner undef }, %inner %inner, 1
  ret %outer %outer
}

define %inner @aggregateSelect(i1 %cond) {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  %value = select i1 %cond, %inner %inner, %inner zeroinitializer
  ret %inner %value
}

define [2 x i32] @scalarSelect(i1 %cond) {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %n = ptrtoint ptr %p to i32
  %a = insertvalue [2 x i32] zeroinitializer, i32 %n, 0
  %value = select i1 %cond, [2 x i32] %a, [2 x i32] zeroinitializer
  ret [2 x i32] %value
}
