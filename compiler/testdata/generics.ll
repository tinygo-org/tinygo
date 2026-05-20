; ModuleID = 'generics.go'
source_filename = "generics.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%"main.Point[float32]" = type { float, float }
%"main.Point[int]" = type { i32, i32 }

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.main(ptr %context) unnamed_addr #1 {
entry:
  %0 = call %"main.Point[float32]" @"main.Add[float32]"(float 0.000000e+00, float 0.000000e+00, float 0.000000e+00, float 0.000000e+00, ptr undef)
  %1 = call %"main.Point[int]" @"main.Add[int]"(i32 0, i32 0, i32 0, i32 0, ptr undef)
  ret void
}

; Function Attrs: nounwind
define linkonce_odr hidden %"main.Point[float32]" @"main.Add[float32]"(float %a.X, float %a.Y, float %b.X, float %b.Y, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %a = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %a, ptr nonnull %stackalloc, ptr undef) #3
  store float %a.X, ptr %a, align 4
  %a.repack9 = getelementptr inbounds nuw i8, ptr %a, i32 4
  store float %a.Y, ptr %a.repack9, align 4
  %b = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %b, ptr nonnull %stackalloc, ptr undef) #3
  store float %b.X, ptr %b, align 4
  %b.repack11 = getelementptr inbounds nuw i8, ptr %b, i32 4
  store float %b.Y, ptr %b.repack11, align 4
  call void @main.checkSize(i32 4, ptr undef) #3
  call void @main.checkSize(i32 8, ptr undef) #3
  %complit = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #3
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  br i1 false, label %deref.throw1, label %deref.next2

deref.next2:                                      ; preds = %deref.next
  %0 = load float, ptr %a, align 4
  %1 = load float, ptr %b, align 4
  %2 = fadd float %0, %1
  br i1 false, label %deref.throw3, label %deref.next4

deref.next4:                                      ; preds = %deref.next2
  br i1 false, label %deref.throw5, label %deref.next6

deref.next6:                                      ; preds = %deref.next4
  %3 = getelementptr inbounds nuw i8, ptr %b, i32 4
  %4 = getelementptr inbounds nuw i8, ptr %a, i32 4
  %5 = load float, ptr %4, align 4
  %6 = load float, ptr %3, align 4
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %deref.next6
  store float %2, ptr %complit, align 4
  br i1 false, label %store.throw7, label %store.next8

store.next8:                                      ; preds = %store.next
  %7 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %8 = fadd float %5, %6
  store float %8, ptr %7, align 4
  %.unpack = load float, ptr %complit, align 4
  %9 = insertvalue %"main.Point[float32]" poison, float %.unpack, 0
  %10 = insertvalue %"main.Point[float32]" %9, float %8, 1
  ret %"main.Point[float32]" %10

deref.throw:                                      ; preds = %entry
  unreachable

deref.throw1:                                     ; preds = %deref.next
  unreachable

deref.throw3:                                     ; preds = %deref.next2
  unreachable

deref.throw5:                                     ; preds = %deref.next4
  unreachable

store.throw:                                      ; preds = %deref.next6
  unreachable

store.throw7:                                     ; preds = %store.next
  unreachable
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #2

declare void @main.checkSize(i32, ptr) #0

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define linkonce_odr hidden %"main.Point[int]" @"main.Add[int]"(i32 %a.X, i32 %a.Y, i32 %b.X, i32 %b.Y, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %a = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %a, ptr nonnull %stackalloc, ptr undef) #3
  store i32 %a.X, ptr %a, align 4
  %a.repack9 = getelementptr inbounds nuw i8, ptr %a, i32 4
  store i32 %a.Y, ptr %a.repack9, align 4
  %b = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %b, ptr nonnull %stackalloc, ptr undef) #3
  store i32 %b.X, ptr %b, align 4
  %b.repack11 = getelementptr inbounds nuw i8, ptr %b, i32 4
  store i32 %b.Y, ptr %b.repack11, align 4
  call void @main.checkSize(i32 4, ptr undef) #3
  call void @main.checkSize(i32 8, ptr undef) #3
  %complit = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #3
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  br i1 false, label %deref.throw1, label %deref.next2

deref.next2:                                      ; preds = %deref.next
  %0 = load i32, ptr %a, align 4
  %1 = load i32, ptr %b, align 4
  %2 = add i32 %0, %1
  br i1 false, label %deref.throw3, label %deref.next4

deref.next4:                                      ; preds = %deref.next2
  br i1 false, label %deref.throw5, label %deref.next6

deref.next6:                                      ; preds = %deref.next4
  %3 = getelementptr inbounds nuw i8, ptr %b, i32 4
  %4 = getelementptr inbounds nuw i8, ptr %a, i32 4
  %5 = load i32, ptr %4, align 4
  %6 = load i32, ptr %3, align 4
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %deref.next6
  store i32 %2, ptr %complit, align 4
  br i1 false, label %store.throw7, label %store.next8

store.next8:                                      ; preds = %store.next
  %7 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %8 = add i32 %5, %6
  store i32 %8, ptr %7, align 4
  %.unpack = load i32, ptr %complit, align 4
  %9 = insertvalue %"main.Point[int]" poison, i32 %.unpack, 0
  %10 = insertvalue %"main.Point[int]" %9, i32 %8, 1
  ret %"main.Point[int]" %10

deref.throw:                                      ; preds = %entry
  unreachable

deref.throw1:                                     ; preds = %deref.next
  unreachable

deref.throw3:                                     ; preds = %deref.next2
  unreachable

deref.throw5:                                     ; preds = %deref.next4
  unreachable

store.throw:                                      ; preds = %deref.next6
  unreachable

store.throw7:                                     ; preds = %store.next
  unreachable
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nounwind }
