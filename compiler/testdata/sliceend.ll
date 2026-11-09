; ModuleID = 'sliceend.go'
source_filename = "sliceend.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.sliceTail(ptr %ints.data, i32 %ints.len, i32 %ints.cap, i32 %low, ptr %context) unnamed_addr #1 {
entry:
  %slice.lowhigh = icmp ugt i32 %low, %ints.len
  br i1 %slice.lowhigh, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %0 = sub i32 %ints.cap, %low
  %1 = icmp eq i32 %ints.cap, %low
  %2 = select i1 %1, i32 0, i32 %low
  %3 = getelementptr inbounds i32, ptr %ints.data, i32 %2
  %4 = sub nuw i32 %ints.len, %low
  %5 = insertvalue { ptr, i32, i32 } undef, ptr %3, 0
  %6 = insertvalue { ptr, i32, i32 } %5, i32 %4, 1
  %7 = insertvalue { ptr, i32, i32 } %6, i32 %0, 2
  ret { ptr, i32, i32 } %7

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret { ptr, i32, i32 } undef
}

declare void @runtime.slicePanic(ptr) #0

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.sliceTailMax(ptr %ints.data, i32 %ints.len, i32 %ints.cap, i32 %low, i32 %high, i32 %max, ptr %context) unnamed_addr #1 {
entry:
  %slice.lowhigh = icmp ugt i32 %low, %high
  %slice.highmax = icmp ugt i32 %high, %max
  %slice.lowmax = or i1 %slice.lowhigh, %slice.highmax
  %slice.maxcap = icmp ugt i32 %max, %ints.cap
  %slice.lowcap = or i1 %slice.lowmax, %slice.maxcap
  br i1 %slice.lowcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %0 = sub i32 %max, %low
  %1 = icmp eq i32 %max, %low
  %2 = select i1 %1, i32 0, i32 %low
  %3 = getelementptr inbounds i32, ptr %ints.data, i32 %2
  %4 = sub nuw i32 %high, %low
  %5 = insertvalue { ptr, i32, i32 } undef, ptr %3, 0
  %6 = insertvalue { ptr, i32, i32 } %5, i32 %4, 1
  %7 = insertvalue { ptr, i32, i32 } %6, i32 %0, 2
  ret { ptr, i32, i32 } %7

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret { ptr, i32, i32 } undef
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.sliceArrayTail(ptr dereferenceable_or_null(16) %a, i32 %low, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %a, null
  br i1 %0, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %slice.lowhigh = icmp ugt i32 %low, 4
  br i1 %slice.lowhigh, label %slice.throw1, label %slice.next2

slice.next2:                                      ; preds = %slice.next
  %slice.len = sub nuw nsw i32 4, %low
  %slice.cap = sub nuw nsw i32 4, %low
  %1 = icmp eq i32 %low, 4
  %2 = select i1 %1, i32 0, i32 %low
  %slice.ptr = getelementptr inbounds nuw i32, ptr %a, i32 %2
  %3 = insertvalue { ptr, i32, i32 } undef, ptr %slice.ptr, 0
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %slice.len, 1
  %5 = insertvalue { ptr, i32, i32 } %4, i32 %slice.cap, 2
  ret { ptr, i32, i32 } %5

slice.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw1, %slice.throw
  ret { ptr, i32, i32 } undef

slice.throw1:                                     ; preds = %slice.next
  call void @runtime.slicePanic(ptr undef) #2
  br label %unwind.return
}

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define hidden %runtime._string @main.stringTail(ptr readonly %s.data, i32 %s.len, i32 %low, ptr %context) unnamed_addr #1 {
entry:
  %slice.lowhigh = icmp ugt i32 %low, %s.len
  br i1 %slice.lowhigh, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %0 = sub nuw i32 %s.len, %low
  %1 = icmp eq i32 %s.len, %low
  %2 = select i1 %1, i32 0, i32 %low
  %3 = getelementptr inbounds i8, ptr %s.data, i32 %2
  %4 = insertvalue %runtime._string undef, ptr %3, 0
  %5 = insertvalue %runtime._string %4, i32 %0, 1
  ret %runtime._string %5

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret %runtime._string undef
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.sliceConstMax(ptr %ints.data, i32 %ints.len, i32 %ints.cap, i32 %low, ptr %context) unnamed_addr #1 {
entry:
  %0 = add i32 %low, 2
  %slice.lowmax = icmp ugt i32 %low, -3
  %slice.maxcap = icmp ugt i32 %0, %ints.cap
  %slice.lowcap = or i1 %slice.lowmax, %slice.maxcap
  br i1 %slice.lowcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %1 = getelementptr inbounds i32, ptr %ints.data, i32 %low
  %2 = insertvalue { ptr, i32, i32 } undef, ptr %1, 0
  %3 = insertvalue { ptr, i32, i32 } %2, i32 1, 1
  %4 = insertvalue { ptr, i32, i32 } %3, i32 2, 2
  ret { ptr, i32, i32 } %4

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret { ptr, i32, i32 } undef
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind }
