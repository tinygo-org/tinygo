target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%runtime.hashmap = type { %runtime.hashmap*, i8*, i32, i8, i8, i8 }

@answer = constant [6 x i8] c"answer"

declare nonnull %runtime.hashmap* @runtime.hashmapMake(i8, i8, i32)

declare void @runtime.hashmapStringSet(%runtime.hashmap* nocapture, i8*, i32, i8* nocapture readonly)

declare i1 @runtime.hashmapStringGet(%runtime.hashmap* nocapture, i8*, i32, i8* nocapture)

define void @testUnused() {
  ret void
}

define i32 @testReadonly() {
  %map = call %runtime.hashmap* @runtime.hashmapMake(i8 4, i8 4, i32 0)
  %hashmap.value = alloca i32, align 4
  store i32 42, i32* %hashmap.value, align 4
  %hashmap.value.bitcast = bitcast i32* %hashmap.value to i8*
  call void @runtime.hashmapStringSet(%runtime.hashmap* %map, i8* getelementptr inbounds ([6 x i8], [6 x i8]* @answer, i32 0, i32 0), i32 6, i8* %hashmap.value.bitcast)
  %hashmap.value2 = alloca i32, align 4
  %hashmap.value2.bitcast = bitcast i32* %hashmap.value2 to i8*
  %commaOk = call i1 @runtime.hashmapStringGet(%runtime.hashmap* %map, i8* getelementptr inbounds ([6 x i8], [6 x i8]* @answer, i32 0, i32 0), i32 6, i8* %hashmap.value2.bitcast)
  %loadedValue = load i32, i32* %hashmap.value2, align 4
  ret i32 %loadedValue
}

define %runtime.hashmap* @testUsed() {
  %1 = call %runtime.hashmap* @runtime.hashmapMake(i8 4, i8 4, i32 0)
  ret %runtime.hashmap* %1
}
