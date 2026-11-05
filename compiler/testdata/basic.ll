; ModuleID = 'basic.go'
source_filename = "basic.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%main.plainStruct = type { [16 x i32], ptr }
%main.kv = type { float, i32, i32, i32 }
%main.kv.0 = type { i8, i32, i32, i32 }

@cmpconst = internal unnamed_addr constant %main.plainStruct zeroinitializer, align 4
@main.kvGlobal = hidden global %main.kv zeroinitializer, align 4
@main.a = hidden global { ptr, i32, i32 } zeroinitializer, align 4
@main.b = hidden global [2 x ptr] zeroinitializer, align 4

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.addInt(i32 %x, i32 %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = add i32 %x, %y
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i1 @main.equalInt(i32 %x, i32 %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq i32 %x, %y
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.equalByteArray(ptr dereferenceable_or_null(4096) %x, ptr dereferenceable_or_null(4096) %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = icmp eq ptr %y, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %arraycmp = call i1 @runtime.memequal(ptr nonnull %x, ptr nonnull %y, i32 4096, ptr undef) #3
  ret i1 %arraycmp

deref.throw:                                      ; preds = %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

declare void @runtime.nilPanic(ptr) #0

declare i1 @runtime.memequal(ptr, ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.notEqualByteArray(ptr dereferenceable_or_null(4096) %x, ptr dereferenceable_or_null(4096) %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = icmp eq ptr %y, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %arraycmp = call i1 @runtime.memequal(ptr nonnull %x, ptr nonnull %y, i32 4096, ptr undef) #3
  %2 = xor i1 %arraycmp, true
  ret i1 %2

deref.throw:                                      ; preds = %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.equalPlainStruct(ptr dereferenceable_or_null(68) %x, ptr dereferenceable_or_null(68) %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = icmp eq ptr %y, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %arraycmp = call i1 @runtime.memequal(ptr nonnull %x, ptr nonnull %y, i32 68, ptr undef) #3
  ret i1 %arraycmp

deref.throw:                                      ; preds = %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.isZeroPlainStruct(ptr dereferenceable_or_null(68) %x, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %arraycmp = call i1 @runtime.memequal(ptr nonnull %x, ptr nonnull @cmpconst, i32 68, ptr undef) #3
  ret i1 %arraycmp

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.equalMutatedByteArray(ptr dereferenceable_or_null(5) %x, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load i8, ptr %x, align 1
  %.elt2 = getelementptr inbounds nuw i8, ptr %x, i32 1
  %.unpack3 = load i8, ptr %.elt2, align 1
  %.elt4 = getelementptr inbounds nuw i8, ptr %x, i32 2
  %.unpack5 = load i8, ptr %.elt4, align 1
  %.elt6 = getelementptr inbounds nuw i8, ptr %x, i32 3
  %.unpack7 = load i8, ptr %.elt6, align 1
  %.elt8 = getelementptr inbounds nuw i8, ptr %x, i32 4
  %.unpack9 = load i8, ptr %.elt8, align 1
  br i1 false, label %deref.throw, label %gep.next

gep.next:                                         ; preds = %deref.next
  store i8 1, ptr %x, align 1
  br i1 false, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %gep.next
  %t0 = call align 1 dereferenceable(5) ptr @runtime.alloc(i32 5, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %t0, ptr nonnull %stackalloc, ptr undef) #3
  store i8 %.unpack, ptr %t0, align 1
  %t0.repack10 = getelementptr inbounds nuw i8, ptr %t0, i32 1
  store i8 %.unpack3, ptr %t0.repack10, align 1
  %t0.repack12 = getelementptr inbounds nuw i8, ptr %t0, i32 2
  store i8 %.unpack5, ptr %t0.repack12, align 1
  %t0.repack14 = getelementptr inbounds nuw i8, ptr %t0, i32 3
  store i8 %.unpack7, ptr %t0.repack14, align 1
  %t0.repack16 = getelementptr inbounds nuw i8, ptr %t0, i32 4
  store i8 %.unpack9, ptr %t0.repack16, align 1
  %arraycmp = call i1 @runtime.memequal(ptr nonnull %t0, ptr nonnull %x, i32 5, ptr undef) #3
  ret i1 %arraycmp

deref.throw:                                      ; preds = %gep.next, %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #2

; Function Attrs: nounwind
define hidden i32 @main.divInt(i32 %x, i32 %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq i32 %y, 0
  br i1 %0, label %divbyzero.throw, label %divbyzero.next

divbyzero.next:                                   ; preds = %entry
  %1 = icmp eq i32 %y, -1
  %2 = icmp eq i32 %x, -2147483648
  %3 = and i1 %1, %2
  %4 = select i1 %3, i32 1, i32 %y
  %5 = sdiv i32 %x, %4
  ret i32 %5

divbyzero.throw:                                  ; preds = %entry
  call void @runtime.divideByZeroPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %divbyzero.throw
  ret i32 undef
}

declare void @runtime.divideByZeroPanic(ptr) #0

; Function Attrs: nounwind
define hidden i32 @main.divUint(i32 %x, i32 %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq i32 %y, 0
  br i1 %0, label %divbyzero.throw, label %divbyzero.next

divbyzero.next:                                   ; preds = %entry
  %1 = udiv i32 %x, %y
  ret i32 %1

divbyzero.throw:                                  ; preds = %entry
  call void @runtime.divideByZeroPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %divbyzero.throw
  ret i32 undef
}

; Function Attrs: nounwind
define hidden i32 @main.remInt(i32 %x, i32 %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq i32 %y, 0
  br i1 %0, label %divbyzero.throw, label %divbyzero.next

divbyzero.next:                                   ; preds = %entry
  %1 = icmp eq i32 %y, -1
  %2 = icmp eq i32 %x, -2147483648
  %3 = and i1 %1, %2
  %4 = select i1 %3, i32 1, i32 %y
  %5 = srem i32 %x, %4
  ret i32 %5

divbyzero.throw:                                  ; preds = %entry
  call void @runtime.divideByZeroPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %divbyzero.throw
  ret i32 undef
}

; Function Attrs: nounwind
define hidden i32 @main.remUint(i32 %x, i32 %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq i32 %y, 0
  br i1 %0, label %divbyzero.throw, label %divbyzero.next

divbyzero.next:                                   ; preds = %entry
  %1 = urem i32 %x, %y
  ret i32 %1

divbyzero.throw:                                  ; preds = %entry
  call void @runtime.divideByZeroPanic(ptr undef) #3
  br label %unwind.return

unwind.return:                                    ; preds = %divbyzero.throw
  ret i32 undef
}

; Function Attrs: nounwind
define hidden i1 @main.floatEQ(float %x, float %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = fcmp oeq float %x, %y
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.floatNE(float %x, float %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = fcmp une float %x, %y
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.floatLower(float %x, float %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = fcmp olt float %x, %y
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.floatLowerEqual(float %x, float %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = fcmp ole float %x, %y
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.floatGreater(float %x, float %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = fcmp ogt float %x, %y
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.floatGreaterEqual(float %x, float %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = fcmp oge float %x, %y
  ret i1 %0
}

; Function Attrs: nounwind
define hidden float @main.complexReal(float %x.r, float %x.i, ptr %context) unnamed_addr #1 {
entry:
  ret float %x.r
}

; Function Attrs: nounwind
define hidden float @main.complexImag(float %x.r, float %x.i, ptr %context) unnamed_addr #1 {
entry:
  ret float %x.i
}

; Function Attrs: nounwind
define hidden { float, float } @main.complexAdd(float %x.r, float %x.i, float %y.r, float %y.i, ptr %context) unnamed_addr #1 {
entry:
  %0 = fadd float %x.r, %y.r
  %1 = fadd float %x.i, %y.i
  %2 = insertvalue { float, float } undef, float %0, 0
  %3 = insertvalue { float, float } %2, float %1, 1
  ret { float, float } %3
}

; Function Attrs: nounwind
define hidden { float, float } @main.complexSub(float %x.r, float %x.i, float %y.r, float %y.i, ptr %context) unnamed_addr #1 {
entry:
  %0 = fsub float %x.r, %y.r
  %1 = fsub float %x.i, %y.i
  %2 = insertvalue { float, float } undef, float %0, 0
  %3 = insertvalue { float, float } %2, float %1, 1
  ret { float, float } %3
}

; Function Attrs: nounwind
define hidden i64 @main.shiftNested(i64 %x, ptr %context) unnamed_addr #1 {
entry:
  %0 = lshr i64 %x, 8
  ret i64 %0
}

; Function Attrs: nounwind
define hidden { float, float } @main.complexMul(float %x.r, float %x.i, float %y.r, float %y.i, ptr %context) unnamed_addr #1 {
entry:
  %0 = fmul float %x.r, %y.r
  %1 = fmul float %x.i, %y.i
  %2 = fsub float %0, %1
  %3 = fmul float %x.r, %y.i
  %4 = fmul float %x.i, %y.r
  %5 = fadd float %3, %4
  %6 = insertvalue { float, float } undef, float %2, 0
  %7 = insertvalue { float, float } %6, float %5, 1
  ret { float, float } %7
}

; Function Attrs: nounwind
define hidden void @main.foo(ptr %context) unnamed_addr #1 {
entry:
  call void @"main.foo$1"(%main.kv.0 zeroinitializer, ptr undef)
  ret void
}

; Function Attrs: nounwind
define internal void @"main.foo$1"(%main.kv.0 %b, ptr %context) unnamed_addr #1 {
entry:
  ret void
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nounwind }
