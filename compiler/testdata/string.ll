; ModuleID = 'string.go'
source_filename = "string.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }

@"main$string" = internal unnamed_addr constant [3 x i8] c"foo", align 1

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._string @main.someString(ptr %context) unnamed_addr #2 {
entry:
  ret %runtime._string { ptr @"main$string", i32 3 }
}

; Function Attrs: nounwind
define hidden %runtime._string @main.zeroLengthString(ptr %context) unnamed_addr #2 {
entry:
  ret %runtime._string zeroinitializer
}

; Function Attrs: nounwind
define hidden i32 @main.stringLen(ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #2 {
entry:
  ret i32 %s.len
}

; Function Attrs: nounwind
define hidden i8 @main.stringIndex(ptr readonly %s.data, i32 %s.len, i32 %index, ptr %context) unnamed_addr #2 {
entry:
  %.not = icmp ult i32 %index, %s.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %entry
  %0 = getelementptr inbounds i8, ptr %s.data, i32 %index
  %1 = load i8, ptr %0, align 1
  ret i8 %1

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #4
  unreachable
}

declare void @runtime.lookupPanic(ptr) #1

; Function Attrs: nounwind
define hidden i1 @main.stringCompareEqual(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #2 {
entry:
  %streq.len.eq = icmp eq i32 %s1.len, %s2.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %s1.data, ptr %s2.data, i32 %s1.len) #4
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

declare i32 @memcmp(ptr nocapture readonly, ptr nocapture readonly, i32)

; Function Attrs: nounwind
define hidden i1 @main.stringCompareUnequal(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #2 {
entry:
  %streq.len.eq = icmp eq i32 %s1.len, %s2.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %s1.data, ptr %s2.data, i32 %s1.len) #4
  %streq.memcmp.eq = icmp ne i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %streq.not = phi i1 [ true, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %streq.not
}

; Function Attrs: nounwind
define hidden i1 @main.stringCompareLarger(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #2 {
entry:
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %s2.len, i32 %s1.len)
  %strlt.memcmp = call i32 @memcmp(ptr %s2.data, ptr %s1.data, i32 %strlt.min.len) #4
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %s2.len, %s1.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nocallback nofree nosync nounwind speculatable willreturn memory(none)
declare i32 @llvm.umin.i32(i32, i32) #3

; Function Attrs: nounwind
define hidden i8 @main.stringLookup(ptr readonly %s.data, i32 %s.len, i8 %x, ptr %context) unnamed_addr #2 {
entry:
  %0 = zext i8 %x to i32
  %.not = icmp ugt i32 %s.len, %0
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %entry
  %1 = getelementptr inbounds nuw i8, ptr %s.data, i32 %0
  %2 = load i8, ptr %1, align 1
  ret i8 %2

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #4
  unreachable
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nocallback nofree nosync nounwind speculatable willreturn memory(none) }
attributes #4 = { nounwind }
