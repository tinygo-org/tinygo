; ModuleID = 'volatile.go'
source_filename = "volatile.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasip1"

declare void @runtime.trackPointer(ptr readonly captures(none), ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i8 @main.load8(ptr dereferenceable_or_null(1) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = load volatile i8, ptr %p, align 1
  ret i8 %1

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i8 undef
}

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define hidden i32 @main.load32(ptr dereferenceable_or_null(4) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = load volatile i32, ptr %p, align 4
  ret i32 %1

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i32 undef
}

; Function Attrs: nounwind
define hidden void @main.store8(ptr dereferenceable_or_null(1) %p, i8 %v, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  store volatile i8 %v, ptr %p, align 1
  ret void

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret void
}

; Function Attrs: nounwind
define hidden void @main.store32(ptr dereferenceable_or_null(4) %p, i32 %v, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  store volatile i32 %v, ptr %p, align 4
  ret void

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.regGet(ptr dereferenceable_or_null(8) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = load volatile i32, ptr %p, align 4
  ret i32 %1

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret i32 undef
}

; Function Attrs: nounwind
define hidden void @main.regSet(ptr dereferenceable_or_null(8) %p, i32 %v, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  store volatile i32 %v, ptr %p, align 4
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret void
}

; Function Attrs: nounwind
define hidden void @main.regSetBits(ptr dereferenceable_or_null(8) %p, i32 %v, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = load volatile i32, ptr %p, align 4
  %2 = or i32 %1, %v
  store volatile i32 %2, ptr %p, align 4
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret void
}

; Function Attrs: nounwind
define hidden void @main.regClearBits(ptr dereferenceable_or_null(8) %p, i32 %v, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = load volatile i32, ptr %p, align 4
  %2 = xor i32 %v, -1
  %3 = and i32 %1, %2
  store volatile i32 %3, ptr %p, align 4
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret void
}

; Function Attrs: nounwind
define hidden i1 @main.regHasBits(ptr dereferenceable_or_null(8) %p, i8 %v, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = getelementptr inbounds nuw i8, ptr %p, i32 4
  %2 = load volatile i8, ptr %1, align 1
  %3 = and i8 %2, %v
  %4 = icmp ne i8 %3, 0
  ret i1 %4

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden void @main.regReplaceBits(ptr dereferenceable_or_null(8) %p, i32 %v, i32 %mask, i8 %pos, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = load volatile i32, ptr %p, align 4
  %shift.overflow = icmp ugt i8 %pos, 31
  %2 = zext nneg i8 %pos to i32
  %3 = shl i32 %mask, %2
  %shift.overflow1 = icmp ugt i8 %pos, 31
  %4 = zext nneg i8 %pos to i32
  %5 = shl i32 %v, %4
  %shift.result2 = select i1 %shift.overflow1, i32 0, i32 %5
  %6 = xor i32 %3, -1
  %7 = select i1 %shift.overflow, i32 -1, i32 %6
  %8 = and i32 %1, %7
  %9 = or i32 %8, %shift.result2
  store volatile i32 %9, ptr %p, align 4
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret void
}

; Function Attrs: nounwind
define hidden void @main.regPointerSet(ptr dereferenceable_or_null(4) %r, i32 %v, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %r, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  store volatile i32 %v, ptr %r, align 4
  ret void

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret void
}

; Function Attrs: nounwind
define hidden { ptr, ptr } @main.regMethodValue(ptr dereferenceable_or_null(8) %p, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %1 = insertvalue { ptr, ptr } undef, ptr %p, 0
  %2 = insertvalue { ptr, ptr } %1, ptr @"(*runtime/volatile.Register32).Set$bound", 1
  call void @runtime.trackPointer(ptr nonnull %p, ptr nonnull %stackalloc, ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull @"(*runtime/volatile.Register32).Set$bound", ptr nonnull %stackalloc, ptr undef) #2
  ret { ptr, ptr } %2

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret { ptr, ptr } undef
}

; Function Attrs: nounwind
define linkonce_odr hidden void @"(*runtime/volatile.Register32).Set$bound"(i32 %value, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %context, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  store volatile i32 %value, ptr %context, align 4
  ret void

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret void
}

; Function Attrs: nounwind
define hidden { ptr, ptr } @main.regPointerMethodValue(ptr dereferenceable_or_null(4) %r, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = insertvalue { ptr, ptr } undef, ptr %r, 0
  %1 = insertvalue { ptr, ptr } %0, ptr @"(*runtime/volatile.Register32).Set$bound", 1
  call void @runtime.trackPointer(ptr %r, ptr nonnull %stackalloc, ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull @"(*runtime/volatile.Register32).Set$bound", ptr nonnull %stackalloc, ptr undef) #2
  ret { ptr, ptr } %1
}

; Function Attrs: nounwind
define hidden void @main.regCallMethodValue(ptr dereferenceable_or_null(8) %p, i32 %v, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp eq ptr %p, null
  br i1 %0, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  call void @runtime.trackPointer(ptr nonnull %p, ptr nonnull %stackalloc, ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull @"(*runtime/volatile.Register32).Set$bound", ptr nonnull %stackalloc, ptr undef) #2
  call void @"(*runtime/volatile.Register32).Set$bound"(i32 %v, ptr nonnull %p)
  ret void

gep.throw:                                        ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %gep.throw
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.load32Alias(ptr dereferenceable_or_null(4) %p, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %p, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = load volatile i32, ptr %p, align 4
  ret i32 %1

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i32 undef
}

; Function Attrs: nounwind
define hidden i32 @main.regGetAlias(ptr dereferenceable_or_null(4) %r, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %r, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = load volatile i32, ptr %r, align 4
  ret i32 %1

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i32 undef
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind }
