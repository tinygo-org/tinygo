; ModuleID = 'pragma.go'
source_filename = "pragma.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32--wasi"

@extern_global = external global [0 x i8], align 1
@main.alignedGlobal = hidden global [4 x i32] zeroinitializer, align 32
@main.alignedGlobal16 = hidden global [4 x i32] zeroinitializer, align 16

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define void @extern_func() #0 {
entry:
  ret void
}

define hidden void @somepkg.someFunction1(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

declare void @somepkg.someFunction2(i8*, i8*)

; Function Attrs: inlinehint
define hidden void @main.inlineFunc(i8* %context, i8* %parentHandle) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: noinline
define hidden void @main.noinlineFunc(i8* %context, i8* %parentHandle) unnamed_addr #2 {
entry:
  ret void
}

attributes #0 = { "wasm-export-name"="extern_func" }
attributes #1 = { inlinehint }
attributes #2 = { noinline }
