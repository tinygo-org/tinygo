target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%runtime._string = type { ptr, i32 }

declare %runtime._string @runtime.stringFromBytes(ptr, i32, i32, ptr)

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr)

declare void @runtime.trackPointer(ptr, ptr, ptr)

declare void @useString(ptr, i32)

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(ptr captures(none)) #0

define i1 @main.bytesEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef)
  ret i1 %0
}

define i1 @main.stringAndBytesEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %s.data, i32 %s.len, ptr %context) {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef)
  ret i1 %0
}

define i1 @main.keepStringConversion(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) {
entry:
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  call void @useString(ptr %1, i32 %a.len)
  %2 = call i1 @runtime.stringEqual(ptr %1, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef)
  ret i1 %2
}

define i32 @main.equalAndLen(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %s.data, i32 %s.len, ptr %context) {
entry:
  %equal = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef)
  %equal.ext = zext i1 %equal to i32
  %result = add i32 %a.len, %equal.ext
  ret i32 %result
}

define i1 @main.equalBeforeOtherUse(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %s.data, i32 %s.len, ptr %context) {
entry:
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  %equal = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef)
  call void @useString(ptr %1, i32 %a.len)
  ret i1 %equal
}

define i1 @main.twoComparisons(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %s.data, i32 %s.len, ptr %context) {
entry:
  %equal1 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef)
  %equal2 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef)
  %result = and i1 %equal1, %equal2
  ret i1 %result
}

define i1 @main.keepComparisonAfterMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %s.data, i32 %s.len, ptr %context) {
entry:
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  %equal1 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef)
  store i8 1, ptr %a.data, align 1
  %equal2 = call i1 @runtime.stringEqual(ptr %1, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef)
  %result = and i1 %equal1, %equal2
  ret i1 %result
}

define i1 @main.keepAfterLifetimeEnd(ptr %s.data, i32 %s.len, ptr %context) {
entry:
  %a.data = alloca [4 x i8], align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 4, i32 4, ptr undef)
  %1 = extractvalue %runtime._string %0, 0
  call void @llvm.lifetime.end.p0(ptr %a.data)
  %equal = call i1 @runtime.stringEqual(ptr %1, i32 4, ptr %s.data, i32 %s.len, ptr undef)
  ret i1 %equal
}

attributes #0 = { nocallback nofree nosync nounwind willreturn memory(argmem: readwrite) }
