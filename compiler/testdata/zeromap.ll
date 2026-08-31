; ModuleID = 'zeromap.go'
source_filename = "zeromap.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%main.hasPadding = type { i1, i32, i1 }
%runtime._string = type { ptr, i32 }

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
declare void @llvm.lifetime.start.p0(ptr nocapture) #3

declare i1 @runtime.hashmapGenericGet(ptr dereferenceable_or_null(48), ptr nocapture, ptr nocapture, i32, ptr) #0

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(ptr nocapture) #3

; Function Attrs: noinline nounwind
define hidden void @main.testZeroSet(ptr dereferenceable_or_null(48) %m, i1 %s.b1, i32 %s.i, i1 %s.b2, ptr %context) unnamed_addr #2 {
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

declare void @runtime.hashmapGenericSet(ptr dereferenceable_or_null(48), ptr nocapture, ptr nocapture, ptr) #0

; Function Attrs: noinline nounwind
define hidden i32 @main.testZeroArrayGet(ptr dereferenceable_or_null(48) %m, [2 x %main.hasPadding] %s, ptr %context) unnamed_addr #2 {
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
define hidden void @main.testZeroArraySet(ptr dereferenceable_or_null(48) %m, [2 x %main.hasPadding] %s, ptr %context) unnamed_addr #2 {
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
  %0 = call ptr @runtime.hashmapMakeGeneric(i32 16, i32 4, i32 8, ptr null, ptr nonnull @"hashmapKeyHash.struct{string; string}", ptr null, ptr nonnull @"hashmapKeyEqual.struct{string; string}", ptr undef) #4
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef) #4
  ret ptr %0
}

; Function Attrs: nounwind
define linkonce_odr i32 @"hashmapKeyHash.struct{string; string}"(ptr %0, i32 %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %hash = call i32 @runtime.hashmapStringPtrHash(ptr %0, i32 8, i32 %2, ptr undef) #4
  %4 = getelementptr inbounds nuw i8, ptr %0, i32 8
  %hash1 = call i32 @runtime.hashmapStringPtrHash(ptr nonnull %4, i32 8, i32 %2, ptr undef) #4
  %5 = mul i32 %hash, 31
  %6 = xor i32 %5, %hash1
  ret i32 %6
}

declare i32 @runtime.hashmapStringPtrHash(ptr, i32, i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr i1 @"hashmapKeyEqual.struct{string; string}"(ptr %0, ptr %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %x.str.unpack = load ptr, ptr %0, align 4
  %x.str.elt4 = getelementptr inbounds nuw i8, ptr %0, i32 4
  %x.str.unpack5 = load i32, ptr %x.str.elt4, align 4
  %y.str.unpack = load ptr, ptr %1, align 4
  %y.str.elt7 = getelementptr inbounds nuw i8, ptr %1, i32 4
  %y.str.unpack8 = load i32, ptr %y.str.elt7, align 4
  %eq = call i1 @runtime.stringEqual(ptr %x.str.unpack, i32 %x.str.unpack5, ptr %y.str.unpack, i32 %y.str.unpack8, ptr undef) #4
  %4 = getelementptr inbounds nuw i8, ptr %0, i32 8
  %5 = getelementptr inbounds nuw i8, ptr %1, i32 8
  %x.str1.unpack = load ptr, ptr %4, align 4
  %x.str1.elt10 = getelementptr inbounds nuw i8, ptr %0, i32 12
  %x.str1.unpack11 = load i32, ptr %x.str1.elt10, align 4
  %y.str2.unpack = load ptr, ptr %5, align 4
  %y.str2.elt13 = getelementptr inbounds nuw i8, ptr %1, i32 12
  %y.str2.unpack14 = load i32, ptr %y.str2.elt13, align 4
  %eq3 = call i1 @runtime.stringEqual(ptr %x.str1.unpack, i32 %x.str1.unpack11, ptr %y.str2.unpack, i32 %y.str2.unpack14, ptr undef) #4
  %6 = and i1 %eq, %eq3
  ret i1 %6
}

declare i1 @runtime.stringEqual(ptr readonly, i32, ptr readonly, i32, ptr) #0

declare ptr @runtime.hashmapMakeGeneric(i32, i32, i32, ptr, ptr, ptr, ptr, ptr) #0

; Function Attrs: noinline nounwind
define hidden ptr @main.makeShortStringArrayMap(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.hashmapMakeGeneric(i32 16, i32 4, i32 8, ptr null, ptr nonnull @"hashmapKeyHash.[2]string", ptr null, ptr nonnull @"hashmapKeyEqual.[2]string", ptr undef) #4
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef) #4
  ret ptr %0
}

; Function Attrs: nounwind
define linkonce_odr i32 @"hashmapKeyHash.[2]string"(ptr %0, i32 %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %hash = call i32 @runtime.hashmapStringPtrHash(ptr %0, i32 8, i32 %2, ptr undef) #4
  %4 = getelementptr inbounds nuw i8, ptr %0, i32 8
  %hash1 = call i32 @runtime.hashmapStringPtrHash(ptr nonnull %4, i32 8, i32 %2, ptr undef) #4
  %5 = mul i32 %hash, 31
  %6 = xor i32 %5, %hash1
  ret i32 %6
}

; Function Attrs: nounwind
define linkonce_odr i1 @"hashmapKeyEqual.[2]string"(ptr %0, ptr %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  %x.str.unpack = load ptr, ptr %0, align 4
  %x.str.elt4 = getelementptr inbounds nuw i8, ptr %0, i32 4
  %x.str.unpack5 = load i32, ptr %x.str.elt4, align 4
  %y.str.unpack = load ptr, ptr %1, align 4
  %y.str.elt7 = getelementptr inbounds nuw i8, ptr %1, i32 4
  %y.str.unpack8 = load i32, ptr %y.str.elt7, align 4
  %eq = call i1 @runtime.stringEqual(ptr %x.str.unpack, i32 %x.str.unpack5, ptr %y.str.unpack, i32 %y.str.unpack8, ptr undef) #4
  %4 = getelementptr inbounds nuw i8, ptr %0, i32 8
  %5 = getelementptr inbounds nuw i8, ptr %1, i32 8
  %x.str1.unpack = load ptr, ptr %4, align 4
  %x.str1.elt10 = getelementptr inbounds nuw i8, ptr %0, i32 12
  %x.str1.unpack11 = load i32, ptr %x.str1.elt10, align 4
  %y.str2.unpack = load ptr, ptr %5, align 4
  %y.str2.elt13 = getelementptr inbounds nuw i8, ptr %1, i32 12
  %y.str2.unpack14 = load i32, ptr %y.str2.elt13, align 4
  %eq3 = call i1 @runtime.stringEqual(ptr %x.str1.unpack, i32 %x.str1.unpack11, ptr %y.str2.unpack, i32 %y.str2.unpack14, ptr undef) #4
  %6 = and i1 %eq, %eq3
  ret i1 %6
}

; Function Attrs: noinline nounwind
define hidden ptr @main.makeLongStringArrayMap(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.hashmapMakeGeneric(i32 40, i32 4, i32 8, ptr null, ptr nonnull @"hashmapKeyHash.[5]string", ptr null, ptr nonnull @"hashmapKeyEqual.[5]string", ptr undef) #4
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef) #4
  ret ptr %0
}

; Function Attrs: nounwind
define linkonce_odr i32 @"hashmapKeyHash.[5]string"(ptr %0, i32 %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  br label %hash.array.body

hash.array.body:                                  ; preds = %hash.array.body, %entry
  %i = phi i32 [ 0, %entry ], [ %7, %hash.array.body ]
  %hash.acc = phi i32 [ 0, %entry ], [ %6, %hash.array.body ]
  %4 = getelementptr inbounds nuw %runtime._string, ptr %0, i32 %i
  %hash = call i32 @runtime.hashmapStringPtrHash(ptr %4, i32 8, i32 %2, ptr undef) #4
  %5 = mul i32 %hash.acc, 31
  %6 = xor i32 %5, %hash
  %7 = add nuw nsw i32 %i, 1
  %8 = icmp samesign ult i32 %i, 4
  br i1 %8, label %hash.array.body, label %hash.array.done

hash.array.done:                                  ; preds = %hash.array.body
  ret i32 %6
}

; Function Attrs: nounwind
define linkonce_odr i1 @"hashmapKeyEqual.[5]string"(ptr %0, ptr %1, i32 %2, ptr %3) unnamed_addr #1 {
entry:
  br label %eq.array.body

eq.array.body:                                    ; preds = %eq.array.body, %entry
  %i = phi i32 [ 0, %entry ], [ %6, %eq.array.body ]
  %4 = getelementptr inbounds %runtime._string, ptr %0, i32 %i
  %5 = getelementptr inbounds %runtime._string, ptr %1, i32 %i
  %x.str.unpack = load ptr, ptr %4, align 4
  %x.str.elt1 = getelementptr inbounds nuw i8, ptr %4, i32 4
  %x.str.unpack2 = load i32, ptr %x.str.elt1, align 4
  %y.str.unpack = load ptr, ptr %5, align 4
  %y.str.elt4 = getelementptr inbounds nuw i8, ptr %5, i32 4
  %y.str.unpack5 = load i32, ptr %y.str.elt4, align 4
  %eq = call i1 @runtime.stringEqual(ptr %x.str.unpack, i32 %x.str.unpack2, ptr %y.str.unpack, i32 %y.str.unpack5, ptr undef) #4
  %6 = add i32 %i, 1
  %7 = icmp ult i32 %6, 5
  %.not7 = and i1 %7, %eq
  br i1 %.not7, label %eq.array.body, label %eq.array.done

eq.array.done:                                    ; preds = %eq.array.body
  ret i1 %eq
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
