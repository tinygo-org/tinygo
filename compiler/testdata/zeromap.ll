; ModuleID = 'zeromap.go'
source_filename = "zeromap.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%main.hasPadding = type { i1, i32, i1 }

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define hidden i32 @main.testZeroGet(ptr dereferenceable_or_null(40) %m, i1 %s.b1, i32 %s.i, i1 %s.b2, ptr %context) unnamed_addr #3 {
entry:
  %hashmap.key = alloca %main.hasPadding, align 8
  %hashmap.value = alloca i32, align 4
  %0 = insertvalue %main.hasPadding zeroinitializer, i1 %s.b1, 0
  %1 = insertvalue %main.hasPadding %0, i32 %s.i, 1
  %2 = insertvalue %main.hasPadding %1, i1 %s.b2, 2
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %hashmap.value)
  call void @llvm.lifetime.start.p0(i64 12, ptr nonnull %hashmap.key)
  store %main.hasPadding %2, ptr %hashmap.key, align 8
  %3 = getelementptr inbounds i8, ptr %hashmap.key, i32 1
  call void @runtime.memzero(ptr nonnull %3, i32 3, ptr undef) #5
  %4 = getelementptr inbounds i8, ptr %hashmap.key, i32 9
  call void @runtime.memzero(ptr nonnull %4, i32 3, ptr undef) #5
  %5 = call i1 @runtime.hashmapBinaryGet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, i32 4, ptr undef) #5
  call void @llvm.lifetime.end.p0(i64 12, ptr nonnull %hashmap.key)
  %6 = load i32, ptr %hashmap.value, align 4
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %hashmap.value)
  ret i32 %6
}

; Function Attrs: argmemonly nocallback nofree nosync nounwind willreturn
declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture) #4

declare void @runtime.memzero(ptr, i32, ptr) #1

declare i1 @runtime.hashmapBinaryGet(ptr dereferenceable_or_null(40), ptr, ptr, i32, ptr) #1

; Function Attrs: argmemonly nocallback nofree nosync nounwind willreturn
declare void @llvm.lifetime.end.p0(i64 immarg, ptr nocapture) #4

; Function Attrs: noinline nounwind
define hidden void @main.testZeroSet(ptr dereferenceable_or_null(40) %m, i1 %s.b1, i32 %s.i, i1 %s.b2, ptr %context) unnamed_addr #3 {
entry:
  %hashmap.key = alloca %main.hasPadding, align 8
  %hashmap.value = alloca i32, align 4
  %0 = insertvalue %main.hasPadding zeroinitializer, i1 %s.b1, 0
  %1 = insertvalue %main.hasPadding %0, i32 %s.i, 1
  %2 = insertvalue %main.hasPadding %1, i1 %s.b2, 2
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %hashmap.value)
  store i32 5, ptr %hashmap.value, align 4
  call void @llvm.lifetime.start.p0(i64 12, ptr nonnull %hashmap.key)
  store %main.hasPadding %2, ptr %hashmap.key, align 8
  %3 = getelementptr inbounds i8, ptr %hashmap.key, i32 1
  call void @runtime.memzero(ptr nonnull %3, i32 3, ptr undef) #5
  %4 = getelementptr inbounds i8, ptr %hashmap.key, i32 9
  call void @runtime.memzero(ptr nonnull %4, i32 3, ptr undef) #5
  call void @runtime.hashmapBinarySet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, ptr undef) #5
  call void @llvm.lifetime.end.p0(i64 12, ptr nonnull %hashmap.key)
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %hashmap.value)
  ret void
}

declare void @runtime.hashmapBinarySet(ptr dereferenceable_or_null(40), ptr, ptr, ptr) #1

; Function Attrs: noinline nounwind
define hidden i32 @main.testZeroArrayGet(ptr dereferenceable_or_null(40) %m, [2 x %main.hasPadding] %s, ptr %context) unnamed_addr #3 {
entry:
  %hashmap.key = alloca [2 x %main.hasPadding], align 8
  %hashmap.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %hashmap.value)
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %hashmap.key)
  %s.elt = extractvalue [2 x %main.hasPadding] %s, 0
  store %main.hasPadding %s.elt, ptr %hashmap.key, align 8
  %hashmap.key.repack1 = getelementptr inbounds [2 x %main.hasPadding], ptr %hashmap.key, i32 0, i32 1
  %s.elt2 = extractvalue [2 x %main.hasPadding] %s, 1
  store %main.hasPadding %s.elt2, ptr %hashmap.key.repack1, align 4
  %0 = getelementptr inbounds i8, ptr %hashmap.key, i32 1
  call void @runtime.memzero(ptr nonnull %0, i32 3, ptr undef) #5
  %1 = getelementptr inbounds i8, ptr %hashmap.key, i32 9
  call void @runtime.memzero(ptr nonnull %1, i32 3, ptr undef) #5
  %2 = getelementptr inbounds i8, ptr %hashmap.key, i32 13
  call void @runtime.memzero(ptr nonnull %2, i32 3, ptr undef) #5
  %3 = getelementptr inbounds i8, ptr %hashmap.key, i32 21
  call void @runtime.memzero(ptr nonnull %3, i32 3, ptr undef) #5
  %4 = call i1 @runtime.hashmapBinaryGet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, i32 4, ptr undef) #5
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %hashmap.key)
  %5 = load i32, ptr %hashmap.value, align 4
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %hashmap.value)
  ret i32 %5
}

; Function Attrs: noinline nounwind
define hidden void @main.testZeroArraySet(ptr dereferenceable_or_null(40) %m, [2 x %main.hasPadding] %s, ptr %context) unnamed_addr #3 {
entry:
  %hashmap.key = alloca [2 x %main.hasPadding], align 8
  %hashmap.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(i64 4, ptr nonnull %hashmap.value)
  store i32 5, ptr %hashmap.value, align 4
  call void @llvm.lifetime.start.p0(i64 24, ptr nonnull %hashmap.key)
  %s.elt = extractvalue [2 x %main.hasPadding] %s, 0
  store %main.hasPadding %s.elt, ptr %hashmap.key, align 8
  %hashmap.key.repack1 = getelementptr inbounds [2 x %main.hasPadding], ptr %hashmap.key, i32 0, i32 1
  %s.elt2 = extractvalue [2 x %main.hasPadding] %s, 1
  store %main.hasPadding %s.elt2, ptr %hashmap.key.repack1, align 4
  %0 = getelementptr inbounds i8, ptr %hashmap.key, i32 1
  call void @runtime.memzero(ptr nonnull %0, i32 3, ptr undef) #5
  %1 = getelementptr inbounds i8, ptr %hashmap.key, i32 9
  call void @runtime.memzero(ptr nonnull %1, i32 3, ptr undef) #5
  %2 = getelementptr inbounds i8, ptr %hashmap.key, i32 13
  call void @runtime.memzero(ptr nonnull %2, i32 3, ptr undef) #5
  %3 = getelementptr inbounds i8, ptr %hashmap.key, i32 21
  call void @runtime.memzero(ptr nonnull %3, i32 3, ptr undef) #5
  call void @runtime.hashmapBinarySet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, ptr undef) #5
  call void @llvm.lifetime.end.p0(i64 24, ptr nonnull %hashmap.key)
  call void @llvm.lifetime.end.p0(i64 4, ptr nonnull %hashmap.value)
  ret void
}

; Function Attrs: nounwind
define hidden void @main.main(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #3 = { noinline nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #4 = { argmemonly nocallback nofree nosync nounwind willreturn }
attributes #5 = { nounwind }
