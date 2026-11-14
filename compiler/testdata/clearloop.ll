; ModuleID = 'clearloop.go'
source_filename = "clearloop.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasip1"

%main.pair = type { i32, %runtime._string, %runtime._interface }
%runtime._string = type { ptr, i32 }
%runtime._interface = type { ptr, ptr }

@main.global = hidden global { ptr, i32, i32 } zeroinitializer, align 4

declare void @runtime.trackPointer(ptr readonly captures(none), ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @"(*main.rasterizer).rangeField"(ptr dereferenceable_or_null(24) %z, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp eq ptr %z, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %gep.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %.unpack = load ptr, ptr %z, align 4
  %.elt2 = getelementptr inbounds nuw i8, ptr %z, i32 4
  %.unpack3 = load i32, ptr %.elt2, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #5
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %deref.next
  %1 = shl i32 %.unpack3, 2
  call void @llvm.memset.p0.i32(ptr align 4 %.unpack, i8 0, i32 %1, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void

gep.throw:                                        ; preds = %gep.next, %entry
  call void @runtime.nilPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret void
}

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: write)
declare void @llvm.memset.p0.i32(ptr writeonly captures(none), i8, i32, i1 immarg) #2

; Function Attrs: nounwind
define hidden void @main.rangeGlobal(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %.unpack = load ptr, ptr @main.global, align 4
  %.unpack2 = load i32, ptr getelementptr inbounds nuw (i8, ptr @main.global, i32 4), align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #5
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %entry
  %0 = shl i32 %.unpack2, 2
  call void @llvm.memset.p0.i32(ptr align 4 %.unpack, i8 0, i32 %0, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.rangeLocal(ptr %s.data, i32 range(i32 0, 1073741824) %s.len, i32 range(i32 0, 1073741824) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %entry
  %0 = shl nuw i32 %s.len, 2
  call void @llvm.memset.p0.i32(ptr align 4 %s.data, i8 0, i32 %0, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.rangeBool(ptr %s.data, i32 range(i32 0, -2147483648) %s.len, i32 range(i32 0, -2147483648) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %entry
  call void @llvm.memset.p0.i32(ptr align 1 %s.data, i8 0, i32 %s.len, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.rangeIntField(ptr dereferenceable_or_null(24) %z, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp eq ptr %z, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %gep.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %.unpack = load ptr, ptr %z, align 4
  %.elt2 = getelementptr inbounds nuw i8, ptr %z, i32 4
  %.unpack3 = load i32, ptr %.elt2, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #5
  %1 = icmp sgt i32 %.unpack3, 0
  br i1 %1, label %rangeint.body, label %rangeint.done

rangeint.body:                                    ; preds = %deref.next
  %2 = shl i32 %.unpack3, 2
  call void @llvm.memset.p0.i32(ptr align 4 %.unpack, i8 0, i32 %2, i1 false)
  br label %rangeint.done

rangeint.done:                                    ; preds = %rangeint.body, %deref.next
  ret void

gep.throw:                                        ; preds = %gep.next, %entry
  call void @runtime.nilPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret void
}

; Function Attrs: nounwind
define hidden void @main.forField(ptr dereferenceable_or_null(24) %z, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  br label %for.loop

for.loop:                                         ; preds = %entry
  %0 = icmp eq ptr %z, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %for.loop
  br i1 false, label %gep.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %.unpack = load ptr, ptr %z, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %z, i32 4
  %.unpack2 = load i32, ptr %.elt1, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #5
  %1 = shl i32 %.unpack2, 2
  call void @llvm.memset.p0.i32(ptr align 4 %.unpack, i8 0, i32 %1, i1 false)
  br label %for.done

for.body:                                         ; No predecessors!
  unreachable

for.done:                                         ; preds = %deref.next
  ret void

gep.throw:                                        ; preds = %gep.next, %for.loop
  call void @runtime.nilPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret void
}

; Function Attrs: nounwind
define hidden void @main.forLocal(ptr %s.data, i32 range(i32 0, -2147483648) %s.len, i32 range(i32 0, -2147483648) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %for.loop

for.loop:                                         ; preds = %entry
  %0 = shl nuw i32 %s.len, 1
  call void @llvm.memset.p0.i32(ptr align 2 %s.data, i8 0, i32 %0, i1 false)
  br label %for.done

for.body:                                         ; No predecessors!
  unreachable

for.done:                                         ; preds = %for.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.structZero(ptr %s.data, i32 range(i32 0, 214748365) %s.len, i32 range(i32 0, 214748365) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %entry
  %0 = mul nuw i32 %s.len, 20
  call void @llvm.memset.p0.i32(ptr align 4 %s.data, i8 0, i32 %0, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.structZeroVar(ptr %s.data, i32 range(i32 0, 214748365) %s.len, i32 range(i32 0, 214748365) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %entry
  %0 = mul nuw i32 %s.len, 20
  call void @llvm.memset.p0.i32(ptr align 4 %s.data, i8 0, i32 %0, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.arrayZero(ptr %s.data, i32 range(i32 0, 268435456) %s.len, i32 range(i32 0, 268435456) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %entry
  %0 = shl nuw i32 %s.len, 4
  call void @llvm.memset.p0.i32(ptr align 4 %s.data, i8 0, i32 %0, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.complexZero(ptr %s.data, i32 range(i32 0, 268435456) %s.len, i32 range(i32 0, 268435456) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %entry
  %0 = shl nuw i32 %s.len, 4
  call void @llvm.memset.p0.i32(ptr align 8 %s.data, i8 0, i32 %0, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.stringZero(ptr %s.data, i32 range(i32 0, 536870912) %s.len, i32 range(i32 0, 536870912) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %entry
  %0 = shl nuw i32 %s.len, 3
  call void @llvm.memset.p0.i32(ptr align 4 %s.data, i8 0, i32 %0, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.interfaceZero(ptr %s.data, i32 range(i32 0, 536870912) %s.len, i32 range(i32 0, 536870912) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %entry
  %0 = shl nuw i32 %s.len, 3
  call void @llvm.memset.p0.i32(ptr align 4 %s.data, i8 0, i32 %0, i1 false)
  br label %rangeindex.done

rangeindex.body:                                  ; No predecessors!
  unreachable

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void
}

; Function Attrs: nounwind
define hidden void @main.nonZero(ptr %s.data, i32 range(i32 0, 1073741824) %s.len, i32 range(i32 0, 1073741824) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next, %entry
  %0 = phi i32 [ -1, %entry ], [ %1, %lookup.next ]
  %1 = add i32 %0, 1
  %2 = icmp ult i32 %1, %s.len
  br i1 %2, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  br i1 false, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %rangeindex.body
  %3 = getelementptr inbounds i32, ptr %s.data, i32 %1
  store i32 1, ptr %3, align 4
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void

lookup.throw:                                     ; preds = %rangeindex.body
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret void
}

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(inaccessiblemem: write)
declare void @llvm.assume(i1 noundef) #3

declare void @runtime.lookupPanic(ptr) #0

; Function Attrs: nounwind
define hidden void @main.extraStatement(ptr %s.data, i32 range(i32 0, 1073741824) %s.len, i32 range(i32 0, 1073741824) %s.cap, ptr %t.data, i32 range(i32 0, 1073741824) %t.len, i32 range(i32 0, 1073741824) %t.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next6, %entry
  %0 = phi i32 [ -1, %entry ], [ %1, %lookup.next6 ]
  %1 = add i32 %0, 1
  %2 = icmp ult i32 %1, %s.len
  br i1 %2, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  br i1 false, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %rangeindex.body
  %3 = getelementptr inbounds i32, ptr %s.data, i32 %1
  store i32 0, ptr %3, align 4
  %4 = icmp eq i32 %t.len, 0
  br i1 %4, label %lookup.throw, label %lookup.next3

lookup.next3:                                     ; preds = %lookup.next
  br i1 false, label %lookup.throw, label %lookup.next6

lookup.next6:                                     ; preds = %lookup.next3
  %5 = load i32, ptr %t.data, align 4
  %6 = add i32 %5, 1
  store i32 %6, ptr %t.data, align 4
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void

lookup.throw:                                     ; preds = %lookup.next3, %lookup.next, %rangeindex.body
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret void
}

declare { ptr, i32, i32 } @main.getSlice(ptr) #0

; Function Attrs: nounwind
define hidden void @main.fromCall(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call { ptr, i32, i32 } @main.getSlice(ptr undef) #5
  %1 = extractvalue { ptr, i32, i32 } %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %len = extractvalue { ptr, i32, i32 } %0, 1
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next, %entry
  %2 = phi i32 [ -1, %entry ], [ %3, %lookup.next ]
  %3 = add i32 %2, 1
  %4 = icmp ult i32 %3, %len
  br i1 %4, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  %5 = call { ptr, i32, i32 } @main.getSlice(ptr undef) #5
  %6 = extractvalue { ptr, i32, i32 } %5, 0
  call void @runtime.trackPointer(ptr %6, ptr nonnull %stackalloc, ptr undef) #5
  %indexaddr.len = extractvalue { ptr, i32, i32 } %5, 1
  %.not = icmp ult i32 %3, %indexaddr.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %rangeindex.body
  %indexaddr.ptr = extractvalue { ptr, i32, i32 } %5, 0
  %7 = getelementptr inbounds i32, ptr %indexaddr.ptr, i32 %3
  store i32 0, ptr %7, align 4
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void

lookup.throw:                                     ; preds = %rangeindex.body
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret void
}

; Function Attrs: nounwind
define hidden void @main.differentSlices(ptr dereferenceable_or_null(24) %z, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp eq ptr %z, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  br i1 false, label %gep.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %.unpack = load ptr, ptr %z, align 4
  %.elt3 = getelementptr inbounds nuw i8, ptr %z, i32 4
  %.unpack4 = load i32, ptr %.elt3, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #5
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next, %deref.next
  %1 = phi i32 [ -1, %deref.next ], [ %2, %lookup.next ]
  %2 = add i32 %1, 1
  %3 = icmp ult i32 %2, %.unpack4
  br i1 %3, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  br i1 false, label %gep.throw, label %gep.next1

gep.next1:                                        ; preds = %rangeindex.body
  br i1 false, label %gep.throw, label %deref.next2

deref.next2:                                      ; preds = %gep.next1
  %4 = getelementptr inbounds nuw i8, ptr %z, i32 12
  %.unpack7 = load ptr, ptr %4, align 4
  %.elt8 = getelementptr inbounds nuw i8, ptr %z, i32 16
  %.unpack9 = load i32, ptr %.elt8, align 4
  call void @runtime.trackPointer(ptr %.unpack7, ptr nonnull %stackalloc, ptr undef) #5
  %.not = icmp ult i32 %2, %.unpack9
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %deref.next2
  %5 = getelementptr inbounds float, ptr %.unpack7, i32 %2
  store float 0.000000e+00, ptr %5, align 4
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void

gep.throw:                                        ; preds = %gep.next1, %rangeindex.body, %gep.next, %entry
  call void @runtime.nilPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw, %gep.throw
  ret void

lookup.throw:                                     ; preds = %deref.next2
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return
}

; Function Attrs: nounwind
define hidden i32 @main.indexUsedAfter(ptr %s.data, i32 range(i32 0, 1073741824) %s.len, i32 range(i32 0, 1073741824) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next, %entry
  %0 = phi i32 [ 0, %entry ], [ %2, %lookup.next ]
  %1 = phi i32 [ -1, %entry ], [ %2, %lookup.next ]
  %2 = add i32 %1, 1
  %3 = icmp ult i32 %2, %s.len
  br i1 %3, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  br i1 false, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %rangeindex.body
  %4 = getelementptr inbounds i32, ptr %s.data, i32 %2
  store i32 0, ptr %4, align 4
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret i32 %0

lookup.throw:                                     ; preds = %rangeindex.body
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i32 undef
}

; Function Attrs: nounwind
define hidden void @main.stepTwo(ptr %s.data, i32 range(i32 0, 1073741824) %s.len, i32 range(i32 0, 1073741824) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %for.loop

for.loop:                                         ; preds = %lookup.next, %entry
  %0 = phi i32 [ 0, %entry ], [ %3, %lookup.next ]
  %1 = icmp slt i32 %0, %s.len
  br i1 %1, label %for.body, label %for.done

for.body:                                         ; preds = %for.loop
  %.not = icmp ult i32 %0, %s.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %for.body
  %2 = getelementptr inbounds i32, ptr %s.data, i32 %0
  store i32 0, ptr %2, align 4
  %3 = add i32 %0, 2
  br label %for.loop

for.done:                                         ; preds = %for.loop
  ret void

lookup.throw:                                     ; preds = %for.body
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret void
}

; Function Attrs: nounwind
define hidden void @main.zeroSize(ptr %s.data, i32 range(i32 0, -2147483648) %s.len, i32 range(i32 0, -2147483648) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next, %entry
  %0 = phi i32 [ -1, %entry ], [ %1, %lookup.next ]
  %1 = add i32 %0, 1
  %2 = icmp ult i32 %1, %s.len
  br i1 %2, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  br i1 false, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %rangeindex.body
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void

lookup.throw:                                     ; preds = %rangeindex.body
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret void
}

; Function Attrs: nounwind
define hidden void @main.explicitFields(ptr %s.data, i32 range(i32 0, 214748365) %s.len, i32 range(i32 0, 214748365) %s.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next, %entry
  %0 = phi i32 [ -1, %entry ], [ %1, %lookup.next ]
  %1 = add i32 %0, 1
  %2 = icmp ult i32 %1, %s.len
  br i1 %2, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  %complit = call align 4 dereferenceable(20) ptr @runtime.alloc(i32 20, ptr nonnull inttoptr (i32 1675 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #5
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %rangeindex.body
  store i32 0, ptr %complit, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack2.unpack = load ptr, ptr %.elt1, align 4
  %.unpack2.elt5 = getelementptr inbounds nuw i8, ptr %complit, i32 8
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.elt3 = getelementptr inbounds nuw i8, ptr %complit, i32 12
  %.unpack4.unpack = load ptr, ptr %.elt3, align 4
  %.unpack4.elt8 = getelementptr inbounds nuw i8, ptr %complit, i32 16
  %.unpack4.unpack9 = load ptr, ptr %.unpack4.elt8, align 4
  call void @runtime.trackPointer(ptr %.unpack2.unpack, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %.unpack4.unpack, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %.unpack4.unpack9, ptr nonnull %stackalloc, ptr undef) #5
  %.not = icmp ult i32 %1, %s.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %store.next
  %3 = getelementptr inbounds %main.pair, ptr %s.data, i32 %1
  store i32 0, ptr %3, align 4
  %.repack11 = getelementptr inbounds nuw i8, ptr %3, i32 4
  store ptr %.unpack2.unpack, ptr %.repack11, align 4
  %.repack11.repack15 = getelementptr inbounds nuw i8, ptr %3, i32 8
  store i32 %.unpack2.unpack6, ptr %.repack11.repack15, align 4
  %.repack13 = getelementptr inbounds nuw i8, ptr %3, i32 12
  store ptr %.unpack4.unpack, ptr %.repack13, align 4
  %.repack13.repack17 = getelementptr inbounds nuw i8, ptr %3, i32 16
  store ptr %.unpack4.unpack9, ptr %.repack13.repack17, align 4
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void

store.throw:                                      ; preds = %rangeindex.body
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw, %store.throw
  ret void

lookup.throw:                                     ; preds = %store.next
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #4

; Function Attrs: nounwind
define hidden void @main.storeBeforeLoop(ptr dereferenceable_or_null(12) %p, ptr dereferenceable_or_null(12) %q, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp eq ptr %p, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load ptr, ptr %p, align 4
  %.elt3 = getelementptr inbounds nuw i8, ptr %p, i32 4
  %.unpack4 = load i32, ptr %.elt3, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #5
  %1 = icmp eq ptr %q, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %.unpack7 = load ptr, ptr %q, align 4
  %.elt8 = getelementptr inbounds nuw i8, ptr %q, i32 4
  %.unpack9 = load i32, ptr %.elt8, align 4
  %.elt10 = getelementptr inbounds nuw i8, ptr %q, i32 8
  %.unpack11 = load i32, ptr %.elt10, align 4
  call void @runtime.trackPointer(ptr %.unpack7, ptr nonnull %stackalloc, ptr undef) #5
  br i1 false, label %deref.throw, label %store.next

store.next:                                       ; preds = %deref.next1
  store ptr %.unpack7, ptr %p, align 4
  %p.repack12 = getelementptr inbounds nuw i8, ptr %p, i32 4
  store i32 %.unpack9, ptr %p.repack12, align 4
  %p.repack14 = getelementptr inbounds nuw i8, ptr %p, i32 8
  store i32 %.unpack11, ptr %p.repack14, align 4
  br label %rangeindex.loop

rangeindex.loop:                                  ; preds = %lookup.next, %store.next
  %2 = phi i32 [ -1, %store.next ], [ %3, %lookup.next ]
  %3 = add i32 %2, 1
  %4 = icmp ult i32 %3, %.unpack4
  br i1 %4, label %rangeindex.body, label %rangeindex.done

rangeindex.body:                                  ; preds = %rangeindex.loop
  br i1 false, label %deref.throw, label %deref.next2

deref.next2:                                      ; preds = %rangeindex.body
  %.unpack16 = load ptr, ptr %p, align 4
  %.elt17 = getelementptr inbounds nuw i8, ptr %p, i32 4
  %.unpack18 = load i32, ptr %.elt17, align 4
  call void @runtime.trackPointer(ptr %.unpack16, ptr nonnull %stackalloc, ptr undef) #5
  %.not = icmp ult i32 %3, %.unpack18
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %deref.next2
  %5 = getelementptr inbounds i32, ptr %.unpack16, i32 %3
  store i32 0, ptr %5, align 4
  br label %rangeindex.loop

rangeindex.done:                                  ; preds = %rangeindex.loop
  ret void

deref.throw:                                      ; preds = %rangeindex.body, %deref.next1, %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw, %deref.throw
  ret void

lookup.throw:                                     ; preds = %deref.next2
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nocallback nofree nounwind willreturn memory(argmem: write) }
attributes #3 = { nocallback nofree nosync nounwind willreturn memory(inaccessiblemem: write) }
attributes #4 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #5 = { nounwind }
