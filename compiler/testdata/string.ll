; ModuleID = 'string.go'
source_filename = "string.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }

@"main$string" = internal unnamed_addr constant [3 x i8] c"foo", align 1

declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._string @main.someString(ptr %context) unnamed_addr #1 {
entry:
  ret %runtime._string { ptr @"main$string", i32 3 }
}

; Function Attrs: nounwind
define hidden %runtime._string @main.zeroLengthString(ptr %context) unnamed_addr #1 {
entry:
  ret %runtime._string zeroinitializer
}

; Function Attrs: nounwind
define hidden i32 @main.stringLen(ptr %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  ret i32 %s.len
}

; Function Attrs: nounwind
define hidden i8 @main.stringIndex(ptr %s.data, i32 %s.len, i32 %index, ptr %context) unnamed_addr #1 {
entry:
  %.not = icmp ult i32 %index, %s.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %entry
  %0 = getelementptr inbounds i8, ptr %s.data, i32 %index
  %1 = load i8, ptr %0, align 1
  ret i8 %1

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #2
  unreachable
}

declare void @runtime.lookupPanic(ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.stringCompareEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #2
  ret i1 %0
}

declare i1 @runtime.stringEqual(ptr, i32, ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.stringCompareUnequal(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #2
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden i1 @main.stringCompareLarger(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @runtime.stringLess(ptr %s2.data, i32 %s2.len, ptr %s1.data, i32 %s1.len, ptr undef) #2
  ret i1 %0
}

declare i1 @runtime.stringLess(ptr, i32, ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden i8 @main.stringLookup(ptr %s.data, i32 %s.len, i8 %x, ptr %context) unnamed_addr #1 {
entry:
  %0 = zext i8 %x to i32
  %.not = icmp ult i32 %0, %s.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %entry
  %1 = getelementptr inbounds i8, ptr %s.data, i32 %0
  %2 = load i8, ptr %1, align 1
  ret i8 %2

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #2
  unreachable
}

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind }
