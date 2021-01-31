; ModuleID = 'float.go'
source_filename = "float.go"
target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define hidden i32 @main.f32tou32(float %v, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %positive = fcmp oge float %v, 0.000000e+00
  %withinmax = fcmp ole float %v, 0x41EFFFFFC0000000
  %inbounds = and i1 %positive, %withinmax
  %saturated = sext i1 %positive to i32
  %normal = fptoui float %v to i32
  %0 = select i1 %inbounds, i32 %normal, i32 %saturated
  ret i32 %0
}

define hidden float @main.maxu32f(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret float 0x41F0000000000000
}

define hidden i32 @main.maxu32tof32(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret i32 -1
}

define hidden { i32, i32, i32, i32 } @main.inftoi32(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret { i32, i32, i32, i32 } { i32 -1, i32 0, i32 2147483647, i32 -2147483648 }
}

define hidden i32 @main.u32tof32tou32(i32 %v, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = uitofp i32 %v to float
  %withinmax = fcmp ole float %0, 0x41EFFFFFC0000000
  %normal = fptoui float %0 to i32
  %1 = select i1 %withinmax, i32 %normal, i32 -1
  ret i32 %1
}

define hidden float @main.f32tou32tof32(float %v, i8* %context, i8* %parentHandle) unnamed_addr {
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

define hidden i8 @main.f32tou8(float %v, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %positive = fcmp oge float %v, 0.000000e+00
  %withinmax = fcmp ole float %v, 2.550000e+02
  %inbounds = and i1 %positive, %withinmax
  %saturated = sext i1 %positive to i8
  %normal = fptoui float %v to i8
  %0 = select i1 %inbounds, i8 %normal, i8 %saturated
  ret i8 %0
}

define hidden i8 @main.f32toi8(float %v, i8* %context, i8* %parentHandle) unnamed_addr {
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
