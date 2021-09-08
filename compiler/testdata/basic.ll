; ModuleID = 'basic.go'
source_filename = "basic.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32--wasi"

%main.kv = type { float }
%main.kv.0 = type { i8 }

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define hidden i32 @main.addInt(i32 %x, i32 %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = add i32 %x, %y
  ret i32 %0
}

define hidden i1 @main.equalInt(i32 %x, i32 %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = icmp eq i32 %x, %y
  ret i1 %0
}

define hidden i1 @main.floatEQ(float %x, float %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = fcmp oeq float %x, %y
  ret i1 %0
}

define hidden i1 @main.floatNE(float %x, float %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = fcmp une float %x, %y
  ret i1 %0
}

define hidden i1 @main.floatLower(float %x, float %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = fcmp olt float %x, %y
  ret i1 %0
}

define hidden i1 @main.floatLowerEqual(float %x, float %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = fcmp ole float %x, %y
  ret i1 %0
}

define hidden i1 @main.floatGreater(float %x, float %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = fcmp ogt float %x, %y
  ret i1 %0
}

define hidden i1 @main.floatGreaterEqual(float %x, float %y, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = fcmp oge float %x, %y
  ret i1 %0
}

define hidden float @main.complexReal(float %x.r, float %x.i, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret float %x.r
}

define hidden float @main.complexImag(float %x.r, float %x.i, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret float %x.i
}

define hidden { float, float } @main.complexAdd(float %x.r, float %x.i, float %y.r, float %y.i, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = fadd float %x.r, %y.r
  %1 = fadd float %x.i, %y.i
  %2 = insertvalue { float, float } undef, float %0, 0
  %3 = insertvalue { float, float } %2, float %1, 1
  ret { float, float } %3
}

define hidden { float, float } @main.complexSub(float %x.r, float %x.i, float %y.r, float %y.i, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = fsub float %x.r, %y.r
  %1 = fsub float %x.i, %y.i
  %2 = insertvalue { float, float } undef, float %0, 0
  %3 = insertvalue { float, float } %2, float %1, 1
  ret { float, float } %3
}

define hidden { float, float } @main.complexMul(float %x.r, float %x.i, float %y.r, float %y.i, i8* %context, i8* %parentHandle) unnamed_addr {
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

define hidden void @main.foo(%main.kv* dereferenceable_or_null(4) %a, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  call void @"main.foo$1"(%main.kv.0* null, i8* undef, i8* undef)
  ret void
}

define hidden void @"main.foo$1"(%main.kv.0* dereferenceable_or_null(1) %b, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}
