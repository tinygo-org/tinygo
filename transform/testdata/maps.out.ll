target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@answer = constant [6 x i8] c"answer"

declare nonnull ptr @runtime.hashmapMake(i8, i8, i32)

declare void @runtime.hashmapStringSet(ptr nocapture, ptr, i32, ptr nocapture readonly)

declare i1 @runtime.hashmapStringGet(ptr nocapture, ptr, i32, ptr nocapture)

define void @testUnused() {
  ret void
}

define i32 @testReadonly() {
  %map = call ptr @runtime.hashmapMake(i8 4, i8 4, i32 0)
  %hashmap.value = alloca i32, align 4
  store i32 42, ptr %hashmap.value, align 4
  call void @runtime.hashmapStringSet(ptr %map, ptr @answer, i32 6, ptr %hashmap.value)
  %hashmap.value2 = alloca i32, align 4
  %commaOk = call i1 @runtime.hashmapStringGet(ptr %map, ptr @answer, i32 6, ptr %hashmap.value2)
  %loadedValue = load i32, ptr %hashmap.value2, align 4
  ret i32 %loadedValue
}

define ptr @testUsed() {
  %1 = call ptr @runtime.hashmapMake(i8 4, i8 4, i32 0)
  ret ptr %1
}
