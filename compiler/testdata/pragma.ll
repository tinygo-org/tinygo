; ModuleID = 'pragma.go'
source_filename = "pragma.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

@extern_global = external global [0 x i8], align 1
@main.alignedGlobal = hidden global [4 x i32] zeroinitializer, align 32
@main.alignedGlobal16 = hidden global [4 x i32] zeroinitializer, align 16
@main.globalInSection = hidden global i32 0, section ".special_global_section", align 4
@undefinedGlobalNotInSection = external global i32, align 4
@main.multipleGlobalPragmas = hidden global i32 0, section ".global_section", align 1024

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*) #0

declare void @runtime.trackPointer(i8* nocapture readonly, i8*) #0

; Function Attrs: nounwind
define hidden void @main.init(i8* %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define void @extern_func() #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @somepkg.someFunction1(i8* %context) unnamed_addr #1 {
entry:
  ret void
}

declare void @somepkg.someFunction2(i8*) #0

; Function Attrs: inlinehint nounwind
define hidden void @main.inlineFunc(i8* %context) unnamed_addr #3 {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define hidden void @main.noinlineFunc(i8* %context) unnamed_addr #4 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.functionInSection(i8* %context) unnamed_addr #1 section ".special_function_section" {
entry:
  ret void
}

; Function Attrs: nounwind
define void @exportedFunctionInSection() #5 section ".special_function_section" {
entry:
  ret void
}

declare void @main.undefinedFunctionNotInSection(i8*) #0

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-export-name"="extern_func" }
attributes #3 = { inlinehint nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #4 = { noinline nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #5 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-export-name"="exportedFunctionInSection" }
