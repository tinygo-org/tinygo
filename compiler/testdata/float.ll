; ModuleID = 'float.go'
source_filename = "float.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.f32tou32(float %v, ptr %context) unnamed_addr #2 {
entry:
  %positive = fcmp oge float %v, 0.000000e+00
  %withinmax = fcmp ole float %v, 0x41EFFFFFC0000000
  %inbounds = and i1 %positive, %withinmax
  %saturated = sext i1 %positive to i32
  %normal = fptoui float %v to i32
  %0 = select i1 %inbounds, i32 %normal, i32 %saturated
  ret i32 %0
}

; Function Attrs: nounwind
define hidden float @main.maxu32f(ptr %context) unnamed_addr #2 {
entry:
  ret float 0x41F0000000000000
}

; Function Attrs: nounwind
define hidden i32 @main.maxu32tof32(ptr %context) unnamed_addr #2 {
entry:
  ret i32 -1
}

; Function Attrs: nounwind
define hidden { i32, i32, i32, i32 } @main.inftoi32(ptr %context) unnamed_addr #2 {
entry:
  ret { i32, i32, i32, i32 } { i32 -1, i32 0, i32 2147483647, i32 -2147483648 }
}

; Function Attrs: nounwind
define hidden i32 @main.u32tof32tou32(i32 %v, ptr %context) unnamed_addr #2 {
entry:
  %0 = uitofp i32 %v to float
  %withinmax = fcmp ole float %0, 0x41EFFFFFC0000000
  %normal = fptoui float %0 to i32
  %1 = select i1 %withinmax, i32 %normal, i32 -1
  ret i32 %1
}

; Function Attrs: nounwind
define hidden float @main.f32tou32tof32(float %v, ptr %context) unnamed_addr #2 {
entry:
  %positive = fcmp oge float %v, 0.000000e+00
  %withinmax = fcmp ole float %v, 0x41EFFFFFC0000000
  %inbounds = and i1 %positive, %withinmax
  %saturated = sext i1 %positive to i32
  %normal = fptoui float %v to i32
  %0 = select i1 %inbounds, i32 %normal, i32 %saturated
  %1 = uitofp i32 %0 to float
  ret float %1
}

; Function Attrs: nounwind
define hidden i8 @main.f32tou8(float %v, ptr %context) unnamed_addr #2 {
entry:
  %positive = fcmp oge float %v, 0.000000e+00
  %withinmax = fcmp ole float %v, 2.550000e+02
  %inbounds = and i1 %positive, %withinmax
  %saturated = sext i1 %positive to i8
  %normal = fptoui float %v to i8
  %0 = select i1 %inbounds, i8 %normal, i8 %saturated
  ret i8 %0
}

; Function Attrs: nounwind
define hidden i8 @main.f32toi8(float %v, ptr %context) unnamed_addr #2 {
entry:
  %abovemin = fcmp oge float %v, -1.280000e+02
  %belowmax = fcmp ole float %v, 1.270000e+02
  %inbounds = and i1 %abovemin, %belowmax
  %saturated = select i1 %abovemin, i8 127, i8 -128
  %isnan = fcmp uno float %v, 0.000000e+00
  %remapped = select i1 %isnan, i8 0, i8 %saturated
  %normal = fptosi float %v to i8
  %0 = select i1 %inbounds, i8 %normal, i8 %remapped
  ret i8 %0
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
