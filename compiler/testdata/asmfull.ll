; ModuleID = 'asmfull_bug.go'
source_filename = "asmfull_bug.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._interface = type { ptr, ptr }

@"reflect/types.type:basic:uint32" = linkonce_odr constant { i8, ptr } { i8 -54, ptr @"reflect/types.type:pointer:basic:uint32" }, align 4
@"reflect/types.type:pointer:basic:uint32" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:uint32" }, align 4
@"main$string" = internal unnamed_addr constant [5 x i8] c"input", align 1
@"main$string.1" = internal unnamed_addr constant [5 x i8] c"input", align 1

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.testAsmFullBug(ptr %context) unnamed_addr #2 {
entry:
  %hashmap.value1 = alloca %runtime._interface, align 8
  %hashmap.value = alloca %runtime._interface, align 8
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.hashmapMake(i32 8, i32 8, i32 8, i8 1, ptr undef) #4
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:uint32", ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull inttoptr (i32 42 to ptr), ptr nonnull %stackalloc, ptr undef) #4
  call void @llvm.lifetime.start.p0(i64 8, ptr nonnull %hashmap.value)
  store ptr @"reflect/types.type:basic:uint32", ptr %hashmap.value, align 4
  %hashmap.value.repack2 = getelementptr inbounds i8, ptr %hashmap.value, i32 4
  store ptr inttoptr (i32 42 to ptr), ptr %hashmap.value.repack2, align 4
  call void @runtime.hashmapStringSet(ptr %0, ptr nonnull @"main$string", i32 5, ptr nonnull %hashmap.value, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 8, ptr nonnull %hashmap.value)
  %1 = call i32 asm sideeffect "mov $0, ${1}", "=&r,r"(i32 42) #4
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:uint32", ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull inttoptr (i32 44 to ptr), ptr nonnull %stackalloc, ptr undef) #4
  call void @llvm.lifetime.start.p0(i64 8, ptr nonnull %hashmap.value1)
  store ptr @"reflect/types.type:basic:uint32", ptr %hashmap.value1, align 4
  %hashmap.value1.repack3 = getelementptr inbounds i8, ptr %hashmap.value1, i32 4
  store ptr inttoptr (i32 44 to ptr), ptr %hashmap.value1.repack3, align 4
  call void @runtime.hashmapStringSet(ptr %0, ptr nonnull @"main$string.1", i32 5, ptr nonnull %hashmap.value1, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 8, ptr nonnull %hashmap.value1)
  ret i32 %1
}

declare ptr @runtime.hashmapMake(i32, i32, i32, i8, ptr) #1

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture) #3

declare void @runtime.hashmapStringSet(ptr dereferenceable_or_null(40), ptr, i32, ptr, ptr) #1

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(i64 immarg, ptr nocapture) #3

; Function Attrs: nounwind
define hidden void @main.main(ptr %context) unnamed_addr #2 {
entry:
  %0 = call i32 @main.testAsmFullBug(ptr undef)
  ret void
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nocallback nofree nosync nounwind willreturn memory(argmem: readwrite) }
attributes #4 = { nounwind }
