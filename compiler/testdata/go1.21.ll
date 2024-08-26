; ModuleID = 'go1.21.go'
source_filename = "go1.21.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.min1(i32 %a, ptr %context) unnamed_addr #2 {
entry:
  ret i32 %a
}

; Function Attrs: nounwind
define hidden i32 @main.min2(i32 %a, i32 %b, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i32 @llvm.smin.i32(i32 %a, i32 %b)
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i32 @main.min3(i32 %a, i32 %b, i32 %c, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i32 @llvm.smin.i32(i32 %a, i32 %b)
  %1 = call i32 @llvm.smin.i32(i32 %0, i32 %c)
  ret i32 %1
}

; Function Attrs: nounwind
define hidden i32 @main.min4(i32 %a, i32 %b, i32 %c, i32 %d, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i32 @llvm.smin.i32(i32 %a, i32 %b)
  %1 = call i32 @llvm.smin.i32(i32 %0, i32 %c)
  %2 = call i32 @llvm.smin.i32(i32 %1, i32 %d)
  ret i32 %2
}

; Function Attrs: nounwind
define hidden i8 @main.minUint8(i8 %a, i8 %b, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i8 @llvm.umin.i8(i8 %a, i8 %b)
  ret i8 %0
}

; Function Attrs: nounwind
define hidden i32 @main.minUnsigned(i32 %a, i32 %b, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i32 @llvm.umin.i32(i32 %a, i32 %b)
  ret i32 %0
}

; Function Attrs: nounwind
define hidden float @main.minFloat32(float %a, float %b, ptr %context) unnamed_addr #2 {
entry:
  %0 = fcmp olt float %a, %b
  %1 = select i1 %0, float %a, float %b
  ret float %1
}

; Function Attrs: nounwind
define hidden double @main.minFloat64(double %a, double %b, ptr %context) unnamed_addr #2 {
entry:
  %0 = fcmp olt double %a, %b
  %1 = select i1 %0, double %a, double %b
  ret double %1
}

; Function Attrs: nounwind
define hidden %runtime._string @main.minString(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr %context) unnamed_addr #2 {
entry:
  %0 = insertvalue %runtime._string zeroinitializer, ptr %a.data, 0
  %1 = insertvalue %runtime._string %0, i32 %a.len, 1
  %2 = insertvalue %runtime._string zeroinitializer, ptr %b.data, 0
  %3 = insertvalue %runtime._string %2, i32 %b.len, 1
  %stackalloc = alloca i8, align 1
  %4 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef) #5
  %5 = select i1 %4, %runtime._string %1, %runtime._string %3
  %6 = extractvalue %runtime._string %5, 0
  call void @runtime.trackPointer(ptr %6, ptr nonnull %stackalloc, ptr undef) #5
  ret %runtime._string %5
}

declare i1 @runtime.stringLess(ptr, i32, ptr, i32, ptr) #1

; Function Attrs: nounwind
define hidden i32 @main.maxInt(i32 %a, i32 %b, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i32 @llvm.smax.i32(i32 %a, i32 %b)
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i32 @main.maxUint(i32 %a, i32 %b, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i32 @llvm.umax.i32(i32 %a, i32 %b)
  ret i32 %0
}

; Function Attrs: nounwind
define hidden float @main.maxFloat32(float %a, float %b, ptr %context) unnamed_addr #2 {
entry:
  %0 = fcmp ogt float %a, %b
  %1 = select i1 %0, float %a, float %b
  ret float %1
}

; Function Attrs: nounwind
define hidden %runtime._string @main.maxString(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr %context) unnamed_addr #2 {
entry:
  %0 = insertvalue %runtime._string zeroinitializer, ptr %a.data, 0
  %1 = insertvalue %runtime._string %0, i32 %a.len, 1
  %2 = insertvalue %runtime._string zeroinitializer, ptr %b.data, 0
  %3 = insertvalue %runtime._string %2, i32 %b.len, 1
  %stackalloc = alloca i8, align 1
  %4 = call i1 @runtime.stringLess(ptr %b.data, i32 %b.len, ptr %a.data, i32 %a.len, ptr undef) #5
  %5 = select i1 %4, %runtime._string %1, %runtime._string %3
  %6 = extractvalue %runtime._string %5, 0
  call void @runtime.trackPointer(ptr %6, ptr nonnull %stackalloc, ptr undef) #5
  ret %runtime._string %5
}

; Function Attrs: nounwind
define hidden void @main.clearSlice(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #2 {
entry:
  %0 = shl i32 %s.len, 2
  call void @llvm.memset.p0.i32(ptr align 4 %s.data, i8 0, i32 %0, i1 false)
  ret void
}

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: write)
declare void @llvm.memset.p0.i32(ptr nocapture writeonly, i8, i32, i1 immarg) #3

; Function Attrs: nounwind
define hidden void @main.clearZeroSizedSlice(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.clearMap(ptr dereferenceable_or_null(40) %m, ptr %context) unnamed_addr #2 {
entry:
  call void @runtime.hashmapClear(ptr %m, ptr undef) #5
  ret void
}

declare void @runtime.hashmapClear(ptr dereferenceable_or_null(40), ptr) #1

; Function Attrs: nocallback nofree nosync nounwind speculatable willreturn memory(none)
declare i32 @llvm.smin.i32(i32, i32) #4

; Function Attrs: nocallback nofree nosync nounwind speculatable willreturn memory(none)
declare i8 @llvm.umin.i8(i8, i8) #4

; Function Attrs: nocallback nofree nosync nounwind speculatable willreturn memory(none)
declare i32 @llvm.umin.i32(i32, i32) #4

; Function Attrs: nocallback nofree nosync nounwind speculatable willreturn memory(none)
declare i32 @llvm.smax.i32(i32, i32) #4

; Function Attrs: nocallback nofree nosync nounwind speculatable willreturn memory(none)
declare i32 @llvm.umax.i32(i32, i32) #4

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nocallback nofree nounwind willreturn memory(argmem: write) }
attributes #4 = { nocallback nofree nosync nounwind speculatable willreturn memory(none) }
attributes #5 = { nounwind }
