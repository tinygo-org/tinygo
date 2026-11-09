; ModuleID = 'zeromap.go'
source_filename = "zeromap.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%main.hasPadding = type { i1, i32, i1 }

@"runtime/gc.layout:44-545555550500" = linkonce_odr unnamed_addr constant { i32, [6 x i8] } { i32 44, [6 x i8] c"TUUU\05\00" }
@"runtime.hashmapType:struct{string; string}:int" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 329 to ptr), ptr inttoptr (i32 3 to ptr), ptr @"runtime/gc.layout:44-545555550500" }
@"runtime.hashmapType:[2]string:int" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 69 to ptr), ptr inttoptr (i32 3 to ptr), ptr @"runtime/gc.layout:44-545555550500" }
@"runtime/gc.layout:92-545555555555555555550500" = linkonce_odr unnamed_addr constant { i32, [12 x i8] } { i32 92, [12 x i8] c"TUUUUUUUUU\05\00" }
@"runtime.hashmapType:[5]string:int" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 69 to ptr), ptr inttoptr (i32 3 to ptr), ptr @"runtime/gc.layout:92-545555555555555555550500" }

declare void @runtime.trackPointer(ptr readonly captures(none), ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: noinline nounwind
define hidden i32 @main.testZeroGet(ptr dereferenceable_or_null(52) %m, i1 %s.b1, i32 %s.i, i1 %s.b2, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca %main.hasPadding, align 8
  %hashmap.value = alloca i32, align 4
  %0 = insertvalue %main.hasPadding zeroinitializer, i1 %s.b1, 0
  %1 = insertvalue %main.hasPadding %0, i32 %s.i, 1
  %2 = insertvalue %main.hasPadding %1, i1 %s.b2, 2
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.value)
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.key)
  store %main.hasPadding %2, ptr %hashmap.key, align 4
  %3 = call i1 @runtime.hashmapGenericGet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, i32 4, ptr undef) #4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.key)
  %4 = load i32, ptr %hashmap.value, align 4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.value)
  ret i32 %4
}

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.start.p0(ptr captures(none)) #3

declare i1 @runtime.hashmapGenericGet(ptr dereferenceable_or_null(52), ptr captures(none), ptr captures(none), i32, ptr) #0

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(ptr captures(none)) #3

; Function Attrs: noinline nounwind
define hidden void @main.testZeroSet(ptr dereferenceable_or_null(52) %m, i1 %s.b1, i32 %s.i, i1 %s.b2, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca %main.hasPadding, align 8
  %hashmap.value = alloca i32, align 4
  %0 = insertvalue %main.hasPadding zeroinitializer, i1 %s.b1, 0
  %1 = insertvalue %main.hasPadding %0, i32 %s.i, 1
  %2 = insertvalue %main.hasPadding %1, i1 %s.b2, 2
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.value)
  store i32 5, ptr %hashmap.value, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.key)
  store %main.hasPadding %2, ptr %hashmap.key, align 4
  call void @runtime.hashmapGenericSet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, ptr undef) #4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.key)
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.value)
  ret void
}

declare void @runtime.hashmapGenericSet(ptr dereferenceable_or_null(52), ptr captures(none), ptr captures(none), ptr) #0

; Function Attrs: noinline nounwind
define hidden i32 @main.testZeroArrayGet(ptr dereferenceable_or_null(52) %m, [2 x %main.hasPadding] %s, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca [2 x %main.hasPadding], align 8
  %hashmap.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.value)
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.key)
  %s.elt = extractvalue [2 x %main.hasPadding] %s, 0
  store %main.hasPadding %s.elt, ptr %hashmap.key, align 4
  %hashmap.key.repack1 = getelementptr inbounds nuw i8, ptr %hashmap.key, i32 12
  %s.elt2 = extractvalue [2 x %main.hasPadding] %s, 1
  store %main.hasPadding %s.elt2, ptr %hashmap.key.repack1, align 4
  %0 = call i1 @runtime.hashmapGenericGet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, i32 4, ptr undef) #4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.key)
  %1 = load i32, ptr %hashmap.value, align 4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.value)
  ret i32 %1
}

; Function Attrs: noinline nounwind
define hidden void @main.testZeroArraySet(ptr dereferenceable_or_null(52) %m, [2 x %main.hasPadding] %s, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca [2 x %main.hasPadding], align 8
  %hashmap.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.value)
  store i32 5, ptr %hashmap.value, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.key)
  %s.elt = extractvalue [2 x %main.hasPadding] %s, 0
  store %main.hasPadding %s.elt, ptr %hashmap.key, align 4
  %hashmap.key.repack1 = getelementptr inbounds nuw i8, ptr %hashmap.key, i32 12
  %s.elt2 = extractvalue [2 x %main.hasPadding] %s, 1
  store %main.hasPadding %s.elt2, ptr %hashmap.key.repack1, align 4
  call void @runtime.hashmapGenericSet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, ptr undef) #4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.key)
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.value)
  ret void
}

; Function Attrs: noinline nounwind
define hidden ptr @main.makeStringStructMap(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.hashmapMakeGeneric(i32 16, i32 4, i32 8, ptr nonnull @"runtime.hashmapType:struct{string; string}:int", ptr null, ptr nonnull @_tinygo_array_hash_str, ptr null, ptr nonnull @_tinygo_array_cmp_str, ptr undef) #4
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef) #4
  ret ptr %0
}

; Function Attrs: nounwind
define linkonce_odr i32 @_tinygo_array_hash_str(ptr readonly captures(none) %key, i32 %size, i32 %seed, ptr %context) unnamed_addr #1 {
entry:
  br label %loop

loop:                                             ; preds = %loop, %entry
  %offset.prev = phi i32 [ %size, %entry ], [ %offset, %loop ]
  %hash.prev = phi i32 [ 0, %entry ], [ %hash, %loop ]
  %offset = add nsw i32 %offset.prev, -8
  %elem = getelementptr inbounds i8, ptr %key, i32 %offset
  %hash.elem = call i32 @runtime.hashmapStringPtrHash(ptr dereferenceable(8) %elem, i32 8, i32 %seed, ptr undef) #4
  %hash.prev.mul = mul i32 %hash.prev, 31
  %hash = xor i32 %hash.prev.mul, %hash.elem
  %done = icmp eq i32 %offset, 0
  br i1 %done, label %ret, label %loop

ret:                                              ; preds = %loop
  ret i32 %hash
}

declare i32 @runtime.hashmapStringPtrHash(ptr, i32, i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr i1 @_tinygo_array_cmp_str(ptr readonly captures(none) %x, ptr readonly captures(none) %y, i32 %size, ptr %context) unnamed_addr #1 {
entry:
  br label %loop

loop:                                             ; preds = %next, %entry
  %offset.prev = phi i32 [ %size, %entry ], [ %offset, %next ]
  %offset = add nsw i32 %offset.prev, -8
  %x.elem = getelementptr inbounds i8, ptr %x, i32 %offset
  %y.elem = getelementptr inbounds i8, ptr %y, i32 %offset
  %eq.elem = call i1 @runtime.memCompareString(ptr dereferenceable(8) %x.elem, ptr dereferenceable(8) %y.elem, i32 8, ptr undef) #4
  br i1 %eq.elem, label %next, label %ret

next:                                             ; preds = %loop
  %done = icmp eq i32 %offset, 0
  br i1 %done, label %ret, label %loop

ret:                                              ; preds = %next, %loop
  ret i1 %eq.elem
}

declare i1 @runtime.memCompareString(ptr dereferenceable_or_null(8), ptr dereferenceable_or_null(8), i32, ptr) #0

declare ptr @runtime.hashmapMakeGeneric(i32, i32, i32, ptr, ptr, ptr, ptr, ptr, ptr) #0

; Function Attrs: noinline nounwind
define hidden ptr @main.makeShortStringArrayMap(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.hashmapMakeGeneric(i32 16, i32 4, i32 8, ptr nonnull @"runtime.hashmapType:[2]string:int", ptr null, ptr nonnull @_tinygo_array_hash_str, ptr null, ptr nonnull @_tinygo_array_cmp_str, ptr undef) #4
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef) #4
  ret ptr %0
}

; Function Attrs: noinline nounwind
define hidden ptr @main.makeLongStringArrayMap(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.hashmapMakeGeneric(i32 40, i32 4, i32 8, ptr nonnull @"runtime.hashmapType:[5]string:int", ptr null, ptr nonnull @_tinygo_array_hash_str, ptr null, ptr nonnull @_tinygo_array_cmp_str, ptr undef) #4
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef) #4
  ret ptr %0
}

; Function Attrs: noinline nounwind
define hidden void @main.testBinarySet(ptr dereferenceable_or_null(52) %m, i32 %key, i32 %value, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca i32, align 4
  %hashmap.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.value)
  store i32 %value, ptr %hashmap.value, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.key)
  store i32 %key, ptr %hashmap.key, align 4
  call void @runtime.hashmapBinarySet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, ptr undef) #4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.key)
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.value)
  ret void
}

declare void @runtime.hashmapBinarySet(ptr dereferenceable_or_null(52), ptr captures(none), ptr captures(none), ptr) #0

; Function Attrs: noinline nounwind
define hidden i32 @main.testBinaryGet(ptr dereferenceable_or_null(52) %m, i32 %key, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca i32, align 4
  %hashmap.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.value)
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.key)
  store i32 %key, ptr %hashmap.key, align 4
  %0 = call i1 @runtime.hashmapBinaryGet(ptr %m, ptr nonnull %hashmap.key, ptr nonnull %hashmap.value, i32 4, ptr undef) #4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.key)
  %1 = load i32, ptr %hashmap.value, align 4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.value)
  ret i32 %1
}

declare i1 @runtime.hashmapBinaryGet(ptr dereferenceable_or_null(52), ptr captures(none), ptr captures(none), i32, ptr) #0

; Function Attrs: noinline nounwind
define hidden void @main.testBinaryDelete(ptr dereferenceable_or_null(52) %m, i32 %key, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.key = alloca i32, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.key)
  store i32 %key, ptr %hashmap.key, align 4
  call void @runtime.hashmapBinaryDelete(ptr %m, ptr nonnull %hashmap.key, ptr undef) #4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.key)
  ret void
}

declare void @runtime.hashmapBinaryDelete(ptr dereferenceable_or_null(52), ptr captures(none), ptr) #0

; Function Attrs: noinline nounwind
define hidden void @main.testStringSet(ptr dereferenceable_or_null(52) %m, ptr readonly %key.data, i32 %key.len, i32 %value, ptr %context) unnamed_addr #2 {
entry:
  %hashmap.value = alloca i32, align 4
  call void @llvm.lifetime.start.p0(ptr nonnull %hashmap.value)
  store i32 %value, ptr %hashmap.value, align 4
  call void @runtime.hashmapStringSet(ptr %m, ptr %key.data, i32 %key.len, ptr nonnull %hashmap.value, ptr undef) #4
  call void @llvm.lifetime.end.p0(ptr nonnull %hashmap.value)
  ret void
}

declare void @runtime.hashmapStringSet(ptr dereferenceable_or_null(52), ptr readonly, i32, ptr captures(none), ptr) #0

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
