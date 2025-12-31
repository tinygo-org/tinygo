target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@str1 = internal global [4 x i8] c"aacd"
@str2 = internal global [4 x i8] c"aazw"

@cmpLen0 = global i32 0
@cmp12 = global i32 0
@cmp21 = global i32 0
@cmp11 = global i32 0
@cmp22 = global i32 0
@cmp12Partial = global i32 0
@cmp21Partial = global i32 0

define void @runtime.initAll() unnamed_addr {
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
  call void @cmpAndStore(ptr @cmpLen0, ptr null, ptr null, i64 0)
  call void @cmpAndStore(ptr @cmp12, ptr @str1, ptr @str2, i64 4)
  call void @cmpAndStore(ptr @cmp21, ptr @str2, ptr @str1, i64 4)
  call void @cmpAndStore(ptr @cmp11, ptr @str1, ptr @str1, i64 4)
  call void @cmpAndStore(ptr @cmp22, ptr @str2, ptr @str2, i64 4)
  call void @cmpAndStore(ptr @cmp12Partial, ptr getelementptr inbounds (i8, ptr @str1, i32 1), ptr getelementptr inbounds (i8, ptr @str2, i32 1), i64 1)
  ret void
}

define internal void @cmpAndStore(ptr %dst, ptr %lhs, ptr %rhs, i64 %len) unnamed_addr {
  %cmp = call i32 @memcmp(ptr %lhs, ptr %rhs, i64 %len)
  store i32 %cmp, ptr %dst
  ret void
}

declare i32 @memcmp(ptr nocapture readonly, ptr nocapture readonly, i64)
