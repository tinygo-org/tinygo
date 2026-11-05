; ModuleID = 'lenrange-wasmexport.go'
source_filename = "lenrange-wasmexport.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

@llvm.used = appending global [1 x ptr] [ptr @"main.exportString#wasmexport"]

declare void @runtime.trackPointer(ptr readonly captures(none), ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i1 @main.exportString(ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp slt i32 %s.len, 0
  ret i1 %0
}

; Function Attrs: nounwind
define i1 @"main.exportString#wasmexport"(ptr %0, i32 %1) #2 {
entry:
  call void @runtime.wasmExportCheckRun(ptr undef) #4
  %status = alloca { ptr, i32, i1, i1 }, align 8
  %done.gep = getelementptr inbounds nuw i8, ptr %status, i32 8
  store i1 false, ptr %done.gep, align 1
  store ptr %0, ptr %status, align 4
  %2 = getelementptr inbounds nuw i8, ptr %status, i32 4
  store i32 %1, ptr %2, align 4
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.exportString$gowrapper-wasmexport" to i32), ptr nonnull %status, i32 65536, ptr undef) #4
  call void @runtime.wasmExportRun(ptr nonnull %done.gep, ptr undef) #4
  %3 = getelementptr inbounds nuw i8, ptr %status, i32 9
  %retval = load i1, ptr %3, align 1
  ret i1 %retval
}

declare void @runtime.wasmExportCheckRun(ptr) #0

declare void @runtime.exitGoroutine(ptr) #0

; Function Attrs: nounwind
define linkonce_odr void @"main.exportString$gowrapper-wasmexport"(ptr %0) unnamed_addr #3 {
entry:
  %1 = load ptr, ptr %0, align 4
  %2 = getelementptr inbounds nuw i8, ptr %0, i32 4
  %3 = load i32, ptr %2, align 4
  %4 = call i1 @main.exportString(ptr %1, i32 %3, ptr null)
  %result.ptr = getelementptr inbounds nuw i8, ptr %0, i32 9
  store i1 %4, ptr %result.ptr, align 1
  %done.gep = getelementptr inbounds nuw i8, ptr %0, i32 8
  store i1 true, ptr %done.gep, align 1
  call void @runtime.wasmExportExit(ptr undef) #4
  unreachable
}

declare void @runtime.wasmExportExit(ptr) #0

declare void @"internal/task.start"(i32, ptr, i32, ptr) #0

declare void @runtime.wasmExportRun(ptr dereferenceable_or_null(1), ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.callExportString(ptr readonly %s.data, i32 range(i32 0, -2147483648) %s.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @"github.com/tinygo-org/tinygo/compiler/testdata/lenrangedeps.WasmExportString"(ptr %s.data, i32 %s.len, ptr undef) #4
  ret i1 %0
}

declare i1 @"github.com/tinygo-org/tinygo/compiler/testdata/lenrangedeps.WasmExportString"(ptr readonly, i32, ptr) #0

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-export-name"="exportString" }
attributes #3 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-gowrapper"="main.exportString" }
attributes #4 = { nounwind }
