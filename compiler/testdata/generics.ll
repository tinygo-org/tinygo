; ModuleID = 'generics.go'
source_filename = "generics.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%main.Point = type { %runtime._interface, %runtime._interface }
%runtime._interface = type { i32, i8* }
%main.Point.0 = type { float, float }

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*) #0

declare void @runtime.trackPointer(i8* nocapture readonly, i8*) #0

; Function Attrs: nounwind
define hidden void @main.init(i8* %context) unnamed_addr #1 {
entry:
  ret void
}

declare %main.Point @main.Add(%main.Point, %main.Point, i8*) #0

; Function Attrs: nounwind
define hidden void @main.main(i8* %context) unnamed_addr #1 {
entry:
  %bf = alloca %main.Point.0, align 8
  %af = alloca %main.Point.0, align 8
  %af.repack = getelementptr inbounds %main.Point.0, %main.Point.0* %af, i32 0, i32 0
  store float 0.000000e+00, float* %af.repack, align 8
  %af.repack1 = getelementptr inbounds %main.Point.0, %main.Point.0* %af, i32 0, i32 1
  store float 0.000000e+00, float* %af.repack1, align 4
  %0 = bitcast %main.Point.0* %af to i8*
  call void @runtime.trackPointer(i8* nonnull %0, i8* undef) #2
  %bf.repack = getelementptr inbounds %main.Point.0, %main.Point.0* %bf, i32 0, i32 0
  store float 0.000000e+00, float* %bf.repack, align 8
  %bf.repack2 = getelementptr inbounds %main.Point.0, %main.Point.0* %bf, i32 0, i32 1
  store float 0.000000e+00, float* %bf.repack2, align 4
  %1 = bitcast %main.Point.0* %bf to i8*
  call void @runtime.trackPointer(i8* nonnull %1, i8* undef) #2
  %.elt = getelementptr inbounds %main.Point.0, %main.Point.0* %af, i32 0, i32 0
  %.unpack = load float, float* %.elt, align 8
  %.elt3 = getelementptr inbounds %main.Point.0, %main.Point.0* %af, i32 0, i32 1
  %.unpack4 = load float, float* %.elt3, align 4
  %.elt5 = getelementptr inbounds %main.Point.0, %main.Point.0* %bf, i32 0, i32 0
  %.unpack6 = load float, float* %.elt5, align 8
  %.elt7 = getelementptr inbounds %main.Point.0, %main.Point.0* %bf, i32 0, i32 1
  %.unpack8 = load float, float* %.elt7, align 4
  %2 = call %main.Point.0 @"main.Add[float32]"(float %.unpack, float %.unpack4, float %.unpack6, float %.unpack8, i8* undef)
  ret void
}

; Function Attrs: nounwind
define linkonce_odr hidden %main.Point.0 @"main.Add[float32]"(float %a.X, float %a.Y, float %b.X, float %b.Y, i8* %context) unnamed_addr #1 {
entry:
  %complit = alloca %main.Point.0, align 8
  %b = alloca %main.Point.0, align 8
  %a = alloca %main.Point.0, align 8
  %a.repack = getelementptr inbounds %main.Point.0, %main.Point.0* %a, i32 0, i32 0
  store float 0.000000e+00, float* %a.repack, align 8
  %a.repack9 = getelementptr inbounds %main.Point.0, %main.Point.0* %a, i32 0, i32 1
  store float 0.000000e+00, float* %a.repack9, align 4
  %0 = bitcast %main.Point.0* %a to i8*
  call void @runtime.trackPointer(i8* nonnull %0, i8* undef) #2
  %a.repack10 = getelementptr inbounds %main.Point.0, %main.Point.0* %a, i32 0, i32 0
  store float %a.X, float* %a.repack10, align 8
  %a.repack11 = getelementptr inbounds %main.Point.0, %main.Point.0* %a, i32 0, i32 1
  store float %a.Y, float* %a.repack11, align 4
  %b.repack = getelementptr inbounds %main.Point.0, %main.Point.0* %b, i32 0, i32 0
  store float 0.000000e+00, float* %b.repack, align 8
  %b.repack13 = getelementptr inbounds %main.Point.0, %main.Point.0* %b, i32 0, i32 1
  store float 0.000000e+00, float* %b.repack13, align 4
  %1 = bitcast %main.Point.0* %b to i8*
  call void @runtime.trackPointer(i8* nonnull %1, i8* undef) #2
  %b.repack14 = getelementptr inbounds %main.Point.0, %main.Point.0* %b, i32 0, i32 0
  store float %b.X, float* %b.repack14, align 8
  %b.repack15 = getelementptr inbounds %main.Point.0, %main.Point.0* %b, i32 0, i32 1
  store float %b.Y, float* %b.repack15, align 4
  %complit.repack = getelementptr inbounds %main.Point.0, %main.Point.0* %complit, i32 0, i32 0
  store float 0.000000e+00, float* %complit.repack, align 8
  %complit.repack17 = getelementptr inbounds %main.Point.0, %main.Point.0* %complit, i32 0, i32 1
  store float 0.000000e+00, float* %complit.repack17, align 4
  %2 = bitcast %main.Point.0* %complit to i8*
  call void @runtime.trackPointer(i8* nonnull %2, i8* undef) #2
  %3 = getelementptr inbounds %main.Point.0, %main.Point.0* %complit, i32 0, i32 0
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  br i1 false, label %deref.throw1, label %deref.next2

deref.next2:                                      ; preds = %deref.next
  %4 = getelementptr inbounds %main.Point.0, %main.Point.0* %b, i32 0, i32 0
  %5 = getelementptr inbounds %main.Point.0, %main.Point.0* %a, i32 0, i32 0
  %6 = load float, float* %5, align 8
  %7 = load float, float* %4, align 8
  %8 = fadd float %6, %7
  br i1 false, label %deref.throw3, label %deref.next4

deref.next4:                                      ; preds = %deref.next2
  br i1 false, label %deref.throw5, label %deref.next6

deref.next6:                                      ; preds = %deref.next4
  %9 = getelementptr inbounds %main.Point.0, %main.Point.0* %b, i32 0, i32 1
  %10 = getelementptr inbounds %main.Point.0, %main.Point.0* %a, i32 0, i32 1
  %11 = load float, float* %10, align 4
  %12 = load float, float* %9, align 4
  br i1 false, label %store.throw, label %store.next

store.next:                                       ; preds = %deref.next6
  store float %8, float* %3, align 8
  br i1 false, label %store.throw7, label %store.next8

store.next8:                                      ; preds = %store.next
  %13 = getelementptr inbounds %main.Point.0, %main.Point.0* %complit, i32 0, i32 1
  %14 = fadd float %11, %12
  store float %14, float* %13, align 4
  %.elt = getelementptr inbounds %main.Point.0, %main.Point.0* %complit, i32 0, i32 0
  %.unpack = load float, float* %.elt, align 8
  %15 = insertvalue %main.Point.0 undef, float %.unpack, 0
  %16 = insertvalue %main.Point.0 %15, float %14, 1
  ret %main.Point.0 %16

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

declare void @runtime.nilPanic(i8*) #0

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind }
