; ModuleID = 'zeromap.go'
source_filename = "zeromap.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%main.hasPadding = type { i1, i32, i1 }

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define hidden i32 @main.testZeroGet(ptr dereferenceable_or_null(48) %m, i1 %s.b1, i32 %s.i, i1 %s.b2, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca %main.hasPadding, align 8
  %hashmap.value = alloca i32, align 4
  %0 = insertvalue %main.hasPadding zeroinitializer, i1 %s.b1, 0
  %1 = insertvalue %main.hasPadding %0, i32 %s.i, 1
  %2 = insertvalue %main.hasPadding %1, i1 %s.b2, 2
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %hashmap.value)
  call void @llvm.lifetime.start.p0(i64 12, ptr nonnull %hashmap.key)
  store %main.hasPadding %2, ptr %hashmap.key, align 4
  %3 = call i1 @runtime.hashmapGenericGet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, i32 4, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 12, ptr nonnull %hashmap.key)
  %4 = load i32, ptr %hashmap.value, align 4
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %hashmap.value)
  ret i32 %4
}

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture) #3

declare i1 @runtime.hashmapGenericGet(ptr dereferenceable_or_null(48), ptr nocapture, ptr nocapture, i32, ptr) #0

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(i64 immarg, ptr nocapture) #3

; Function Attrs: noinline nounwind
define hidden void @main.testZeroSet(ptr dereferenceable_or_null(48) %m, i1 %s.b1, i32 %s.i, i1 %s.b2, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca %main.hasPadding, align 8
  %hashmap.value = alloca i32, align 4
  %0 = insertvalue %main.hasPadding zeroinitializer, i1 %s.b1, 0
  %1 = insertvalue %main.hasPadding %0, i32 %s.i, 1
  %2 = insertvalue %main.hasPadding %1, i1 %s.b2, 2
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %hashmap.value)
  store i32 5, ptr %hashmap.value, align 4
  call void @llvm.lifetime.start.p0(i64 12, ptr nonnull %hashmap.key)
  store %main.hasPadding %2, ptr %hashmap.key, align 4
  call void @runtime.hashmapGenericSet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 12, ptr nonnull %hashmap.key)
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %hashmap.value)
  ret void
}

declare void @runtime.hashmapGenericSet(ptr dereferenceable_or_null(48), ptr nocapture, ptr nocapture, ptr) #0

; Function Attrs: noinline nounwind
define hidden i32 @main.testZeroArrayGet(ptr dereferenceable_or_null(48) %m, [2 x %main.hasPadding] %s, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca [2 x %main.hasPadding], align 8
  %hashmap.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %hashmap.value)
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %hashmap.key)
  %s.elt = extractvalue [2 x %main.hasPadding] %s, 0
  store %main.hasPadding %s.elt, ptr %hashmap.key, align 4
  %hashmap.key.repack1 = getelementptr inbounds nuw i8, ptr %hashmap.key, i32 12
  %s.elt2 = extractvalue [2 x %main.hasPadding] %s, 1
  store %main.hasPadding %s.elt2, ptr %hashmap.key.repack1, align 4
  %0 = call i1 @runtime.hashmapGenericGet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, i32 4, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %hashmap.key)
  %1 = load i32, ptr %hashmap.value, align 4
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %hashmap.value)
  ret i32 %1
}

; Function Attrs: noinline nounwind
define hidden void @main.testZeroArraySet(ptr dereferenceable_or_null(48) %m, [2 x %main.hasPadding] %s, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca [2 x %main.hasPadding], align 8
  %hashmap.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %hashmap.value)
  store i32 5, ptr %hashmap.value, align 4
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %hashmap.key)
  %s.elt = extractvalue [2 x %main.hasPadding] %s, 0
  store %main.hasPadding %s.elt, ptr %hashmap.key, align 4
  %hashmap.key.repack1 = getelementptr inbounds nuw i8, ptr %hashmap.key, i32 12
  %s.elt2 = extractvalue [2 x %main.hasPadding] %s, 1
  store %main.hasPadding %s.elt2, ptr %hashmap.key.repack1, align 4
  call void @runtime.hashmapGenericSet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, ptr undef) #4
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %hashmap.key)
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %hashmap.value)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.main(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { noinline nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nocallback nofree nosync nounwind willreturn memory(argmem: readwrite) }
attributes #4 = { nounwind }
