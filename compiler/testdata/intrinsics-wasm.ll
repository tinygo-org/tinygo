; ModuleID = 'intrinsics.go'
source_filename = "intrinsics.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32--wasi"

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define hidden double @main.mySqrt(double %x, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = call double @llvm.sqrt.f64(double %x)
  ret double %0
}

; Function Attrs: nounwind readnone speculatable willreturn
declare double @llvm.sqrt.f64(double) #0

define hidden double @main.myTrunc(double %x, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = call double @llvm.trunc.f64(double %x)
  ret double %0
}

; Function Attrs: nounwind readnone speculatable willreturn
declare double @llvm.trunc.f64(double) #0

attributes #0 = { nounwind readnone speculatable willreturn }
