; ModuleID = 'lenrange.go'
source_filename = "lenrange.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%main.sliceField = type { { ptr, i32, i32 }, i32 }

@tinygo.indirect-abi = appending global [1 x ptr] [ptr @main.indirectResult]
@llvm.used = appending global [2 x ptr] [ptr @main.indirectResult, ptr @exportedArgs]

declare void @runtime.trackPointer(ptr readonly captures(none), ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @"(main.byteSlice).size"(ptr %s.data, i32 range(i32 0, -2147483648) %s.len, i32 range(i32 0, -2147483648) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  ret i32 %s.len
}

; Function Attrs: nounwind
define hidden i32 @main.sliceArgs(ptr %s.data, i32 range(i32 0, -2147483648) %s.len, i32 range(i32 0, -2147483648) %s.cap, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %0 = add nuw i32 %s.len, %s.cap
  %1 = add i32 %0, %n
  ret i32 %1
}

; Function Attrs: nounwind
define hidden i32 @main.stringArg(ptr readonly %s.data, i32 range(i32 0, -2147483648) %s.len, ptr %context) unnamed_addr #1 {
entry:
  ret i32 %s.len
}

; Function Attrs: nounwind
define hidden i32 @main.wideArgs(ptr %s.data, i32 range(i32 0, 1073741824) %s.len, i32 range(i32 0, 1073741824) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = add nuw nsw i32 %s.len, %s.cap
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i32 @main.namedArgs(ptr %s.data, i32 range(i32 0, -2147483648) %s.len, i32 range(i32 0, -2147483648) %s.cap, ptr readonly %text.data, i32 range(i32 0, -2147483648) %text.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = add nuw i32 %s.len, %s.cap
  %1 = add i32 %0, %text.len
  ret i32 %1
}

; Function Attrs: nounwind
define hidden i32 @main.nestedArg(ptr readonly %s.text.data, i32 range(i32 0, -2147483648) %s.text.len, i32 %s.n, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %s = call align 4 dereferenceable(12) ptr @runtime.alloc(i32 12, ptr nonnull inttoptr (i32 71 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %s, ptr nonnull %stackalloc, ptr undef) #4
  store ptr %s.text.data, ptr %s, align 4
  %s.repack4 = getelementptr inbounds nuw i8, ptr %s, i32 4
  store i32 %s.text.len, ptr %s.repack4, align 4
  %s.repack2 = getelementptr inbounds nuw i8, ptr %s, i32 8
  store i32 %s.n, ptr %s.repack2, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load ptr, ptr %s, align 4
  %.elt6 = getelementptr inbounds nuw i8, ptr %s, i32 4
  %.unpack7 = load i32, ptr %.elt6, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #4
  br i1 false, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %0 = getelementptr inbounds nuw i8, ptr %s, i32 8
  %1 = load i32, ptr %0, align 4
  %2 = add i32 %.unpack7, %1
  ret i32 %2

deref.throw:                                      ; preds = %deref.next, %entry
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i32 undef
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #2

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define hidden i32 @main.unexpandedArg(%main.sliceField %s, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %s1 = call align 4 dereferenceable(16) ptr @runtime.alloc(i32 16, ptr nonnull inttoptr (i32 73 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %s1, ptr nonnull %stackalloc, ptr undef) #4
  %s.elt = extractvalue %main.sliceField %s, 0
  %s.elt.elt = extractvalue { ptr, i32, i32 } %s.elt, 0
  store ptr %s.elt.elt, ptr %s1, align 4
  %s1.repack5 = getelementptr inbounds nuw i8, ptr %s1, i32 4
  %s.elt.elt6 = extractvalue { ptr, i32, i32 } %s.elt, 1
  store i32 %s.elt.elt6, ptr %s1.repack5, align 4
  %s1.repack7 = getelementptr inbounds nuw i8, ptr %s1, i32 8
  %s.elt.elt8 = extractvalue { ptr, i32, i32 } %s.elt, 2
  store i32 %s.elt.elt8, ptr %s1.repack7, align 4
  %s1.repack3 = getelementptr inbounds nuw i8, ptr %s1, i32 12
  %s.elt4 = extractvalue %main.sliceField %s, 1
  store i32 %s.elt4, ptr %s1.repack3, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load ptr, ptr %s1, align 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %s1, i32 4
  %.unpack10 = load i32, ptr %.elt9, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #4
  br i1 false, label %deref.throw, label %deref.next2

deref.next2:                                      ; preds = %deref.next
  %0 = getelementptr inbounds nuw i8, ptr %s1, i32 12
  %1 = load i32, ptr %0, align 4
  %2 = add i32 %.unpack10, %1
  ret i32 %2

deref.throw:                                      ; preds = %deref.next, %entry
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i32 undef
}

; Function Attrs: nounwind
define hidden i32 @main.ordinaryFields(i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %s = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %s, ptr nonnull %stackalloc, ptr undef) #4
  store i32 %s.len, ptr %s, align 4
  %s.repack2 = getelementptr inbounds nuw i8, ptr %s, i32 4
  store i32 %s.cap, ptr %s.repack2, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %0 = getelementptr inbounds nuw i8, ptr %s, i32 4
  %1 = load i32, ptr %s, align 4
  %2 = load i32, ptr %0, align 4
  %3 = add i32 %1, %2
  ret i32 %3

deref.throw:                                      ; preds = %deref.next, %entry
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i32 undef
}

; Function Attrs: nounwind
define hidden void @main.indirectResult(ptr dereferenceable_or_null(1025) %return, ptr %s.data, i32 range(i32 0, -2147483648) %s.len, i32 range(i32 0, -2147483648) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  store [1025 x i8] zeroinitializer, ptr %return, align 1
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.receiverArg(ptr %s.data, i32 range(i32 0, -2147483648) %s.len, i32 range(i32 0, -2147483648) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i32 @"(main.byteSlice).size"(ptr %s.data, i32 %s.len, i32 %s.cap, ptr undef)
  ret i32 %0
}

; Function Attrs: nounwind
define i32 @exportedArgs(ptr captures(none) %s.data, i32 %s.len, i32 %s.cap, ptr readonly captures(none) %text.data, i32 %text.len) #3 {
entry:
  %0 = add i32 %s.len, %s.cap
  %1 = add i32 %0, %text.len
  ret i32 %1
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-export-name"="exportedArgs" }
attributes #4 = { nounwind }
