; ModuleID = 'pragma.go'
source_filename = "pragma.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

@extern_global = external global [0 x i8], align 1
@main.alignedGlobal = hidden global [4 x i32] zeroinitializer, align 32
@main.alignedGlobal16 = hidden global [4 x i32] zeroinitializer, align 16
@llvm.used = appending global [3 x ptr] [ptr @extern_func, ptr @exportedFunctionInSection, ptr @exported]
@main.globalInSection = hidden global i32 0, section ".special_global_section", align 4
@undefinedGlobalNotInSection = external global i32, align 4
@main.multipleGlobalPragmas = hidden global i32 0, section ".global_section", align 1024

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

; Function Attrs: nounwind
define hidden void @main.useGeneric(ptr %context) unnamed_addr #1 {
entry:
  call void @"main.noinlineGenericFunc[int8]"(ptr undef)
  ret void
}

; Function Attrs: noinline nounwind
define linkonce_odr hidden void @"main.noinlineGenericFunc[int8]"(ptr %context) unnamed_addr #4 {
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

declare void @main.declaredImport() #6

declare void @imported() #7

; Function Attrs: nounwind
define void @exported() #8 {
entry:
  ret void
}

declare void @main.undefinedFunctionNotInSection(ptr) #0

declare void @main.doesNotEscapeParam(ptr nocapture dereferenceable_or_null(4), ptr nocapture, i32, i32, ptr nocapture dereferenceable_or_null(36), ptr nocapture, ptr) #0

; Function Attrs: nounwind
define hidden void @main.stillEscapes(ptr dereferenceable_or_null(4) %a, ptr %b.data, i32 %b.len, i32 %b.cap, ptr dereferenceable_or_null(36) %c, ptr %d, ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden ptr @main.doesHeapAlloc(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %new = call align 4 dereferenceable(4) ptr @runtime.alloc_noheap(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #10
  call void @runtime.trackPointer(ptr nonnull %new, ptr nonnull %stackalloc, ptr undef) #10
  ret ptr %new
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc_noheap(i32, ptr, ptr) #9

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-export-name"="extern_func" }
attributes #3 = { inlinehint nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #4 = { noinline nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #5 = { noinline nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-export-name"="exportedFunctionInSection" }
attributes #6 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-import-module"="modulename" "wasm-import-name"="import1" }
attributes #7 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-import-module"="foobar" "wasm-import-name"="imported" }
attributes #8 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-export-name"="exported" }
attributes #9 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #10 = { nounwind }
