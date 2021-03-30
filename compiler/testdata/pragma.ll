; ModuleID = 'pragma.go'
source_filename = "pragma.go"
target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

@extern_global = external global [0 x i8]
@main.alignedGlobal = hidden global [4 x i32] zeroinitializer, align 32
@main.alignedGlobal16 = hidden global [4 x i32] zeroinitializer, align 16

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define void @extern_func() {
entry:
  ret void
}

define hidden void @somepkg.someFunction1(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

declare void @somepkg.someFunction2(i8*, i8*)

; Function Attrs: inlinehint
define hidden void @main.inlineFunc(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: noinline
define hidden void @main.noinlineFunc(i8* %context, i8* %parentHandle) unnamed_addr #1 {
entry:
  ret void
}

attributes #0 = { inlinehint }
attributes #1 = { noinline }
