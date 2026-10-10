; ModuleID = 'string.go'
source_filename = "string.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._string = type { ptr, i32 }
%runtime._interface = type { ptr, ptr }

@"main$string" = internal unnamed_addr constant [3 x i8] c"foo", align 1
@"reflect/types.type:basic:string" = linkonce_odr constant { i8, ptr } { i8 81, ptr @"reflect/types.type:pointer:basic:string" }, align 4
@"reflect/types.type:pointer:basic:string" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:string" }, align 4
@"main$string.1" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.2" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.3" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.4" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.5" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.6" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.7" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.8" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.9" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.10" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.11" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.12" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.13" = internal unnamed_addr constant [3 x i8] c"abc", align 1
@"main$string.14" = internal unnamed_addr constant [3 x i8] c"abc", align 1

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._string @main.someString(ptr %context) unnamed_addr #1 {
entry:
  ret %runtime._string { ptr @"main$string", i32 3 }
}

; Function Attrs: nounwind
define hidden %runtime._string @main.zeroLengthString(ptr %context) unnamed_addr #1 {
entry:
  ret %runtime._string zeroinitializer
}

; Function Attrs: nounwind
define hidden i32 @main.stringLen(ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  ret i32 %s.len
}

; Function Attrs: nounwind
define hidden i8 @main.stringIndex(ptr readonly %s.data, i32 %s.len, i32 %index, ptr %context) unnamed_addr #1 {
entry:
  %.not = icmp ult i32 %index, %s.len
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %entry
  %0 = getelementptr inbounds i8, ptr %s.data, i32 %index
  %1 = load i8, ptr %0, align 1
  ret i8 %1

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i8 undef
}

declare void @runtime.lookupPanic(ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.stringCompareEqual(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #4
  ret i1 %0
}

declare i1 @runtime.stringEqual(ptr readonly, i32, ptr readonly, i32, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.stringCompareUnequal(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #4
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareEqual(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s1.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %s2.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareUnequal(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s1.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %s2.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %s1.data, i32 %s1.len, ptr %s2.data, i32 %s2.len, ptr undef) #4
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareSideEffects(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call { ptr, i32, i32 } @main.mutateBytes(ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr undef)
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #4
  %4 = extractvalue { ptr, i32, i32 } %2, 0
  %5 = extractvalue { ptr, i32, i32 } %2, 1
  call void @runtime.trackPointer(ptr %4, ptr nonnull %stackalloc, ptr undef) #4
  %6 = extractvalue %runtime._string %0, 0
  %7 = extractvalue %runtime._string %0, 1
  %8 = call i1 @runtime.stringEqual(ptr %6, i32 %7, ptr %4, i32 %5, ptr undef) #4
  ret i1 %8
}

declare %runtime._string @runtime.stringFromBytes(ptr nocapture readonly dereferenceable_or_null(1), i32, i32, ptr) #0

; Function Attrs: noinline nounwind
define hidden { ptr, i32, i32 } @main.mutateBytes(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #2 {
entry:
  %0 = icmp eq i32 %s.len, 0
  br i1 %0, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  br i1 false, label %lookup.throw, label %lookup.next3

lookup.next3:                                     ; preds = %lookup.next
  %1 = insertvalue { ptr, i32, i32 } zeroinitializer, ptr %s.data, 0
  %2 = insertvalue { ptr, i32, i32 } %1, i32 %s.len, 1
  %3 = insertvalue { ptr, i32, i32 } %2, i32 %s.cap, 2
  %4 = load i8, ptr %s.data, align 1
  %5 = add i8 %4, 1
  store i8 %5, ptr %s.data, align 1
  ret { ptr, i32, i32 } %3

lookup.throw:                                     ; preds = %lookup.next, %entry
  call void @runtime.lookupPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret { ptr, i32, i32 } undef
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareNil(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %s.data, i32 %s.len, ptr null, i32 0, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareUint8(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareEmpty(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  br i1 false, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 0, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  br i1 false, label %slice.throw, label %slice.next5

slice.next5:                                      ; preds = %slice.next
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %b.data, i32 0, ptr undef) #4
  ret i1 %4

slice.throw:                                      ; preds = %slice.next, %entry
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef
}

declare void @runtime.slicePanic(ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareLocal(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringCompareEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %b.data, i32 %b.len, ptr undef) #4
  %5 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %4, 0
  %6 = insertvalue { i1, %runtime._string } %5, %runtime._string %0, 1
  ret { i1, %runtime._string } %6
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareStore(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr dereferenceable_or_null(8) %dst, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %b.data, i32 %b.len, ptr undef) #4
  %5 = icmp eq ptr %dst, null
  br i1 %5, label %store.throw, label %store.next

store.next:                                       ; preds = %entry
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %dst, align 4
  %dst.repack1 = getelementptr inbounds nuw i8, ptr %dst, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %dst.repack1, align 4
  ret i1 %4

store.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %store.throw
  ret i1 undef
}

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define hidden { i1, %runtime._interface } @main.byteSliceStringCompareBox(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %b.data, i32 %b.len, ptr undef) #4
  %5 = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %5, ptr nonnull %stackalloc, ptr undef) #4
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %5, align 4
  %.repack1 = getelementptr inbounds nuw i8, ptr %5, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %.repack1, align 4
  %6 = insertvalue %runtime._interface { ptr @"reflect/types.type:basic:string", ptr undef }, ptr %5, 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %5, ptr nonnull %stackalloc, ptr undef) #4
  %7 = insertvalue { i1, %runtime._interface } zeroinitializer, i1 %4, 0
  %8 = insertvalue { i1, %runtime._interface } %7, %runtime._interface %6, 1
  ret { i1, %runtime._interface } %8
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #3

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call { ptr, i32, i32 } @main.mutateBytes(ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef)
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %4 = extractvalue %runtime._string %0, 0
  %5 = extractvalue %runtime._string %0, 1
  %6 = call i1 @runtime.stringEqual(ptr %4, i32 %5, ptr %b.data, i32 %b.len, ptr undef) #4
  ret i1 %6
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringCompareAfterMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %b.data, i32 %b.len, ptr undef) #4
  %5 = call { ptr, i32, i32 } @main.mutateBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %6 = extractvalue { ptr, i32, i32 } %5, 0
  call void @runtime.trackPointer(ptr %6, ptr nonnull %stackalloc, ptr undef) #4
  %7 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %4, 0
  %8 = insertvalue { i1, %runtime._string } %7, %runtime._string %0, 1
  ret { i1, %runtime._string } %8
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareSlices(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.highmax = icmp ult i32 %a.cap, 2
  br i1 %slice.highmax, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 2, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %slice.highmax1 = icmp ult i32 %b.cap, 2
  br i1 %slice.highmax1, label %slice.throw, label %slice.next5

slice.next5:                                      ; preds = %slice.next
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %b.data, i32 2, ptr undef) #4
  ret i1 %4

slice.throw:                                      ; preds = %slice.next, %entry
  call void @runtime.slicePanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareLiteral(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr nonnull @"main$string.1", i32 3, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringCompare(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceNamedStringCompare(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef) #4
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden %runtime._string @main.namedByteSliceToString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  ret %runtime._string %0
}

; Function Attrs: nounwind
define hidden %runtime._string @main.namedByteSliceToNamedString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  ret %runtime._string %0
}

; Function Attrs: nounwind
define hidden %runtime._string @main.namedRuneSliceToString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromRunes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  ret %runtime._string %0
}

declare %runtime._string @runtime.stringFromRunes(ptr nocapture readonly, i32, i32, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareLess(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef) #4
  ret i1 %0
}

declare i1 @runtime.stringLess(ptr readonly, i32, ptr readonly, i32, ptr) #0

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringLessEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringLess(ptr %2, i32 %3, ptr %b.data, i32 %b.len, ptr undef) #4
  %5 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %4, 0
  %6 = insertvalue { i1, %runtime._string } %5, %runtime._string %0, 1
  ret { i1, %runtime._string } %6
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringLessMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, i8 %c, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 %c, ptr %b.data, align 1
  call void @runtime.trackPointer(ptr nonnull %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %3 = extractvalue %runtime._string %0, 0
  %4 = extractvalue %runtime._string %0, 1
  %5 = call i1 @runtime.stringLess(ptr %3, i32 %4, ptr nonnull %b.data, i32 %b.len, ptr undef) #4
  ret i1 %5

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringLess(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden { i1, i1 } @main.byteSliceStringLessReuse(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call %runtime._string @runtime.stringFromBytes(ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #4
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #4
  %4 = extractvalue %runtime._string %0, 0
  %5 = extractvalue %runtime._string %0, 1
  %6 = extractvalue %runtime._string %2, 0
  %7 = extractvalue %runtime._string %2, 1
  %8 = call i1 @runtime.stringLess(ptr %4, i32 %5, ptr %6, i32 %7, ptr undef) #4
  %9 = extractvalue %runtime._string %2, 0
  %10 = extractvalue %runtime._string %2, 1
  %11 = extractvalue %runtime._string %0, 0
  %12 = extractvalue %runtime._string %0, 1
  %13 = call i1 @runtime.stringLess(ptr %9, i32 %10, ptr %11, i32 %12, ptr undef) #4
  %14 = insertvalue { i1, i1 } zeroinitializer, i1 %8, 0
  %15 = insertvalue { i1, i1 } %14, i1 %13, 1
  ret { i1, i1 } %15
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringLessLiteral(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr nonnull @"main$string.2", i32 3, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareLessEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %b.data, i32 %b.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringLessEqualEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringLess(ptr %b.data, i32 %b.len, ptr %2, i32 %3, ptr undef) #4
  %5 = xor i1 %4, true
  %6 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %5, 0
  %7 = insertvalue { i1, %runtime._string } %6, %runtime._string %0, 1
  ret { i1, %runtime._string } %7
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringLessEqualMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, i8 %c, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 %c, ptr %b.data, align 1
  call void @runtime.trackPointer(ptr nonnull %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %3 = extractvalue %runtime._string %0, 0
  %4 = extractvalue %runtime._string %0, 1
  %5 = call i1 @runtime.stringLess(ptr nonnull %b.data, i32 %b.len, ptr %3, i32 %4, ptr undef) #4
  %6 = xor i1 %5, true
  ret i1 %6

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringLessEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %b.data, i32 %b.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareGreater(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %b.data, i32 %b.len, ptr %a.data, i32 %a.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringGreaterEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringLess(ptr %b.data, i32 %b.len, ptr %2, i32 %3, ptr undef) #4
  %5 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %4, 0
  %6 = insertvalue { i1, %runtime._string } %5, %runtime._string %0, 1
  ret { i1, %runtime._string } %6
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringGreaterMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, i8 %c, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 %c, ptr %b.data, align 1
  call void @runtime.trackPointer(ptr nonnull %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %3 = extractvalue %runtime._string %0, 0
  %4 = extractvalue %runtime._string %0, 1
  %5 = call i1 @runtime.stringLess(ptr nonnull %b.data, i32 %b.len, ptr %3, i32 %4, ptr undef) #4
  ret i1 %5

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringGreater(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %b.data, i32 %b.len, ptr %a.data, i32 %a.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareGreaterEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef) #4
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringGreaterEqualEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringLess(ptr %2, i32 %3, ptr %b.data, i32 %b.len, ptr undef) #4
  %5 = xor i1 %4, true
  %6 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %5, 0
  %7 = insertvalue { i1, %runtime._string } %6, %runtime._string %0, 1
  ret { i1, %runtime._string } %7
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringGreaterEqualMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, i8 %c, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 %c, ptr %b.data, align 1
  call void @runtime.trackPointer(ptr nonnull %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %3 = extractvalue %runtime._string %0, 0
  %4 = extractvalue %runtime._string %0, 1
  %5 = call i1 @runtime.stringLess(ptr %3, i32 %4, ptr nonnull %b.data, i32 %b.len, ptr undef) #4
  %6 = xor i1 %5, true
  ret i1 %6

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringGreaterEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %b.data, i32 %b.len, ptr undef) #4
  %1 = xor i1 %0, true
  ret i1 %1
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedByteString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %3 = xor i1 %2, true
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %5 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %6 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %7 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %8 = xor i1 %7, true
  %9 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %10 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %11 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %12 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %13 = xor i1 %12, true
  store i1 %0, ptr %complit, align 1
  store i1 %3, ptr %1, align 1
  store i1 %5, ptr %4, align 1
  store i1 %8, ptr %6, align 1
  store i1 %10, ptr %9, align 1
  store i1 %13, ptr %11, align 1
  %14 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack2 = load i1, ptr %.elt1, align 1
  %15 = insertvalue [6 x i1] %14, i1 %.unpack2, 1
  %.elt3 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack4 = load i1, ptr %.elt3, align 1
  %16 = insertvalue [6 x i1] %15, i1 %.unpack4, 2
  %.elt5 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack6 = load i1, ptr %.elt5, align 1
  %17 = insertvalue [6 x i1] %16, i1 %.unpack6, 3
  %.elt7 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack8 = load i1, ptr %.elt7, align 1
  %18 = insertvalue [6 x i1] %17, i1 %.unpack8, 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack10 = load i1, ptr %.elt9, align 1
  %19 = insertvalue [6 x i1] %18, i1 %.unpack10, 5
  ret [6 x i1] %19
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedStringByte(ptr readonly %s.data, i32 %s.len, ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call i1 @runtime.stringEqual(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %3 = xor i1 %2, true
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %5 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %6 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %7 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %8 = xor i1 %7, true
  %9 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %10 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %11 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %12 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %13 = xor i1 %12, true
  store i1 %0, ptr %complit, align 1
  store i1 %3, ptr %1, align 1
  store i1 %5, ptr %4, align 1
  store i1 %8, ptr %6, align 1
  store i1 %10, ptr %9, align 1
  store i1 %13, ptr %11, align 1
  %14 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack2 = load i1, ptr %.elt1, align 1
  %15 = insertvalue [6 x i1] %14, i1 %.unpack2, 1
  %.elt3 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack4 = load i1, ptr %.elt3, align 1
  %16 = insertvalue [6 x i1] %15, i1 %.unpack4, 2
  %.elt5 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack6 = load i1, ptr %.elt5, align 1
  %17 = insertvalue [6 x i1] %16, i1 %.unpack6, 3
  %.elt7 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack8 = load i1, ptr %.elt7, align 1
  %18 = insertvalue [6 x i1] %17, i1 %.unpack8, 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack10 = load i1, ptr %.elt9, align 1
  %19 = insertvalue [6 x i1] %18, i1 %.unpack10, 5
  ret [6 x i1] %19
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedByteLiteral(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr nonnull @"main$string.3", i32 3, ptr undef) #4
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr nonnull @"main$string.4", i32 3, ptr undef) #4
  %3 = xor i1 %2, true
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %5 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr nonnull @"main$string.5", i32 3, ptr undef) #4
  %6 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %7 = call i1 @runtime.stringLess(ptr nonnull @"main$string.6", i32 3, ptr %a.data, i32 %a.len, ptr undef) #4
  %8 = xor i1 %7, true
  %9 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %10 = call i1 @runtime.stringLess(ptr nonnull @"main$string.7", i32 3, ptr %a.data, i32 %a.len, ptr undef) #4
  %11 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %12 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr nonnull @"main$string.8", i32 3, ptr undef) #4
  %13 = xor i1 %12, true
  store i1 %0, ptr %complit, align 1
  store i1 %3, ptr %1, align 1
  store i1 %5, ptr %4, align 1
  store i1 %8, ptr %6, align 1
  store i1 %10, ptr %9, align 1
  store i1 %13, ptr %11, align 1
  %14 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack2 = load i1, ptr %.elt1, align 1
  %15 = insertvalue [6 x i1] %14, i1 %.unpack2, 1
  %.elt3 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack4 = load i1, ptr %.elt3, align 1
  %16 = insertvalue [6 x i1] %15, i1 %.unpack4, 2
  %.elt5 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack6 = load i1, ptr %.elt5, align 1
  %17 = insertvalue [6 x i1] %16, i1 %.unpack6, 3
  %.elt7 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack8 = load i1, ptr %.elt7, align 1
  %18 = insertvalue [6 x i1] %17, i1 %.unpack8, 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack10 = load i1, ptr %.elt9, align 1
  %19 = insertvalue [6 x i1] %18, i1 %.unpack10, 5
  ret [6 x i1] %19
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedLiteralByte(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr nonnull @"main$string.9", i32 3, ptr %a.data, i32 %a.len, ptr undef) #4
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call i1 @runtime.stringEqual(ptr nonnull @"main$string.10", i32 3, ptr %a.data, i32 %a.len, ptr undef) #4
  %3 = xor i1 %2, true
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %5 = call i1 @runtime.stringLess(ptr nonnull @"main$string.11", i32 3, ptr %a.data, i32 %a.len, ptr undef) #4
  %6 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %7 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr nonnull @"main$string.12", i32 3, ptr undef) #4
  %8 = xor i1 %7, true
  %9 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %10 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr nonnull @"main$string.13", i32 3, ptr undef) #4
  %11 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %12 = call i1 @runtime.stringLess(ptr nonnull @"main$string.14", i32 3, ptr %a.data, i32 %a.len, ptr undef) #4
  %13 = xor i1 %12, true
  store i1 %0, ptr %complit, align 1
  store i1 %3, ptr %1, align 1
  store i1 %5, ptr %4, align 1
  store i1 %8, ptr %6, align 1
  store i1 %10, ptr %9, align 1
  store i1 %13, ptr %11, align 1
  %14 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack2 = load i1, ptr %.elt1, align 1
  %15 = insertvalue [6 x i1] %14, i1 %.unpack2, 1
  %.elt3 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack4 = load i1, ptr %.elt3, align 1
  %16 = insertvalue [6 x i1] %15, i1 %.unpack4, 2
  %.elt5 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack6 = load i1, ptr %.elt5, align 1
  %17 = insertvalue [6 x i1] %16, i1 %.unpack6, 3
  %.elt7 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack8 = load i1, ptr %.elt7, align 1
  %18 = insertvalue [6 x i1] %17, i1 %.unpack8, 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack10 = load i1, ptr %.elt9, align 1
  %19 = insertvalue [6 x i1] %18, i1 %.unpack10, 5
  ret [6 x i1] %19
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedNamedByteString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %3 = xor i1 %2, true
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %5 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %6 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %7 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %8 = xor i1 %7, true
  %9 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %10 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %11 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %12 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %13 = xor i1 %12, true
  store i1 %0, ptr %complit, align 1
  store i1 %3, ptr %1, align 1
  store i1 %5, ptr %4, align 1
  store i1 %8, ptr %6, align 1
  store i1 %10, ptr %9, align 1
  store i1 %13, ptr %11, align 1
  %14 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack2 = load i1, ptr %.elt1, align 1
  %15 = insertvalue [6 x i1] %14, i1 %.unpack2, 1
  %.elt3 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack4 = load i1, ptr %.elt3, align 1
  %16 = insertvalue [6 x i1] %15, i1 %.unpack4, 2
  %.elt5 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack6 = load i1, ptr %.elt5, align 1
  %17 = insertvalue [6 x i1] %16, i1 %.unpack6, 3
  %.elt7 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack8 = load i1, ptr %.elt7, align 1
  %18 = insertvalue [6 x i1] %17, i1 %.unpack8, 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack10 = load i1, ptr %.elt9, align 1
  %19 = insertvalue [6 x i1] %18, i1 %.unpack10, 5
  ret [6 x i1] %19
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedNamedStringByte(ptr readonly %s.data, i32 %s.len, ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call i1 @runtime.stringEqual(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %3 = xor i1 %2, true
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %5 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %6 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %7 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %8 = xor i1 %7, true
  %9 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %10 = call i1 @runtime.stringLess(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  %11 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %12 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  %13 = xor i1 %12, true
  store i1 %0, ptr %complit, align 1
  store i1 %3, ptr %1, align 1
  store i1 %5, ptr %4, align 1
  store i1 %8, ptr %6, align 1
  store i1 %10, ptr %9, align 1
  store i1 %13, ptr %11, align 1
  %14 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack2 = load i1, ptr %.elt1, align 1
  %15 = insertvalue [6 x i1] %14, i1 %.unpack2, 1
  %.elt3 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack4 = load i1, ptr %.elt3, align 1
  %16 = insertvalue [6 x i1] %15, i1 %.unpack4, 2
  %.elt5 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack6 = load i1, ptr %.elt5, align 1
  %17 = insertvalue [6 x i1] %16, i1 %.unpack6, 3
  %.elt7 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack8 = load i1, ptr %.elt7, align 1
  %18 = insertvalue [6 x i1] %17, i1 %.unpack8, 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack10 = load i1, ptr %.elt9, align 1
  %19 = insertvalue [6 x i1] %18, i1 %.unpack10, 5
  ret [6 x i1] %19
}

; Function Attrs: nounwind
define hidden i1 @main.mixedLocalByteString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 %s.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.mixedLocalStringByte(ptr readonly %s.data, i32 %s.len, ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %a.data, i32 %a.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.mixedSliceByteString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.highmax = icmp ult i32 %a.cap, 2
  br i1 %slice.highmax, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 2, ptr %s.data, i32 %s.len, ptr undef) #4
  ret i1 %0

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.mixedPrecomputedStringSlice(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.maxcap = icmp ult i32 %s.len, 2
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %0 = call i1 @runtime.stringEqual(ptr %a.data, i32 %a.len, ptr %s.data, i32 2, ptr undef) #4
  ret i1 %0

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden { i1, i1 } @main.mixedByteStringReuse(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %s.data, i32 %s.len, ptr undef) #4
  %5 = extractvalue %runtime._string %0, 0
  %6 = extractvalue %runtime._string %0, 1
  %7 = call i1 @runtime.stringLess(ptr %5, i32 %6, ptr %s.data, i32 %s.len, ptr undef) #4
  %8 = insertvalue { i1, i1 } zeroinitializer, i1 %4, 0
  %9 = insertvalue { i1, i1 } %8, i1 %7, 1
  ret { i1, i1 } %9
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.mixedByteStringEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %s.data, i32 %s.len, ptr undef) #4
  %5 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %4, 0
  %6 = insertvalue { i1, %runtime._string } %5, %runtime._string %0, 1
  ret { i1, %runtime._string } %6
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringStore(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr dereferenceable_or_null(8) %dst, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %s.data, i32 %s.len, ptr undef) #4
  %5 = icmp eq ptr %dst, null
  br i1 %5, label %store.throw, label %store.next

store.next:                                       ; preds = %entry
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %dst, align 4
  %dst.repack1 = getelementptr inbounds nuw i8, ptr %dst, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %dst.repack1, align 4
  ret i1 %4

store.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %store.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden { i1, %runtime._interface } @main.mixedByteStringBox(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %s.data, i32 %s.len, ptr undef) #4
  %5 = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %5, ptr nonnull %stackalloc, ptr undef) #4
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %5, align 4
  %.repack1 = getelementptr inbounds nuw i8, ptr %5, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %.repack1, align 4
  %6 = insertvalue %runtime._interface { ptr @"reflect/types.type:basic:string", ptr undef }, ptr %5, 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %5, ptr nonnull %stackalloc, ptr undef) #4
  %7 = insertvalue { i1, %runtime._interface } zeroinitializer, i1 %4, 0
  %8 = insertvalue { i1, %runtime._interface } %7, %runtime._interface %6, 1
  ret { i1, %runtime._interface } %8
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 122, ptr %b.data, align 1
  %3 = extractvalue %runtime._string %0, 0
  %4 = extractvalue %runtime._string %0, 1
  %5 = call i1 @runtime.stringEqual(ptr %3, i32 %4, ptr %s.data, i32 %s.len, ptr undef) #4
  ret i1 %5

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringCall(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call { ptr, i32, i32 } @main.mutateBytes(ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef)
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #4
  %4 = extractvalue %runtime._string %0, 0
  %5 = extractvalue %runtime._string %0, 1
  %6 = call i1 @runtime.stringLess(ptr %s.data, i32 %s.len, ptr %4, i32 %5, ptr undef) #4
  ret i1 %6
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringAcrossBlock(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr readonly %s.data, i32 %s.len, i1 %mutate, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  br i1 %mutate, label %if.then, label %if.done

if.then:                                          ; preds = %entry
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %if.then
  store i8 122, ptr %b.data, align 1
  br label %if.done

if.done:                                          ; preds = %lookup.next, %entry
  %3 = extractvalue %runtime._string %0, 0
  %4 = extractvalue %runtime._string %0, 1
  %5 = call i1 @runtime.stringEqual(ptr %3, i32 %4, ptr %s.data, i32 %s.len, ptr undef) #4
  ret i1 %5

lookup.throw:                                     ; preds = %if.then
  call void @runtime.lookupPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringLoad(ptr %a.data, i32 %a.len, i32 %a.cap, ptr dereferenceable_or_null(8) %s, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = icmp eq ptr %s, null
  br i1 %2, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load ptr, ptr %s, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %s, i32 4
  %.unpack2 = load i32, ptr %.elt1, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #4
  %3 = extractvalue %runtime._string %0, 0
  %4 = extractvalue %runtime._string %0, 1
  %5 = call i1 @runtime.stringEqual(ptr %3, i32 %4, ptr %.unpack, i32 %.unpack2, ptr undef) #4
  ret i1 %5

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.mixedStringLoadByte(ptr dereferenceable_or_null(8) %s, ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = icmp eq ptr %s, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load ptr, ptr %s, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %s, i32 4
  %.unpack2 = load i32, ptr %.elt1, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #4
  %1 = call i1 @runtime.stringEqual(ptr %.unpack, i32 %.unpack2, ptr %a.data, i32 %a.len, ptr undef) #4
  ret i1 %1

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringSlice(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %slice.maxcap = icmp ult i32 %s.len, 2
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %s.data, i32 2, ptr undef) #4
  ret i1 %4

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringChangeType(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = extractvalue %runtime._string %0, 0
  %3 = extractvalue %runtime._string %0, 1
  %4 = call i1 @runtime.stringEqual(ptr %2, i32 %3, ptr %s.data, i32 %s.len, ptr undef) #4
  ret i1 %4
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringInterface(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %s.typecode, ptr %s.value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #4
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #4
  %2 = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %2, ptr nonnull %stackalloc, ptr undef) #4
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %2, align 4
  %.repack1 = getelementptr inbounds nuw i8, ptr %2, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %.repack1, align 4
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %stackalloc, ptr undef) #4
  call void @runtime.trackPointer(ptr nonnull %2, ptr nonnull %stackalloc, ptr undef) #4
  %3 = call i1 @runtime.interfaceEqual(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %2, ptr %s.typecode, ptr %s.value, ptr undef) #4
  ret i1 %3
}

declare i1 @runtime.interfaceEqual(ptr, ptr, ptr, ptr, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.stringCompareLarger(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @runtime.stringLess(ptr %s2.data, i32 %s2.len, ptr %s1.data, i32 %s1.len, ptr undef) #4
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i8 @main.stringLookup(ptr readonly %s.data, i32 %s.len, i8 %x, ptr %context) unnamed_addr #1 {
entry:
  %0 = zext i8 %x to i32
  %.not = icmp ugt i32 %s.len, %0
  br i1 %.not, label %lookup.next, label %lookup.throw

lookup.next:                                      ; preds = %entry
  %1 = getelementptr inbounds nuw i8, ptr %s.data, i32 %0
  %2 = load i8, ptr %1, align 1
  ret i8 %2

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #4
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i8 undef
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { noinline nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #4 = { nounwind }
