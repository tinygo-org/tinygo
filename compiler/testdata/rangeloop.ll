; ModuleID = 'rangeloop.go'
source_filename = "rangeloop.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.rangeCount(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %rangeindex.body, %entry
  %0 = phi i32 [ 0, %entry ], [ %4, %rangeindex.body ]
  %1 = phi i32 [ -1, %entry ], [ %2, %rangeindex.body ]
  %2 = add i32 %1, 1
  %3 = icmp ult i32 %2, %s.len
  br i1 %3, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  %4 = add i32 %0, 1
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i32 @main.rangeIndexSum(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %rangeindex.body, %entry
  %0 = phi i32 [ 0, %entry ], [ %4, %rangeindex.body ]
  %1 = phi i32 [ -1, %entry ], [ %2, %rangeindex.body ]
  %2 = add i32 %1, 1
  %3 = icmp ult i32 %2, %s.len
  br i1 %3, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  %4 = add i32 %0, %2
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i32 @main.countUp(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %for.loop

for.loop:                                         ; preds = %for.body, %entry
  %0 = phi i32 [ 0, %entry ], [ %3, %for.body ]
  %1 = phi i32 [ 0, %entry ], [ %4, %for.body ]
  %2 = icmp slt i32 %1, %s.len
  br i1 %2, label %for.body, label %for.done

for.body:                                         ; preds = %for.loop
  %3 = add i32 %0, %1
  %4 = add i32 %1, 1
  br label %for.loop

for.done:                                         ; preds = %for.loop
  ret i32 %0
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
