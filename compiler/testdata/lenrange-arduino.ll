; ModuleID = 'lenrange.go'
source_filename = "lenrange.go"
target datalayout = "e-P1-p:16:8-i8:8-i16:8-i32:8-i64:8-f32:8-f64:8-n8:16-a:8"
target triple = "avr"

%main.sliceField = type { { ptr, i16, i16 }, i32 }

@llvm.used = appending global [1 x ptr] [ptr addrspacecast (ptr addrspace(1) @main.indirectResult to ptr)]
@tinygo.indirect-abi = appending global [1 x ptr] [ptr addrspacecast (ptr addrspace(1) @main.indirectResult to ptr)]

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @"(main.byteSlice).size"(ptr %s.data, i16 %s.len, i16 %s.cap, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %len.int = zext i16 %s.len to i32
  ret i32 %len.int
}

; Function Attrs: nounwind
define hidden i32 @main.sliceArgs(ptr %s.data, i16 %s.len, i16 %s.cap, i32 %n, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %len.int = zext i16 %s.len to i32
  %len.int1 = zext i16 %s.cap to i32
  %0 = add nuw nsw i32 %len.int, %len.int1
  %1 = add i32 %0, %n
  ret i32 %1
}

; Function Attrs: nounwind
define hidden i32 @main.stringArg(ptr readonly %s.data, i16 %s.len, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %len.int = zext i16 %s.len to i32
  ret i32 %len.int
}

; Function Attrs: nounwind
define hidden i32 @main.wideArgs(ptr %s.data, i16 range(i16 0, 16384) %s.len, i16 range(i16 0, 16384) %s.cap, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %narrow = add nuw nsw i16 %s.len, %s.cap
  %0 = zext nneg i16 %narrow to i32
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i32 @main.namedArgs(ptr %s.data, i16 %s.len, i16 %s.cap, ptr readonly %text.data, i16 %text.len, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %len.int = zext i16 %s.len to i32
  %len.int1 = zext i16 %s.cap to i32
  %0 = add nuw nsw i32 %len.int, %len.int1
  %len.int3 = zext i16 %text.len to i32
  %1 = add nuw nsw i32 %0, %len.int3
  ret i32 %1
}

; Function Attrs: nounwind
define hidden i32 @main.nestedArg(ptr readonly %s.text.data, i16 %s.text.len, i32 %s.n, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %s = call align 1 dereferenceable(8) addrspace(1) ptr @runtime.alloc(i16 8, ptr nonnull inttoptr (i16 49 to ptr), ptr undef) #0
  store ptr %s.text.data, ptr %s, align 1
  %s.repack4 = getelementptr inbounds nuw i8, ptr %s, i16 2
  store i16 %s.text.len, ptr %s.repack4, align 1
  %s.repack2 = getelementptr inbounds nuw i8, ptr %s, i16 4
  store i32 %s.n, ptr %s.repack2, align 1
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %0 = getelementptr inbounds nuw i8, ptr %s, i16 4
  %.elt6 = getelementptr inbounds nuw i8, ptr %s, i16 2
  %.unpack7 = load i16, ptr %.elt6, align 1
  %len.int = zext i16 %.unpack7 to i32
  %1 = load i32, ptr %0, align 1
  %2 = add i32 %1, %len.int
  ret i32 %2

deref.throw:                                      ; preds = %deref.next, %entry
  unreachable
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i16, ptr, ptr) addrspace(1) #1

declare void @runtime.nilPanic(ptr) addrspace(1)

; Function Attrs: nounwind
define hidden i32 @main.unexpandedArg(%main.sliceField %s, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %s1 = call align 1 dereferenceable(10) addrspace(1) ptr @runtime.alloc(i16 10, ptr nonnull inttoptr (i16 53 to ptr), ptr undef) #0
  %s.elt = extractvalue %main.sliceField %s, 0
  %s.elt.elt = extractvalue { ptr, i16, i16 } %s.elt, 0
  store ptr %s.elt.elt, ptr %s1, align 1
  %s1.repack5 = getelementptr inbounds nuw i8, ptr %s1, i16 2
  %s.elt.elt6 = extractvalue { ptr, i16, i16 } %s.elt, 1
  store i16 %s.elt.elt6, ptr %s1.repack5, align 1
  %s1.repack7 = getelementptr inbounds nuw i8, ptr %s1, i16 4
  %s.elt.elt8 = extractvalue { ptr, i16, i16 } %s.elt, 2
  store i16 %s.elt.elt8, ptr %s1.repack7, align 1
  %s1.repack3 = getelementptr inbounds nuw i8, ptr %s1, i16 6
  %s.elt4 = extractvalue %main.sliceField %s, 1
  store i32 %s.elt4, ptr %s1.repack3, align 1
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next2

deref.next2:                                      ; preds = %deref.next
  %0 = getelementptr inbounds nuw i8, ptr %s1, i16 6
  %.elt9 = getelementptr inbounds nuw i8, ptr %s1, i16 2
  %.unpack10 = load i16, ptr %.elt9, align 1
  %len.int = zext i16 %.unpack10 to i32
  %1 = load i32, ptr %0, align 1
  %2 = add i32 %1, %len.int
  ret i32 %2

deref.throw:                                      ; preds = %deref.next, %entry
  unreachable
}

; Function Attrs: nounwind
define hidden i32 @main.ordinaryFields(i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %s = call align 1 dereferenceable(8) addrspace(1) ptr @runtime.alloc(i16 8, ptr nonnull inttoptr (i16 3 to ptr), ptr undef) #0
  store i32 %s.len, ptr %s, align 1
  %s.repack2 = getelementptr inbounds nuw i8, ptr %s, i16 4
  store i32 %s.cap, ptr %s.repack2, align 1
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %0 = getelementptr inbounds nuw i8, ptr %s, i16 4
  %1 = load i32, ptr %s, align 1
  %2 = load i32, ptr %0, align 1
  %3 = add i32 %1, %2
  ret i32 %3

deref.throw:                                      ; preds = %deref.next, %entry
  unreachable
}

; Function Attrs: nounwind
define hidden void @main.indirectResult(ptr dereferenceable_or_null(1025) %return, ptr %s.data, i16 %s.len, i16 %s.cap, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  store [1025 x i8] zeroinitializer, ptr %return, align 1
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.receiverArg(ptr %s.data, i16 %s.len, i16 %s.cap, ptr %context) unnamed_addr addrspace(1) #0 {
entry:
  %0 = call addrspace(1) i32 @"(main.byteSlice).size"(ptr %s.data, i16 %s.len, i16 %s.cap, ptr undef)
  ret i32 %0
}

; Function Attrs: nounwind
define i32 @exportedArgs(ptr captures(none) %s.data, i16 %s.len, i16 %s.cap, ptr readonly captures(none) %text.data, i16 %text.len) addrspace(1) #0 {
entry:
  %len.int = zext i16 %s.len to i32
  %len.int1 = zext i16 %s.cap to i32
  %0 = add nuw nsw i32 %len.int, %len.int1
  %len.int3 = zext i16 %text.len to i32
  %1 = add nuw nsw i32 %0, %len.int3
  ret i32 %1
}

attributes #0 = { nounwind }
attributes #1 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" }
