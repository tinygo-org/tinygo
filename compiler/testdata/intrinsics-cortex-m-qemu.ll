; ModuleID = 'intrinsics.go'
source_filename = "intrinsics.go"
target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-unknown-unknown-eabi"

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden double @main.mySqrt(double %x, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = call double @math.Sqrt(double %x, i8* undef, i8* undef) #0
  ret double %0
}

declare double @math.Sqrt(double, i8*, i8*)

; Function Attrs: nounwind
define hidden double @main.myTrunc(double %x, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = call double @math.Trunc(double %x, i8* undef, i8* undef) #0
  ret double %0
}

declare double @math.Trunc(double, i8*, i8*)

attributes #0 = { nounwind }
