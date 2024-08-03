; ModuleID = 'slice.go'
source_filename = "slice.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

; Function Attrs: nounwind
declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.sliceLen(ptr %ints.data, i32 %ints.len, i32 %ints.cap, ptr %context) unnamed_addr #1 {
entry:
  ret i32 %ints.len
}

; Function Attrs: nounwind
define hidden i32 @main.sliceCap(ptr %ints.data, i32 %ints.len, i32 %ints.cap, ptr %context) unnamed_addr #1 {
entry:
  ret i32 %ints.cap
}

; Function Attrs: nounwind
define hidden i32 @main.sliceElement(ptr %ints.data, i32 %ints.len, i32 %ints.cap, i32 %index, ptr %context) unnamed_addr #1 {
entry:
  %.not = icmp ult i32 %index, %ints.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %entry
  %0 = getelementptr inbounds i32, ptr %ints.data, i32 %index
  %1 = load i32, ptr %0, align 4
  ret i32 %1

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #3
  unreachable
}

declare void @runtime.lookupPanic(ptr) #2

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.sliceAppendValues(ptr %ints.data, i32 %ints.len, i32 %ints.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %varargs = call align 4 dereferenceable(12) ptr @runtime.alloc(i32 12, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %varargs, ptr nonnull %stackalloc, ptr undef)
  store i32 1, ptr %varargs, align 4
  %0 = getelementptr inbounds [3 x i32], ptr %varargs, i32 0, i32 1
  store i32 2, ptr %0, align 4
  %1 = getelementptr inbounds [3 x i32], ptr %varargs, i32 0, i32 2
  store i32 3, ptr %1, align 4
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %ints.data, ptr nonnull %varargs, i32 %ints.len, i32 %ints.cap, i32 3, i32 4, ptr undef) #3
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  %2 = insertvalue { ptr, i32, i32 } undef, ptr %append.newPtr, 0
  %3 = insertvalue { ptr, i32, i32 } %2, i32 %append.newLen, 1
  %4 = insertvalue { ptr, i32, i32 } %3, i32 %append.newCap, 2
  call void @runtime.trackPointer(ptr %append.newPtr, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %4
}

declare { ptr, i32, i32 } @runtime.sliceAppend(ptr, ptr nocapture readonly, i32, i32, i32, i32, ptr) #2

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.sliceAppendSlice(ptr %ints.data, i32 %ints.len, i32 %ints.cap, ptr %added.data, i32 %added.len, i32 %added.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %append.new = call { ptr, i32, i32 } @runtime.sliceAppend(ptr %ints.data, ptr %added.data, i32 %ints.len, i32 %ints.cap, i32 %added.len, i32 4, ptr undef) #3
  %append.newPtr = extractvalue { ptr, i32, i32 } %append.new, 0
  %append.newLen = extractvalue { ptr, i32, i32 } %append.new, 1
  %append.newCap = extractvalue { ptr, i32, i32 } %append.new, 2
  %0 = insertvalue { ptr, i32, i32 } undef, ptr %append.newPtr, 0
  %1 = insertvalue { ptr, i32, i32 } %0, i32 %append.newLen, 1
  %2 = insertvalue { ptr, i32, i32 } %1, i32 %append.newCap, 2
  call void @runtime.trackPointer(ptr %append.newPtr, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %2
}

; Function Attrs: nounwind
define hidden i32 @main.sliceCopy(ptr %dst.data, i32 %dst.len, i32 %dst.cap, ptr %src.data, i32 %src.len, i32 %src.cap, ptr %context) unnamed_addr #1 {
entry:
  %copy.n = call i32 @runtime.sliceCopy(ptr %dst.data, ptr %src.data, i32 %dst.len, i32 %src.len, i32 4, ptr undef) #3
  ret i32 %copy.n
}

declare i32 @runtime.sliceCopy(ptr nocapture writeonly, ptr nocapture readonly, i32, i32, i32, ptr) #2

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.makeByteSlice(i32 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.maxcap = icmp slt i32 %len, 0
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %makeslice.buf = call align 1 ptr @runtime.alloc(i32 %len, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  %0 = insertvalue { ptr, i32, i32 } undef, ptr %makeslice.buf, 0
  %1 = insertvalue { ptr, i32, i32 } %0, i32 %len, 1
  %2 = insertvalue { ptr, i32, i32 } %1, i32 %len, 2
  call void @runtime.trackPointer(ptr nonnull %makeslice.buf, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %2

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #3
  unreachable
}

declare void @runtime.slicePanic(ptr) #2

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.makeInt16Slice(i32 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.maxcap = icmp slt i32 %len, 0
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %makeslice.cap = shl nuw i32 %len, 1
  %makeslice.buf = call align 2 ptr @runtime.alloc(i32 %makeslice.cap, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  %0 = insertvalue { ptr, i32, i32 } undef, ptr %makeslice.buf, 0
  %1 = insertvalue { ptr, i32, i32 } %0, i32 %len, 1
  %2 = insertvalue { ptr, i32, i32 } %1, i32 %len, 2
  call void @runtime.trackPointer(ptr nonnull %makeslice.buf, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %2

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #3
  unreachable
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.makeArraySlice(i32 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.maxcap = icmp ugt i32 %len, 1431655765
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %makeslice.cap = mul i32 %len, 3
  %makeslice.buf = call align 1 ptr @runtime.alloc(i32 %makeslice.cap, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  %0 = insertvalue { ptr, i32, i32 } undef, ptr %makeslice.buf, 0
  %1 = insertvalue { ptr, i32, i32 } %0, i32 %len, 1
  %2 = insertvalue { ptr, i32, i32 } %1, i32 %len, 2
  call void @runtime.trackPointer(ptr nonnull %makeslice.buf, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %2

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #3
  unreachable
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.makeInt32Slice(i32 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.maxcap = icmp ugt i32 %len, 1073741823
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %makeslice.cap = shl nuw i32 %len, 2
  %makeslice.buf = call align 4 ptr @runtime.alloc(i32 %makeslice.cap, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  %0 = insertvalue { ptr, i32, i32 } undef, ptr %makeslice.buf, 0
  %1 = insertvalue { ptr, i32, i32 } %0, i32 %len, 1
  %2 = insertvalue { ptr, i32, i32 } %1, i32 %len, 2
  call void @runtime.trackPointer(ptr nonnull %makeslice.buf, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %2

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #3
  unreachable
}

; Function Attrs: nounwind
define hidden ptr @main.Add32(ptr %p, i32 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = getelementptr i8, ptr %p, i32 %len
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef)
  ret ptr %0
}

; Function Attrs: nounwind
define hidden ptr @main.Add64(ptr %p, i64 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = trunc i64 %len to i32
  %1 = getelementptr i8, ptr %p, i32 %0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef)
  ret ptr %1
}

; Function Attrs: nounwind
define hidden ptr @main.SliceToArray(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp ult i32 %s.len, 4
  br i1 %0, label %slicetoarray.throw, label %slicetoarray.next

slicetoarray.next:                                ; preds = %entry
  ret ptr %s.data

slicetoarray.throw:                               ; preds = %entry
  call void @runtime.sliceToArrayPointerPanic(ptr undef) #3
  unreachable
}

declare void @runtime.sliceToArrayPointerPanic(ptr) #2

; Function Attrs: nounwind
define hidden ptr @main.SliceToArrayConst(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %makeslice = call align 4 dereferenceable(24) ptr @runtime.alloc(i32 24, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %makeslice, ptr nonnull %stackalloc, ptr undef)
  br i1 false, label %slicetoarray.throw, label %slicetoarray.next

slicetoarray.next:                                ; preds = %entry
  ret ptr %makeslice

slicetoarray.throw:                               ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.SliceInt(ptr dereferenceable_or_null(4) %ptr, i32 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp ugt i32 %len, 1073741823
  %1 = icmp eq ptr %ptr, null
  %2 = icmp ne i32 %len, 0
  %3 = and i1 %1, %2
  %4 = or i1 %3, %0
  br i1 %4, label %unsafe.Slice.throw, label %unsafe.Slice.next

unsafe.Slice.next:                                ; preds = %entry
  %5 = insertvalue { ptr, i32, i32 } undef, ptr %ptr, 0
  %6 = insertvalue { ptr, i32, i32 } %5, i32 %len, 1
  %7 = insertvalue { ptr, i32, i32 } %6, i32 %len, 2
  call void @runtime.trackPointer(ptr %ptr, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %7

unsafe.Slice.throw:                               ; preds = %entry
  call void @runtime.unsafeSlicePanic(ptr undef) #3
  unreachable
}

declare void @runtime.unsafeSlicePanic(ptr) #2

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.SliceUint16(ptr dereferenceable_or_null(1) %ptr, i16 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp eq ptr %ptr, null
  %1 = icmp ne i16 %len, 0
  %2 = and i1 %0, %1
  br i1 %2, label %unsafe.Slice.throw, label %unsafe.Slice.next

unsafe.Slice.next:                                ; preds = %entry
  %3 = zext i16 %len to i32
  %4 = insertvalue { ptr, i32, i32 } undef, ptr %ptr, 0
  %5 = insertvalue { ptr, i32, i32 } %4, i32 %3, 1
  %6 = insertvalue { ptr, i32, i32 } %5, i32 %3, 2
  call void @runtime.trackPointer(ptr %ptr, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %6

unsafe.Slice.throw:                               ; preds = %entry
  call void @runtime.unsafeSlicePanic(ptr undef) #3
  unreachable
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.SliceUint64(ptr dereferenceable_or_null(4) %ptr, i64 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp ugt i64 %len, 1073741823
  %1 = icmp eq ptr %ptr, null
  %2 = icmp ne i64 %len, 0
  %3 = and i1 %1, %2
  %4 = or i1 %3, %0
  br i1 %4, label %unsafe.Slice.throw, label %unsafe.Slice.next

unsafe.Slice.next:                                ; preds = %entry
  %5 = trunc i64 %len to i32
  %6 = insertvalue { ptr, i32, i32 } undef, ptr %ptr, 0
  %7 = insertvalue { ptr, i32, i32 } %6, i32 %5, 1
  %8 = insertvalue { ptr, i32, i32 } %7, i32 %5, 2
  call void @runtime.trackPointer(ptr %ptr, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %8

unsafe.Slice.throw:                               ; preds = %entry
  call void @runtime.unsafeSlicePanic(ptr undef) #3
  unreachable
}

; Function Attrs: nounwind
define hidden { ptr, i32, i32 } @main.SliceInt64(ptr dereferenceable_or_null(4) %ptr, i64 %len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp ugt i64 %len, 1073741823
  %1 = icmp eq ptr %ptr, null
  %2 = icmp ne i64 %len, 0
  %3 = and i1 %1, %2
  %4 = or i1 %3, %0
  br i1 %4, label %unsafe.Slice.throw, label %unsafe.Slice.next

unsafe.Slice.next:                                ; preds = %entry
  %5 = trunc i64 %len to i32
  %6 = insertvalue { ptr, i32, i32 } undef, ptr %ptr, 0
  %7 = insertvalue { ptr, i32, i32 } %6, i32 %5, 1
  %8 = insertvalue { ptr, i32, i32 } %7, i32 %5, 2
  call void @runtime.trackPointer(ptr %ptr, ptr nonnull %stackalloc, ptr undef)
  ret { ptr, i32, i32 } %8

unsafe.Slice.throw:                               ; preds = %entry
  call void @runtime.unsafeSlicePanic(ptr undef) #3
  unreachable
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+exception-handling,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+exception-handling,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { "target-features"="+bulk-memory,+exception-handling,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #3 = { nounwind }
