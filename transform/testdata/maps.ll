target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

@answer = constant [6 x i8] c"answer"

; func(keySize, valueSize uint8, sizeHint uintptr) *runtime.hashmap
declare nonnull ptr @runtime.hashmapMake(i8, i8, i32)

; func(map[string]int, string, unsafe.Pointer)
declare void @runtime.hashmapStringSet(ptr nocapture, ptr, i32, ptr nocapture readonly)

; func(map[string]int, string, unsafe.Pointer)
declare i1 @runtime.hashmapStringGet(ptr nocapture, ptr, i32, ptr nocapture)

define void @testUnused() {
    ; create the map
    %map = call ptr @runtime.hashmapMake(i8 4, i8 4, i32 0)
    ; create the value to be stored
    %hashmap.value = alloca i32
    store i32 42, ptr %hashmap.value
    ; store the value
    call void @runtime.hashmapStringSet(ptr %map, ptr @answer, i32 6, ptr %hashmap.value)
    ret void
}

; Note that the following function should ideally be optimized (it could simply
; return 42), but isn't at the moment.
define i32 @testReadonly() {
    ; create the map
    %map = call ptr @runtime.hashmapMake(i8 4, i8 4, i32 0)

    ; create the value to be stored
    %hashmap.value = alloca i32
    store i32 42, ptr %hashmap.value

    ; store the value
    call void @runtime.hashmapStringSet(ptr %map, ptr @answer, i32 6, ptr %hashmap.value)

    ; load the value back
    %hashmap.value2 = alloca i32
    %commaOk = call i1 @runtime.hashmapStringGet(ptr %map, ptr @answer, i32 6, ptr %hashmap.value2)
    %loadedValue = load i32, ptr %hashmap.value2

    ret i32 %loadedValue
}

define ptr @testUsed() {
    %1 = call ptr @runtime.hashmapMake(i8 4, i8 4, i32 0)
    ret ptr %1
}
