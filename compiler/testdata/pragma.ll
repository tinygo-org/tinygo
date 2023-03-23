; ModuleID = 'pragma.go'
source_filename = "pragma.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

@extern_global = external global [0 x i8], align 1
@main.alignedGlobal = hidden global [4 x i32] zeroinitializer, align 32
@main.alignedGlobal16 = hidden global [4 x i32] zeroinitializer, align 16
@llvm.used = appending global [3 x ptr] [ptr @extern_func, ptr @exportedFunctionInSection, ptr @exported]
@main.globalInSection = hidden global i32 0, section ".special_global_section", align 4
@undefinedGlobalNotInSection = external global i32, align 4
@main.multipleGlobalPragmas = hidden global i32 0, section ".global_section", align 1024

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define void @extern_func() #3 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @somepkg.someFunction1(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

declare void @somepkg.someFunction2(ptr) #1

; Function Attrs: inlinehint nounwind
define hidden void @main.inlineFunc(ptr %context) unnamed_addr #4 {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define hidden void @main.noinlineFunc(ptr %context) unnamed_addr #5 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.useGeneric(ptr %context) unnamed_addr #2 {
entry:
  call void @"main.noinlineGenericFunc[int8]"(ptr undef)
  ret void
}

; Function Attrs: noinline nounwind
define linkonce_odr hidden void @"main.noinlineGenericFunc[int8]"(ptr %context) unnamed_addr #5 {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define hidden void @main.functionInSection(ptr %context) unnamed_addr #5 section ".special_function_section" {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define void @exportedFunctionInSection() #6 section ".special_function_section" {
entry:
  ret void
}

declare void @main.declaredImport() #7

declare void @imported() #8

; Function Attrs: nounwind
define void @exported() #9 {
entry:
  ret void
}

declare void @main.undefinedFunctionNotInSection(ptr) #1

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #3 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" "wasm-export-name"="extern_func" }
attributes #4 = { inlinehint nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #5 = { noinline nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" }
attributes #6 = { noinline nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" "wasm-export-name"="exportedFunctionInSection" }
attributes #7 = { "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" "wasm-import-module"="modulename" "wasm-import-name"="import1" }
attributes #8 = { "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" "wasm-import-module"="foobar" "wasm-import-name"="imported" }
attributes #9 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext" "wasm-export-name"="exported" }
