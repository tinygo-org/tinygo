; ModuleID = 'string.go'
source_filename = "string.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }

@"main$string" = internal unnamed_addr constant [3 x i8] c"foo", align 1

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
define hidden i32 @main.stringLen(ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  ret i32 %s.len
}

; Function Attrs: nounwind
define hidden i8 @main.stringIndex(ptr readonly %s.data, i32 %s.len, i32 %index, ptr %context) unnamed_addr #1 {
entry:
  %.not = icmp ult i32 %index, %s.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %entry
  %0 = getelementptr inbounds i8, ptr %s.data, i32 %index
  %1 = load i8, ptr %0, align 1
  ret i8 %1

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i8 undef
}

declare void @runtime.lookupPanic(ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.stringCompareEqual(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #3
  ret i1 %0
}

declare i1 @runtime.stringEqual(ptr readonly, i32, ptr readonly, i32, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.stringCompareUnequal(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #3
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareEqual(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s1.data, ptr nonnull %stackalloc, ptr undef) #3
  call void @runtime.trackPointer(ptr %s2.data, ptr nonnull %stackalloc, ptr undef) #3
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #3
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareUnequal(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s1.data, ptr nonnull %stackalloc, ptr undef) #3
  call void @runtime.trackPointer(ptr %s2.data, ptr nonnull %stackalloc, ptr undef) #3
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #3
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareSideEffects(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr undef) #3
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #3
  %2 = call { ptr, i32, i32 } @main.mutateBytes(ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr undef)
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #3
  %4 = extractvalue { ptr, i32, i32 } %2, 0
  %5 = extractvalue { ptr, i32, i32 } %2, 1
  %6 = extractvalue { ptr, i32, i32 } %2, 2
  %7 = call %runtime._string @runtime.stringFromBytes(ptr %4, i32 %5, i32 %6, ptr undef) #3
  %8 = extractvalue %runtime._string %7, 0
  call void @runtime.trackPointer(ptr %8, ptr nonnull %stackalloc, ptr undef) #3
  %9 = extractvalue %runtime._string %0, 0
  %10 = extractvalue %runtime._string %0, 1
  %11 = extractvalue %runtime._string %7, 0
  %12 = extractvalue %runtime._string %7, 1
  %13 = call i1 @runtime.stringEqual(ptr %9, i32 %10, ptr %11, i32 %12, ptr undef) #3
  ret i1 %13
}

declare %runtime._string @runtime.stringFromBytes(ptr nocapture readonly dereferenceable_or_null(1), i32, i32, ptr) #0

; Function Attrs: noinline nounwind
define hidden { ptr, i32, i32 } @main.mutateBytes(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #2 {
entry:
  %0 = icmp eq i32 %s.len, 0
  br i1 %0, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  br i1 false, label %lookup.throw, label %lookup.next3

lookup.next3:                                     ; preds = %lookup.next
  %1 = insertvalue { ptr, i32, i32 } zeroinitializer, ptr %s.data, 0
  %2 = insertvalue { ptr, i32, i32 } %1, i32 %s.len, 1
  %3 = insertvalue { ptr, i32, i32 } %2, i32 %s.cap, 2
  %4 = load i8, ptr %s.data, align 1
  %5 = add i8 %4, 1
  store i8 %5, ptr %s.data, align 1
  ret { ptr, i32, i32 } %3

lookup.throw:                                     ; preds = %lookup.next, %entry
  call void @runtime.lookupPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret { ptr, i32, i32 } undef
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareNil(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s.data, ptr nonnull %stackalloc, ptr undef) #3
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #3
  %0 = call i1 @runtime.stringEqual(ptr %s.data, i32 %s.len, ptr null, i32 0, ptr undef) #3
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.stringCompareLarger(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @runtime.stringLess(ptr %s2.data, i32 %s2.len, ptr %s1.data, i32 %s1.len, ptr undef) #3
  ret i1 %0
}

declare i1 @runtime.stringLess(ptr readonly, i32, ptr readonly, i32, ptr) #0

; Function Attrs: nounwind
define hidden i8 @main.stringLookup(ptr readonly %s.data, i32 %s.len, i8 %x, ptr %context) unnamed_addr #1 {
entry:
  %0 = zext i8 %x to i32
  %.not = icmp ugt i32 %s.len, %0
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %entry
  %1 = getelementptr inbounds nuw i8, ptr %s.data, i32 %0
  %2 = load i8, ptr %1, align 1
  ret i8 %2

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i8 undef
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { noinline nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nounwind }
