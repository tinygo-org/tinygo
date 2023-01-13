; ModuleID = 'pragma.go'
source_filename = "pragma.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

@extern_global = external global [0 x i8], align 1
@main.alignedGlobal = hidden global [4 x i32] zeroinitializer, align 32
@main.alignedGlobal16 = hidden global [4 x i32] zeroinitializer, align 16
@llvm.used = appending global [2 x ptr] [ptr @extern_func, ptr @exportedFunctionInSection]
@main.globalInSection = hidden global i32 0, section ".special_global_section", align 4
@undefinedGlobalNotInSection = external global i32, align 4
@main.multipleGlobalPragmas = hidden global i32 0, section ".global_section", align 1024

declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define void @extern_func() #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @somepkg.someFunction1(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

declare void @somepkg.someFunction2(ptr) #0

; Function Attrs: inlinehint nounwind
define hidden void @main.inlineFunc(ptr %context) unnamed_addr #3 {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define hidden void @main.noinlineFunc(ptr %context) unnamed_addr #4 {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define hidden void @main.functionInSection(ptr %context) unnamed_addr #4 section ".special_function_section" {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define void @exportedFunctionInSection() #5 section ".special_function_section" {
entry:
  ret void
}

declare void @main.undefinedFunctionNotInSection(ptr) #0

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-export-name"="extern_func" "wasm-import-module"="env" "wasm-import-name"="extern_func" }
attributes #3 = { inlinehint nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #4 = { noinline nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #5 = { noinline nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "wasm-export-name"="exportedFunctionInSection" "wasm-import-module"="env" "wasm-import-name"="exportedFunctionInSection" }
