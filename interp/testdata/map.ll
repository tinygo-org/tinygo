target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv6m-none-eabi"

%runtime._string = type { i8*, i32 }
%runtime.hashmap = type { %runtime.hashmap*, i8*, i32, i8, i8, i8 }

@main.m = global %runtime.hashmap* null
@main.binaryMap = global %runtime.hashmap* null
@main.stringMap = global %runtime.hashmap* null
@main.init.string = internal unnamed_addr constant [7 x i8] c"CONNECT"

declare %runtime.hashmap* @runtime.hashmapMake(i8, i8, i32, i8* %context, i8* %parentHandle)
declare void @runtime.hashmapBinarySet(%runtime.hashmap*, i8*, i8*, i8* %context, i8* %parentHandle)
declare void @runtime.hashmapStringSet(%runtime.hashmap*, i8*, i32, i8*, i8* %context, i8* %parentHandle)
declare void @llvm.lifetime.end.p0i8(i64, i8*)
declare void @llvm.lifetime.start.p0i8(i64, i8*)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init(i8* undef, i8* null)
  ret void
}

define internal void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
; Test that hashmap optimizations generally work (even with lifetimes).
  %hashmap.key = alloca i8
  %hashmap.value = alloca %runtime._string
  %0 = call %runtime.hashmap* @runtime.hashmapMake(i8 1, i8 8, i32 1, i8* undef, i8* null)
  %hashmap.value.bitcast = bitcast %runtime._string* %hashmap.value to i8*
  call void @llvm.lifetime.start.p0i8(i64 8, i8* %hashmap.value.bitcast)
  store %runtime._string { i8* getelementptr inbounds ([7 x i8], [7 x i8]* @main.init.string, i32 0, i32 0), i32 7 }, %runtime._string* %hashmap.value
  call void @llvm.lifetime.start.p0i8(i64 1, i8* %hashmap.key)
  store i8 1, i8* %hashmap.key
  call void @runtime.hashmapBinarySet(%runtime.hashmap* %0, i8* %hashmap.key, i8* %hashmap.value.bitcast, i8* undef, i8* null)
  call void @llvm.lifetime.end.p0i8(i64 1, i8* %hashmap.key)
  call void @llvm.lifetime.end.p0i8(i64 8, i8* %hashmap.value.bitcast)
  store %runtime.hashmap* %0, %runtime.hashmap** @main.m

  ; Other tests, that can be done in a separate function.
  call void @main.testNonConstantBinarySet()
  call void @main.testNonConstantStringSet()
  ret void
}

; Test that a map loaded from a global can still be used for mapassign
; operations (with binary keys).
define internal void @main.testNonConstantBinarySet() {
  %hashmap.key = alloca i8
  %hashmap.value = alloca i8
  ; Create hashmap from global.
  %map.new = call %runtime.hashmap* @runtime.hashmapMake(i8 1, i8 1, i32 1, i8* undef, i8* null)
  store %runtime.hashmap* %map.new, %runtime.hashmap** @main.binaryMap
  %map = load %runtime.hashmap*, %runtime.hashmap** @main.binaryMap
  ; Do the binary set to the newly loaded map.
  store i8 1, i8* %hashmap.key
  store i8 2, i8* %hashmap.value
  call void @runtime.hashmapBinarySet(%runtime.hashmap* %map, i8* %hashmap.key, i8* %hashmap.value, i8* undef, i8* null)
  ret void
}

; Test that a map loaded from a global can still be used for mapassign
; operations (with string keys).
define internal void @main.testNonConstantStringSet() {
  %hashmap.value = alloca i8
  ; Create hashmap from global.
  %map.new = call %runtime.hashmap* @runtime.hashmapMake(i8 8, i8 1, i32 1, i8* undef, i8* null)
  store %runtime.hashmap* %map.new, %runtime.hashmap** @main.stringMap
  %map = load %runtime.hashmap*, %runtime.hashmap** @main.stringMap
  ; Do the string set to the newly loaded map.
  store i8 2, i8* %hashmap.value
  call void @runtime.hashmapStringSet(%runtime.hashmap* %map, i8* getelementptr inbounds ([7 x i8], [7 x i8]* @main.init.string, i32 0, i32 0), i32 7, i8* %hashmap.value, i8* undef, i8* null)
  ret void
}
