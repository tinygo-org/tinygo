; ModuleID = 'memcmp.go'
source_filename = "memcmp.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

@"runtime.hashmapType:struct{uint8; uint8; uint8}:struct{}" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 3 to ptr), ptr inttoptr (i32 3 to ptr), ptr inttoptr (i32 277 to ptr) }
@"runtime.hashmapType:[16][16]struct{uint8; uint8; uint8}:struct{}" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 3 to ptr), ptr inttoptr (i32 3 to ptr), ptr inttoptr (i32 261401 to ptr) }
@"runtime.hashmapType:struct{float32; float32; float32}:struct{}" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 3 to ptr), ptr inttoptr (i32 3 to ptr), ptr inttoptr (i32 313 to ptr) }
@"runtime/gc.layout:76-04000000000000000000" = linkonce_odr unnamed_addr constant { i32, [10 x i8] } { i32 76, [10 x i8] c"\04\00\00\00\00\00\00\00\00\00" }
@"runtime.hashmapType:[3]struct{float32; float32; float32}:struct{}" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 3 to ptr), ptr inttoptr (i32 3 to ptr), ptr @"runtime/gc.layout:76-04000000000000000000" }
@"runtime/gc.layout:180-04d0c0013470000d1c400307d0c0013470000d1c400307" = linkonce_odr unnamed_addr constant { i32, [23 x i8] } { i32 180, [23 x i8] c"\04\D0\C0\014p\00\0D\1C@\03\07\D0\C0\014p\00\0D\1C@\03\07" }
@"runtime.hashmapType:struct{struct{float32; float64}; struct{bool; uint64; string; interface{}; float32; complex128; interface{Error() string}}; string}:struct{}" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 117653549 to ptr), ptr inttoptr (i32 3 to ptr), ptr @"runtime/gc.layout:180-04d0c0013470000d1c400307d0c0013470000d1c400307" }
@"runtime/gc.layout:116-f43ffc0fffc3fff03ffc0fffc3ff00" = linkonce_odr unnamed_addr constant { i32, [15 x i8] } { i32 116, [15 x i8] c"\F4?\FC\0F\FF\C3\FF\F0?\FC\0F\FF\C3\FF\00" }
@"runtime.hashmapType:struct{[5]interface{}; [16]byte}:struct{}" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 65501 to ptr), ptr inttoptr (i32 3 to ptr), ptr @"runtime/gc.layout:116-f43ffc0fffc3fff03ffc0fffc3ff00" }

declare void @runtime.trackPointer(ptr readonly captures(none), ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call ptr @runtime.hashmapMakeGeneric(i32 3, i32 0, i32 0, ptr nonnull @"runtime.hashmapType:struct{uint8; uint8; uint8}:struct{}", ptr null, ptr nonnull @runtime.hash32, ptr null, ptr nonnull @runtime.memCompareBytes, ptr undef) #2
  call void @runtime.trackPointer(ptr %0, ptr nonnull %stackalloc, ptr undef) #2
  %1 = call ptr @runtime.hashmapMakeGeneric(i32 768, i32 0, i32 0, ptr nonnull @"runtime.hashmapType:[16][16]struct{uint8; uint8; uint8}:struct{}", ptr null, ptr nonnull @runtime.hash32, ptr null, ptr nonnull @runtime.memCompareBytes, ptr undef) #2
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #2
  %2 = call ptr @runtime.hashmapMakeGeneric(i32 12, i32 0, i32 0, ptr nonnull @"runtime.hashmapType:struct{float32; float32; float32}:struct{}", ptr null, ptr nonnull @_tinygo_array_hash_f32, ptr null, ptr nonnull @_tinygo_array_cmp_f32, ptr undef) #2
  call void @runtime.trackPointer(ptr %2, ptr nonnull %stackalloc, ptr undef) #2
  %3 = call ptr @runtime.hashmapMakeGeneric(i32 36, i32 0, i32 0, ptr nonnull @"runtime.hashmapType:[3]struct{float32; float32; float32}:struct{}", ptr null, ptr nonnull @_tinygo_array_hash_f32, ptr null, ptr nonnull @_tinygo_array_cmp_f32, ptr undef) #2
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #2
  %4 = call ptr @runtime.hashmapMakeGeneric(i32 88, i32 0, i32 0, ptr nonnull @"runtime.hashmapType:struct{struct{float32; float64}; struct{bool; uint64; string; interface{}; float32; complex128; interface{Error() string}}; string}:struct{}", ptr null, ptr nonnull @"_tinygo_hash_88{0:16{0:f32,8:f64},16:64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf},80:str}", ptr null, ptr nonnull @"_tinygo_cmp_88{0:16{0:f32,8:f64},16:64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf},80:str}", ptr undef) #2
  call void @runtime.trackPointer(ptr %4, ptr nonnull %stackalloc, ptr undef) #2
  %5 = call ptr @runtime.hashmapMakeGeneric(i32 56, i32 0, i32 0, ptr nonnull @"runtime.hashmapType:struct{[5]interface{}; [16]byte}:struct{}", ptr null, ptr nonnull @"_tinygo_hash_56{0:[5]itf}", ptr null, ptr nonnull @"_tinygo_cmp_56{0:[5]itf}", ptr undef) #2
  call void @runtime.trackPointer(ptr %5, ptr nonnull %stackalloc, ptr undef) #2
  ret void
}

declare i32 @runtime.hash32(ptr, i32, i32, ptr) #0

declare i1 @runtime.memCompareBytes(ptr, ptr, i32, ptr) #0

declare ptr @runtime.hashmapMakeGeneric(i32, i32, i32, ptr, ptr, ptr, ptr, ptr, ptr) #0

; Function Attrs: nounwind
define linkonce_odr i32 @_tinygo_array_hash_f32(ptr readonly captures(none) %key, i32 %size, i32 %seed, ptr %context) unnamed_addr #1 {
entry:
  br label %loop

loop:                                             ; preds = %loop, %entry
  %offset.prev = phi i32 [ %size, %entry ], [ %offset, %loop ]
  %hash.prev = phi i32 [ 0, %entry ], [ %hash, %loop ]
  %offset = add nsw i32 %offset.prev, -4
  %elem = getelementptr inbounds i8, ptr %key, i32 %offset
  %hash.elem = call i32 @runtime.hashmapFloat32Hash(ptr dereferenceable(4) %elem, i32 4, i32 %seed, ptr undef) #2
  %hash.prev.mul = mul i32 %hash.prev, 31
  %hash = xor i32 %hash.prev.mul, %hash.elem
  %done = icmp eq i32 %offset, 0
  br i1 %done, label %ret, label %loop

ret:                                              ; preds = %loop
  ret i32 %hash
}

declare i32 @runtime.hashmapFloat32Hash(ptr, i32, i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr i1 @_tinygo_array_cmp_f32(ptr readonly captures(none) %x, ptr readonly captures(none) %y, i32 %size, ptr %context) unnamed_addr #1 {
entry:
  br label %loop

loop:                                             ; preds = %next, %entry
  %offset.prev = phi i32 [ %size, %entry ], [ %offset, %next ]
  %offset = add nsw i32 %offset.prev, -4
  %x.elem = getelementptr inbounds i8, ptr %x, i32 %offset
  %y.elem = getelementptr inbounds i8, ptr %y, i32 %offset
  %eq.elem = call i1 @runtime.memCompareFloat32(ptr dereferenceable(4) %x.elem, ptr dereferenceable(4) %y.elem, i32 4, ptr undef) #2
  br i1 %eq.elem, label %next, label %ret

next:                                             ; preds = %loop
  %done = icmp eq i32 %offset, 0
  br i1 %done, label %ret, label %loop

ret:                                              ; preds = %next, %loop
  ret i1 %eq.elem
}

declare i1 @runtime.memCompareFloat32(ptr dereferenceable_or_null(4), ptr dereferenceable_or_null(4), i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr i32 @"_tinygo_hash_88{0:16{0:f32,8:f64},16:64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf},80:str}"(ptr readonly captures(none) %key, i32 %size, i32 %seed, ptr %context) unnamed_addr #1 {
  %field.hash = call i32 @"_tinygo_hash_16{0:f32,8:f64}"(ptr dereferenceable(16) %key, i32 16, i32 %seed, ptr undef)
  %field.ptr1 = getelementptr inbounds nuw i8, ptr %key, i32 16
  %field.hash2 = call i32 @"_tinygo_hash_64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf}"(ptr dereferenceable(64) %field.ptr1, i32 64, i32 %seed, ptr undef)
  %1 = mul i32 %field.hash, 31
  %2 = xor i32 %1, %field.hash2
  %field.ptr3 = getelementptr inbounds nuw i8, ptr %key, i32 80
  %field.hash4 = call i32 @runtime.hashmapStringPtrHash(ptr dereferenceable(8) %field.ptr3, i32 8, i32 %seed, ptr undef) #2
  %3 = mul i32 %2, 31
  %4 = xor i32 %3, %field.hash4
  ret i32 %4
}

; Function Attrs: nounwind
define linkonce_odr i32 @"_tinygo_hash_16{0:f32,8:f64}"(ptr readonly captures(none) %key, i32 %size, i32 %seed, ptr %context) unnamed_addr #1 {
  %field.hash = call i32 @runtime.hashmapFloat32Hash(ptr dereferenceable(4) %key, i32 4, i32 %seed, ptr undef) #2
  %field.ptr1 = getelementptr inbounds nuw i8, ptr %key, i32 8
  %field.hash2 = call i32 @runtime.hashmapFloat64Hash(ptr dereferenceable(8) %field.ptr1, i32 8, i32 %seed, ptr undef) #2
  %1 = mul i32 %field.hash, 31
  %2 = xor i32 %1, %field.hash2
  ret i32 %2
}

declare i32 @runtime.hashmapFloat64Hash(ptr, i32, i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr i32 @"_tinygo_hash_64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf}"(ptr readonly captures(none) %key, i32 %size, i32 %seed, ptr %context) unnamed_addr #1 {
  %field.hash = call i32 @runtime.hash32(ptr dereferenceable(1) %key, i32 1, i32 %seed, ptr undef) #2
  %field.ptr1 = getelementptr inbounds nuw i8, ptr %key, i32 8
  %field.hash2 = call i32 @runtime.hash32(ptr dereferenceable(8) %field.ptr1, i32 8, i32 %seed, ptr undef) #2
  %1 = mul i32 %field.hash, 31
  %2 = xor i32 %1, %field.hash2
  %field.ptr3 = getelementptr inbounds nuw i8, ptr %key, i32 16
  %field.hash4 = call i32 @runtime.hashmapStringPtrHash(ptr dereferenceable(8) %field.ptr3, i32 8, i32 %seed, ptr undef) #2
  %3 = mul i32 %2, 31
  %4 = xor i32 %3, %field.hash4
  %field.ptr5 = getelementptr inbounds nuw i8, ptr %key, i32 24
  %field.hash6 = call i32 @runtime.hashmapInterfacePtrHash(ptr dereferenceable(8) %field.ptr5, i32 8, i32 %seed, ptr undef) #2
  %5 = mul i32 %4, 31
  %6 = xor i32 %5, %field.hash6
  %field.ptr7 = getelementptr inbounds nuw i8, ptr %key, i32 32
  %field.hash8 = call i32 @runtime.hashmapFloat32Hash(ptr dereferenceable(4) %field.ptr7, i32 4, i32 %seed, ptr undef) #2
  %7 = mul i32 %6, 31
  %8 = xor i32 %7, %field.hash8
  %field.ptr9 = getelementptr inbounds nuw i8, ptr %key, i32 40
  %field.hash10 = call i32 @_tinygo_array_hash_f64(ptr dereferenceable(16) %field.ptr9, i32 16, i32 %seed, ptr undef)
  %9 = mul i32 %8, 31
  %10 = xor i32 %9, %field.hash10
  %field.ptr11 = getelementptr inbounds nuw i8, ptr %key, i32 56
  %field.hash12 = call i32 @runtime.hashmapInterfacePtrHash(ptr dereferenceable(8) %field.ptr11, i32 8, i32 %seed, ptr undef) #2
  %11 = mul i32 %10, 31
  %12 = xor i32 %11, %field.hash12
  ret i32 %12
}

declare i32 @runtime.hashmapStringPtrHash(ptr, i32, i32, ptr) #0

declare i32 @runtime.hashmapInterfacePtrHash(ptr, i32, i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr i32 @_tinygo_array_hash_f64(ptr readonly captures(none) %key, i32 %size, i32 %seed, ptr %context) unnamed_addr #1 {
entry:
  br label %loop

loop:                                             ; preds = %loop, %entry
  %offset.prev = phi i32 [ %size, %entry ], [ %offset, %loop ]
  %hash.prev = phi i32 [ 0, %entry ], [ %hash, %loop ]
  %offset = add nsw i32 %offset.prev, -8
  %elem = getelementptr inbounds i8, ptr %key, i32 %offset
  %hash.elem = call i32 @runtime.hashmapFloat64Hash(ptr dereferenceable(8) %elem, i32 8, i32 %seed, ptr undef) #2
  %hash.prev.mul = mul i32 %hash.prev, 31
  %hash = xor i32 %hash.prev.mul, %hash.elem
  %done = icmp eq i32 %offset, 0
  br i1 %done, label %ret, label %loop

ret:                                              ; preds = %loop
  ret i32 %hash
}

; Function Attrs: nounwind
define linkonce_odr i1 @"_tinygo_cmp_88{0:16{0:f32,8:f64},16:64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf},80:str}"(ptr readonly captures(none) %x, ptr readonly captures(none) %y, i32 %size, ptr %context) unnamed_addr #1 {
  %x.field = getelementptr inbounds nuw i8, ptr %x, i32 16
  %y.field = getelementptr inbounds nuw i8, ptr %y, i32 16
  %field.eq = call i1 @"_tinygo_cmp_64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf}"(ptr dereferenceable(64) %x.field, ptr dereferenceable(64) %y.field, i32 64, ptr undef)
  br i1 %field.eq, label %1, label %return

1:                                                ; preds = %0
  %field.eq3 = call i1 @"_tinygo_cmp_16{0:f32,8:f64}"(ptr dereferenceable(16) %x, ptr dereferenceable(16) %y, i32 16, ptr undef)
  br i1 %field.eq3, label %2, label %return

2:                                                ; preds = %1
  %x.field4 = getelementptr inbounds nuw i8, ptr %x, i32 80
  %y.field5 = getelementptr inbounds nuw i8, ptr %y, i32 80
  %field.eq6 = call i1 @runtime.memCompareString(ptr dereferenceable(8) %x.field4, ptr dereferenceable(8) %y.field5, i32 8, ptr undef) #2
  br label %return

return:                                           ; preds = %2, %1, %0
  %eq = phi i1 [ false, %0 ], [ false, %1 ], [ %field.eq6, %2 ]
  ret i1 %eq
}

; Function Attrs: nounwind
define linkonce_odr i1 @"_tinygo_cmp_64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf}"(ptr readonly captures(none) %x, ptr readonly captures(none) %y, i32 %size, ptr %context) unnamed_addr #1 {
  %x.field = getelementptr inbounds nuw i8, ptr %x, i32 24
  %y.field = getelementptr inbounds nuw i8, ptr %y, i32 24
  %field.eq = call i1 @runtime.memCompareInterface(ptr dereferenceable(8) %x.field, ptr dereferenceable(8) %y.field, i32 8, ptr undef) #2
  %x.field1 = getelementptr inbounds nuw i8, ptr %x, i32 56
  %y.field2 = getelementptr inbounds nuw i8, ptr %y, i32 56
  %field.eq3 = call i1 @runtime.memCompareInterface(ptr dereferenceable(8) %x.field1, ptr dereferenceable(8) %y.field2, i32 8, ptr undef) #2
  %1 = and i1 %field.eq, %field.eq3
  br i1 %1, label %2, label %return

2:                                                ; preds = %0
  %field.eq6 = call i1 @runtime.memCompareBytes(ptr dereferenceable(1) %x, ptr dereferenceable(1) %y, i32 1, ptr undef) #2
  br i1 %field.eq6, label %3, label %return

3:                                                ; preds = %2
  %x.field7 = getelementptr inbounds nuw i8, ptr %x, i32 8
  %y.field8 = getelementptr inbounds nuw i8, ptr %y, i32 8
  %field.eq9 = call i1 @runtime.memCompareBytes(ptr dereferenceable(8) %x.field7, ptr dereferenceable(8) %y.field8, i32 8, ptr undef) #2
  br i1 %field.eq9, label %4, label %return

4:                                                ; preds = %3
  %x.field10 = getelementptr inbounds nuw i8, ptr %x, i32 32
  %y.field11 = getelementptr inbounds nuw i8, ptr %y, i32 32
  %field.eq12 = call i1 @runtime.memCompareFloat32(ptr dereferenceable(4) %x.field10, ptr dereferenceable(4) %y.field11, i32 4, ptr undef) #2
  br i1 %field.eq12, label %5, label %return

5:                                                ; preds = %4
  %x.field13 = getelementptr inbounds nuw i8, ptr %x, i32 40
  %y.field14 = getelementptr inbounds nuw i8, ptr %y, i32 40
  %field.eq15 = call i1 @_tinygo_array_cmp_f64(ptr dereferenceable(16) %x.field13, ptr dereferenceable(16) %y.field14, i32 16, ptr undef)
  br i1 %field.eq15, label %6, label %return

6:                                                ; preds = %5
  %x.field16 = getelementptr inbounds nuw i8, ptr %x, i32 16
  %y.field17 = getelementptr inbounds nuw i8, ptr %y, i32 16
  %field.eq18 = call i1 @runtime.memCompareString(ptr dereferenceable(8) %x.field16, ptr dereferenceable(8) %y.field17, i32 8, ptr undef) #2
  br label %return

return:                                           ; preds = %6, %5, %4, %3, %2, %0
  %eq = phi i1 [ false, %0 ], [ false, %2 ], [ false, %3 ], [ false, %4 ], [ false, %5 ], [ %field.eq18, %6 ]
  ret i1 %eq
}

declare i1 @runtime.memCompareInterface(ptr dereferenceable_or_null(8), ptr dereferenceable_or_null(8), i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr i1 @_tinygo_array_cmp_f64(ptr readonly captures(none) %x, ptr readonly captures(none) %y, i32 %size, ptr %context) unnamed_addr #1 {
entry:
  br label %loop

loop:                                             ; preds = %next, %entry
  %offset.prev = phi i32 [ %size, %entry ], [ %offset, %next ]
  %offset = add nsw i32 %offset.prev, -8
  %x.elem = getelementptr inbounds i8, ptr %x, i32 %offset
  %y.elem = getelementptr inbounds i8, ptr %y, i32 %offset
  %eq.elem = call i1 @runtime.memCompareFloat64(ptr dereferenceable(8) %x.elem, ptr dereferenceable(8) %y.elem, i32 8, ptr undef) #2
  br i1 %eq.elem, label %next, label %ret

next:                                             ; preds = %loop
  %done = icmp eq i32 %offset, 0
  br i1 %done, label %ret, label %loop

ret:                                              ; preds = %next, %loop
  ret i1 %eq.elem
}

declare i1 @runtime.memCompareFloat64(ptr dereferenceable_or_null(8), ptr dereferenceable_or_null(8), i32, ptr) #0

declare i1 @runtime.memCompareString(ptr dereferenceable_or_null(8), ptr dereferenceable_or_null(8), i32, ptr) #0

; Function Attrs: nounwind
define linkonce_odr i1 @"_tinygo_cmp_16{0:f32,8:f64}"(ptr readonly captures(none) %x, ptr readonly captures(none) %y, i32 %size, ptr %context) unnamed_addr #1 {
  %field.eq = call i1 @runtime.memCompareFloat32(ptr dereferenceable(4) %x, ptr dereferenceable(4) %y, i32 4, ptr undef) #2
  br i1 %field.eq, label %1, label %return

1:                                                ; preds = %0
  %x.field1 = getelementptr inbounds nuw i8, ptr %x, i32 8
  %y.field2 = getelementptr inbounds nuw i8, ptr %y, i32 8
  %field.eq3 = call i1 @runtime.memCompareFloat64(ptr dereferenceable(8) %x.field1, ptr dereferenceable(8) %y.field2, i32 8, ptr undef) #2
  br label %return

return:                                           ; preds = %1, %0
  %eq = phi i1 [ false, %0 ], [ %field.eq3, %1 ]
  ret i1 %eq
}

; Function Attrs: nounwind
define linkonce_odr i32 @"_tinygo_hash_56{0:[5]itf}"(ptr readonly captures(none) %key, i32 %size, i32 %seed, ptr %context) unnamed_addr #1 {
  %field.hash = call i32 @_tinygo_array_hash_itf(ptr dereferenceable(40) %key, i32 40, i32 %seed, ptr undef)
  ret i32 %field.hash
}

; Function Attrs: nounwind
define linkonce_odr i32 @_tinygo_array_hash_itf(ptr readonly captures(none) %key, i32 %size, i32 %seed, ptr %context) unnamed_addr #1 {
entry:
  br label %loop

loop:                                             ; preds = %loop, %entry
  %offset.prev = phi i32 [ %size, %entry ], [ %offset, %loop ]
  %hash.prev = phi i32 [ 0, %entry ], [ %hash, %loop ]
  %offset = add nsw i32 %offset.prev, -8
  %elem = getelementptr inbounds i8, ptr %key, i32 %offset
  %hash.elem = call i32 @runtime.hashmapInterfacePtrHash(ptr dereferenceable(8) %elem, i32 8, i32 %seed, ptr undef) #2
  %hash.prev.mul = mul i32 %hash.prev, 31
  %hash = xor i32 %hash.prev.mul, %hash.elem
  %done = icmp eq i32 %offset, 0
  br i1 %done, label %ret, label %loop

ret:                                              ; preds = %loop
  ret i32 %hash
}

; Function Attrs: nounwind
define linkonce_odr i1 @"_tinygo_cmp_56{0:[5]itf}"(ptr readonly captures(none) %x, ptr readonly captures(none) %y, i32 %size, ptr %context) unnamed_addr #1 {
  %field.eq = call i1 @_tinygo_array_cmp_itf(ptr dereferenceable(40) %x, ptr dereferenceable(40) %y, i32 40, ptr undef)
  ret i1 %field.eq
}

; Function Attrs: nounwind
define linkonce_odr i1 @_tinygo_array_cmp_itf(ptr readonly captures(none) %x, ptr readonly captures(none) %y, i32 %size, ptr %context) unnamed_addr #1 {
entry:
  br label %loop

loop:                                             ; preds = %loop, %entry
  %offset.prev = phi i32 [ %size, %entry ], [ %offset, %loop ]
  %eq.prev = phi i1 [ true, %entry ], [ %eq, %loop ]
  %offset = add nsw i32 %offset.prev, -8
  %x.elem = getelementptr inbounds i8, ptr %x, i32 %offset
  %y.elem = getelementptr inbounds i8, ptr %y, i32 %offset
  %eq.elem = call i1 @runtime.memCompareInterface(ptr dereferenceable(8) %x.elem, ptr dereferenceable(8) %y.elem, i32 8, ptr undef) #2
  %eq = and i1 %eq.prev, %eq.elem
  %done = icmp eq i32 %offset, 0
  br i1 %done, label %ret, label %loop

ret:                                              ; preds = %loop
  ret i1 %eq
}

; Function Attrs: nounwind
define hidden i1 @main.colorEq(i8 %x.r, i8 %x.g, i8 %x.b, i8 %y.r, i8 %y.g, i8 %y.b, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq i8 %x.r, %y.r
  %1 = icmp eq i8 %x.g, %y.g
  %2 = and i1 %0, %1
  %3 = icmp eq i8 %x.b, %y.b
  %4 = and i1 %2, %3
  ret i1 %4
}

; Function Attrs: nounwind
define hidden i1 @main.spriteEq(ptr dereferenceable_or_null(768) %x, ptr dereferenceable_or_null(768) %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = icmp eq ptr %y, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %2 = call i1 @runtime.memCompareBytes(ptr dereferenceable(768) %x, ptr dereferenceable(768) %y, i32 768, ptr undef) #2
  ret i1 %2

deref.throw:                                      ; preds = %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.interfaceArrayEq(ptr dereferenceable_or_null(40) %x, ptr dereferenceable_or_null(40) %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = icmp eq ptr %y, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %2 = call i1 @_tinygo_array_cmp_itf(ptr align 4 dereferenceable(40) %x, ptr align 4 dereferenceable(40) %y, i32 40, ptr undef)
  ret i1 %2

deref.throw:                                      ; preds = %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.triangleEq(ptr dereferenceable_or_null(36) %x, ptr dereferenceable_or_null(36) %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = icmp eq ptr %y, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %2 = call i1 @_tinygo_array_cmp_f32(ptr align 4 dereferenceable(36) %x, ptr align 4 dereferenceable(36) %y, i32 36, ptr undef)
  ret i1 %2

deref.throw:                                      ; preds = %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.optionsEq(ptr dereferenceable_or_null(64) %x, ptr dereferenceable_or_null(64) %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = icmp eq ptr %y, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %2 = call i1 @"_tinygo_cmp_64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf}"(ptr align 8 dereferenceable(64) %x, ptr align 8 dereferenceable(64) %y, i32 64, ptr undef)
  ret i1 %2

deref.throw:                                      ; preds = %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.moreOptionsEq(ptr dereferenceable_or_null(88) %x, ptr dereferenceable_or_null(88) %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = icmp eq ptr %y, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %2 = call i1 @"_tinygo_cmp_88{0:16{0:f32,8:f64},16:64{0:b,8:[8]b,16:str,24:itf,32:f32,40:[2]f64,56:itf},80:str}"(ptr align 8 dereferenceable(88) %x, ptr align 8 dereferenceable(88) %y, i32 88, ptr undef)
  ret i1 %2

deref.throw:                                      ; preds = %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.paddedEq(ptr dereferenceable_or_null(56) %x, ptr dereferenceable_or_null(56) %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = icmp eq ptr %x, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %1 = icmp eq ptr %y, null
  br i1 %1, label %deref.throw, label %deref.next1

deref.next1:                                      ; preds = %deref.next
  %2 = call i1 @"_tinygo_cmp_56{0:[5]itf}"(ptr align 4 dereferenceable(56) %x, ptr align 4 dereferenceable(56) %y, i32 56, ptr undef)
  ret i1 %2

deref.throw:                                      ; preds = %deref.next, %entry
  call void @runtime.nilPanic(ptr undef) #2
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind }
