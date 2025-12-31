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
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i8 undef
}

declare void @runtime.lookupPanic(ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.stringCompareEqual(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %streq.len.eq = icmp eq i32 %s1.len, %s2.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %s1.data, ptr %s2.data, i32 %s1.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

declare i32 @memcmp(ptr nocapture readonly, ptr nocapture readonly, i32)

; Function Attrs: nounwind
define hidden i1 @main.stringCompareUnequal(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %streq.len.eq = icmp eq i32 %s1.len, %s2.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %s1.data, ptr %s2.data, i32 %s1.len) #5
  %streq.memcmp.eq = icmp ne i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %streq.not = phi i1 [ true, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %streq.not
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareEqual(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s1.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %s2.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %s1.len, %s2.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %s1.data, ptr %s2.data, i32 %s1.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareUnequal(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s1.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %s2.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %s1.len, %s2.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %s1.data, ptr %s2.data, i32 %s1.len) #5
  %streq.memcmp.eq = icmp ne i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %streq.not = phi i1 [ true, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %streq.not
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareSideEffects(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %s1.data, i32 %s1.len, i32 %s1.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = call { ptr, i32, i32 } @main.mutateBytes(ptr %s2.data, i32 %s2.len, i32 %s2.cap, ptr undef)
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #5
  %4 = extractvalue { ptr, i32, i32 } %2, 0
  %5 = extractvalue { ptr, i32, i32 } %2, 1
  call void @runtime.trackPointer(ptr %4, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %5
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %4, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %6 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %6
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
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret { ptr, i32, i32 } undef
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareNil(ptr %s.data, i32 %s.len, i32 %s.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %s.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %s.len, 0
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %s.data, ptr null, i32 %s.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareUint8(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, %b.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %b.data, i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareEmpty(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  br i1 false, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 0, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  br i1 false, label %slice.throw, label %slice.next3

slice.next3:                                      ; preds = %slice.next
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, 0
  br i1 %streq.len.eq, label %streq.body, label %streq.next

slice.throw:                                      ; preds = %slice.next, %entry
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef

streq.body:                                       ; preds = %slice.next3
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %b.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %slice.next3
  %2 = phi i1 [ false, %slice.next3 ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %2
}

declare void @runtime.slicePanic(ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareLocal(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, %b.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %b.data, i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringCompareEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %b.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %b.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %2 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  %3 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %2, 0
  %4 = insertvalue { i1, %runtime._string } %3, %runtime._string %0, 1
  ret { i1, %runtime._string } %4
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareStore(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr dereferenceable_or_null(8) %dst, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %b.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %b.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %2 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  %3 = icmp eq ptr %dst, null
  br i1 %3, label %store.throw, label %store.next

store.next:                                       ; preds = %streq.next
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %dst, align 4
  %dst.repack1 = getelementptr inbounds nuw i8, ptr %dst, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %dst.repack1, align 4
  ret i1 %2

store.throw:                                      ; preds = %streq.next
  call void @runtime.nilPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %store.throw
  ret i1 undef
}

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define hidden { i1, %runtime._interface } @main.byteSliceStringCompareBox(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %b.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %b.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %2 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  %3 = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %3, ptr nonnull %stackalloc, ptr undef) #5
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %3, align 4
  %.repack1 = getelementptr inbounds nuw i8, ptr %3, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %.repack1, align 4
  %4 = insertvalue %runtime._interface { ptr @"reflect/types.type:basic:string", ptr undef }, ptr %3, 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %3, ptr nonnull %stackalloc, ptr undef) #5
  %5 = insertvalue { i1, %runtime._interface } zeroinitializer, i1 %2, 0
  %6 = insertvalue { i1, %runtime._interface } %5, %runtime._interface %4, 1
  ret { i1, %runtime._interface } %6
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #3

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = call { ptr, i32, i32 } @main.mutateBytes(ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef)
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %b.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %b.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %4 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %4
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringCompareAfterMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %b.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %b.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %2 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  %3 = call { ptr, i32, i32 } @main.mutateBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef)
  %4 = extractvalue { ptr, i32, i32 } %3, 0
  call void @runtime.trackPointer(ptr %4, ptr nonnull %stackalloc, ptr undef) #5
  %5 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %2, 0
  %6 = insertvalue { i1, %runtime._string } %5, %runtime._string %0, 1
  ret { i1, %runtime._string } %6
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareSlices(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.highmax = icmp ult i32 %a.cap, 2
  br i1 %slice.highmax, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 2, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %slice.highmax1 = icmp ult i32 %b.cap, 2
  br i1 %slice.highmax1, label %slice.throw, label %slice.next3

slice.next3:                                      ; preds = %slice.next
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, 2
  br i1 %streq.len.eq, label %streq.body, label %streq.next

slice.throw:                                      ; preds = %slice.next, %entry
  call void @runtime.slicePanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef

streq.body:                                       ; preds = %slice.next3
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %b.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %slice.next3
  %2 = phi i1 [ false, %slice.next3 ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %2
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareLiteral(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, 3
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr nonnull @"main$string.1", i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringCompare(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, %b.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %b.data, i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceNamedStringCompare(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, %b.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %b.data, i32 %a.len) #5
  %streq.memcmp.eq = icmp ne i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %streq.not = phi i1 [ true, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %streq.not
}

; Function Attrs: nounwind
define hidden %runtime._string @main.namedByteSliceToString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  ret %runtime._string %0
}

; Function Attrs: nounwind
define hidden %runtime._string @main.namedByteSliceToNamedString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  ret %runtime._string %0
}

; Function Attrs: nounwind
define hidden %runtime._string @main.namedRuneSliceToString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromRunes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  ret %runtime._string %0
}

declare %runtime._string @runtime.stringFromRunes(ptr nocapture readonly, i32, i32, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareLess(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %a.len, i32 %b.len)
  %strlt.memcmp = call i32 @memcmp(ptr %a.data, ptr %b.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %a.len, %b.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nocallback nofree nounwind speculatable willreturn memory(none)
declare i32 @llvm.umin.i32(i32, i32) #4

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringLessEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.lhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %strlt.lhs.len, i32 %b.len)
  %strlt.lhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr %strlt.lhs.ptr, ptr %b.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %strlt.lhs.len, %b.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %2 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %strlt.result, 0
  %3 = insertvalue { i1, %runtime._string } %2, %runtime._string %0, 1
  ret { i1, %runtime._string } %3
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringLessMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, i8 %c, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 %c, ptr %b.data, align 1
  call void @runtime.trackPointer(ptr nonnull %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.lhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %strlt.lhs.len, i32 %b.len)
  %strlt.lhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr %strlt.lhs.ptr, ptr nonnull %b.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %strlt.lhs.len, %b.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringLess(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %a.len, i32 %b.len)
  %strlt.memcmp = call i32 @memcmp(ptr %a.data, ptr %b.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %a.len, %b.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden { i1, i1 } @main.byteSliceStringLessReuse(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = call %runtime._string @runtime.stringFromBytes(ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef) #5
  %3 = extractvalue %runtime._string %2, 0
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.lhs.len = extractvalue %runtime._string %0, 1
  %strlt.rhs.len = extractvalue %runtime._string %2, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %strlt.lhs.len, i32 %strlt.rhs.len)
  %strlt.lhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.rhs.ptr = extractvalue %runtime._string %2, 0
  %strlt.memcmp = call i32 @memcmp(ptr %strlt.lhs.ptr, ptr %strlt.rhs.ptr, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %strlt.lhs.len, %strlt.rhs.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %strlt.lhs.len1 = extractvalue %runtime._string %2, 1
  %strlt.rhs.len2 = extractvalue %runtime._string %0, 1
  %strlt.min.len3 = call i32 @llvm.umin.i32(i32 %strlt.lhs.len1, i32 %strlt.rhs.len2)
  %strlt.lhs.ptr4 = extractvalue %runtime._string %2, 0
  %strlt.rhs.ptr5 = extractvalue %runtime._string %0, 0
  %strlt.memcmp6 = call i32 @memcmp(ptr %strlt.lhs.ptr4, ptr %strlt.rhs.ptr5, i32 %strlt.min.len3) #5
  %strlt.memcmp.eq7 = icmp eq i32 %strlt.memcmp6, 0
  %strlt.len.lt8 = icmp ult i32 %strlt.lhs.len1, %strlt.rhs.len2
  %strlt.memcmp.lt9 = icmp slt i32 %strlt.memcmp6, 0
  %strlt.result10 = select i1 %strlt.memcmp.eq7, i1 %strlt.len.lt8, i1 %strlt.memcmp.lt9
  %4 = insertvalue { i1, i1 } zeroinitializer, i1 %strlt.result, 0
  %5 = insertvalue { i1, i1 } %4, i1 %strlt.result10, 1
  ret { i1, i1 } %5
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringLessLiteral(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %a.len, i32 3)
  %strlt.memcmp = call i32 @memcmp(ptr %a.data, ptr nonnull @"main$string.2", i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %a.len, 3
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareLessEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %b.len, i32 %a.len)
  %strlt.memcmp = call i32 @memcmp(ptr %b.data, ptr %a.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp uge i32 %b.len, %a.len
  %strlt.memcmp.lt = icmp sge i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringLessEqualEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.rhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %b.len, i32 %strlt.rhs.len)
  %strlt.rhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr %b.data, ptr %strlt.rhs.ptr, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp uge i32 %b.len, %strlt.rhs.len
  %strlt.memcmp.lt = icmp sge i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %2 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %strlt.result, 0
  %3 = insertvalue { i1, %runtime._string } %2, %runtime._string %0, 1
  ret { i1, %runtime._string } %3
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringLessEqualMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, i8 %c, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 %c, ptr %b.data, align 1
  call void @runtime.trackPointer(ptr nonnull %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.rhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %b.len, i32 %strlt.rhs.len)
  %strlt.rhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr nonnull %b.data, ptr %strlt.rhs.ptr, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp uge i32 %b.len, %strlt.rhs.len
  %strlt.memcmp.lt = icmp sge i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringLessEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %b.len, i32 %a.len)
  %strlt.memcmp = call i32 @memcmp(ptr %b.data, ptr %a.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp uge i32 %b.len, %a.len
  %strlt.memcmp.lt = icmp sge i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareGreater(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %b.len, i32 %a.len)
  %strlt.memcmp = call i32 @memcmp(ptr %b.data, ptr %a.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %b.len, %a.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringGreaterEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.rhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %b.len, i32 %strlt.rhs.len)
  %strlt.rhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr %b.data, ptr %strlt.rhs.ptr, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %b.len, %strlt.rhs.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %2 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %strlt.result, 0
  %3 = insertvalue { i1, %runtime._string } %2, %runtime._string %0, 1
  ret { i1, %runtime._string } %3
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringGreaterMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, i8 %c, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 %c, ptr %b.data, align 1
  call void @runtime.trackPointer(ptr nonnull %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.rhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %b.len, i32 %strlt.rhs.len)
  %strlt.rhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr nonnull %b.data, ptr %strlt.rhs.ptr, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %b.len, %strlt.rhs.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringGreater(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %b.len, i32 %a.len)
  %strlt.memcmp = call i32 @memcmp(ptr %b.data, ptr %a.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %b.len, %a.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringCompareGreaterEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %a.len, i32 %b.len)
  %strlt.memcmp = call i32 @memcmp(ptr %a.data, ptr %b.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp uge i32 %a.len, %b.len
  %strlt.memcmp.lt = icmp sge i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.byteSliceStringGreaterEqualEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.lhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %strlt.lhs.len, i32 %b.len)
  %strlt.lhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr %strlt.lhs.ptr, ptr %b.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp uge i32 %strlt.lhs.len, %b.len
  %strlt.memcmp.lt = icmp sge i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %2 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %strlt.result, 0
  %3 = insertvalue { i1, %runtime._string } %2, %runtime._string %0, 1
  ret { i1, %runtime._string } %3
}

; Function Attrs: nounwind
define hidden i1 @main.byteSliceStringGreaterEqualMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, i8 %c, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 %c, ptr %b.data, align 1
  call void @runtime.trackPointer(ptr nonnull %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.lhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %strlt.lhs.len, i32 %b.len)
  %strlt.lhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr %strlt.lhs.ptr, ptr nonnull %b.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp uge i32 %strlt.lhs.len, %b.len
  %strlt.memcmp.lt = icmp sge i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden i1 @main.namedByteSliceStringGreaterEqual(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %b.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %a.len, i32 %b.len)
  %strlt.memcmp = call i32 @memcmp(ptr %a.data, ptr %b.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp uge i32 %a.len, %b.len
  %strlt.memcmp.lt = icmp sge i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedByteString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq3 = icmp eq i32 %a.len, %s.len
  br i1 %streq.len.eq3, label %streq.body4, label %streq.next5

streq.body4:                                      ; preds = %streq.next
  %streq.memcmp8 = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %a.len) #5
  %streq.memcmp.eq9 = icmp ne i32 %streq.memcmp8, 0
  br label %streq.next5

streq.next5:                                      ; preds = %streq.body4, %streq.next
  %streq.not = phi i1 [ true, %streq.next ], [ %streq.memcmp.eq9, %streq.body4 ]
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %2 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %a.len, i32 %s.len)
  %strlt.memcmp = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %a.len, %s.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %3 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len12 = call i32 @llvm.umin.i32(i32 %s.len, i32 %a.len)
  %strlt.memcmp15 = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %strlt.min.len12) #5
  %strlt.memcmp.eq16 = icmp eq i32 %strlt.memcmp15, 0
  %strlt.len.lt17 = icmp uge i32 %s.len, %a.len
  %strlt.memcmp.lt18 = icmp sge i32 %strlt.memcmp15, 0
  %strlt.result19 = select i1 %strlt.memcmp.eq16, i1 %strlt.len.lt17, i1 %strlt.memcmp.lt18
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len22 = call i32 @llvm.umin.i32(i32 %s.len, i32 %a.len)
  %strlt.memcmp25 = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %strlt.min.len22) #5
  %strlt.memcmp.eq26 = icmp eq i32 %strlt.memcmp25, 0
  %strlt.len.lt27 = icmp ult i32 %s.len, %a.len
  %strlt.memcmp.lt28 = icmp slt i32 %strlt.memcmp25, 0
  %strlt.result29 = select i1 %strlt.memcmp.eq26, i1 %strlt.len.lt27, i1 %strlt.memcmp.lt28
  %5 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len32 = call i32 @llvm.umin.i32(i32 %a.len, i32 %s.len)
  %strlt.memcmp35 = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %strlt.min.len32) #5
  %strlt.memcmp.eq36 = icmp eq i32 %strlt.memcmp35, 0
  %strlt.len.lt37 = icmp uge i32 %a.len, %s.len
  %strlt.memcmp.lt38 = icmp sge i32 %strlt.memcmp35, 0
  %strlt.result39 = select i1 %strlt.memcmp.eq36, i1 %strlt.len.lt37, i1 %strlt.memcmp.lt38
  store i1 %0, ptr %complit, align 1
  store i1 %streq.not, ptr %1, align 1
  store i1 %strlt.result, ptr %2, align 1
  store i1 %strlt.result19, ptr %3, align 1
  store i1 %strlt.result29, ptr %4, align 1
  store i1 %strlt.result39, ptr %5, align 1
  %6 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt41 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack42 = load i1, ptr %.elt41, align 1
  %7 = insertvalue [6 x i1] %6, i1 %.unpack42, 1
  %.elt43 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack44 = load i1, ptr %.elt43, align 1
  %8 = insertvalue [6 x i1] %7, i1 %.unpack44, 2
  %.elt45 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack46 = load i1, ptr %.elt45, align 1
  %9 = insertvalue [6 x i1] %8, i1 %.unpack46, 3
  %.elt47 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack48 = load i1, ptr %.elt47, align 1
  %10 = insertvalue [6 x i1] %9, i1 %.unpack48, 4
  %.elt49 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack50 = load i1, ptr %.elt49, align 1
  %11 = insertvalue [6 x i1] %10, i1 %.unpack50, 5
  ret [6 x i1] %11
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedStringByte(ptr readonly %s.data, i32 %s.len, ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %s.len, %a.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %s.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq3 = icmp eq i32 %s.len, %a.len
  br i1 %streq.len.eq3, label %streq.body4, label %streq.next5

streq.body4:                                      ; preds = %streq.next
  %streq.memcmp8 = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %s.len) #5
  %streq.memcmp.eq9 = icmp ne i32 %streq.memcmp8, 0
  br label %streq.next5

streq.next5:                                      ; preds = %streq.body4, %streq.next
  %streq.not = phi i1 [ true, %streq.next ], [ %streq.memcmp.eq9, %streq.body4 ]
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %2 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %s.len, i32 %a.len)
  %strlt.memcmp = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %s.len, %a.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %3 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len12 = call i32 @llvm.umin.i32(i32 %a.len, i32 %s.len)
  %strlt.memcmp15 = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %strlt.min.len12) #5
  %strlt.memcmp.eq16 = icmp eq i32 %strlt.memcmp15, 0
  %strlt.len.lt17 = icmp uge i32 %a.len, %s.len
  %strlt.memcmp.lt18 = icmp sge i32 %strlt.memcmp15, 0
  %strlt.result19 = select i1 %strlt.memcmp.eq16, i1 %strlt.len.lt17, i1 %strlt.memcmp.lt18
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len22 = call i32 @llvm.umin.i32(i32 %a.len, i32 %s.len)
  %strlt.memcmp25 = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %strlt.min.len22) #5
  %strlt.memcmp.eq26 = icmp eq i32 %strlt.memcmp25, 0
  %strlt.len.lt27 = icmp ult i32 %a.len, %s.len
  %strlt.memcmp.lt28 = icmp slt i32 %strlt.memcmp25, 0
  %strlt.result29 = select i1 %strlt.memcmp.eq26, i1 %strlt.len.lt27, i1 %strlt.memcmp.lt28
  %5 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len32 = call i32 @llvm.umin.i32(i32 %s.len, i32 %a.len)
  %strlt.memcmp35 = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %strlt.min.len32) #5
  %strlt.memcmp.eq36 = icmp eq i32 %strlt.memcmp35, 0
  %strlt.len.lt37 = icmp uge i32 %s.len, %a.len
  %strlt.memcmp.lt38 = icmp sge i32 %strlt.memcmp35, 0
  %strlt.result39 = select i1 %strlt.memcmp.eq36, i1 %strlt.len.lt37, i1 %strlt.memcmp.lt38
  store i1 %0, ptr %complit, align 1
  store i1 %streq.not, ptr %1, align 1
  store i1 %strlt.result, ptr %2, align 1
  store i1 %strlt.result19, ptr %3, align 1
  store i1 %strlt.result29, ptr %4, align 1
  store i1 %strlt.result39, ptr %5, align 1
  %6 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt41 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack42 = load i1, ptr %.elt41, align 1
  %7 = insertvalue [6 x i1] %6, i1 %.unpack42, 1
  %.elt43 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack44 = load i1, ptr %.elt43, align 1
  %8 = insertvalue [6 x i1] %7, i1 %.unpack44, 2
  %.elt45 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack46 = load i1, ptr %.elt45, align 1
  %9 = insertvalue [6 x i1] %8, i1 %.unpack46, 3
  %.elt47 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack48 = load i1, ptr %.elt47, align 1
  %10 = insertvalue [6 x i1] %9, i1 %.unpack48, 4
  %.elt49 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack50 = load i1, ptr %.elt49, align 1
  %11 = insertvalue [6 x i1] %10, i1 %.unpack50, 5
  ret [6 x i1] %11
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedByteLiteral(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, 3
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr nonnull @"main$string.3", i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq2 = icmp eq i32 %a.len, 3
  br i1 %streq.len.eq2, label %streq.body3, label %streq.next4

streq.body3:                                      ; preds = %streq.next
  %streq.memcmp6 = call i32 @memcmp(ptr %a.data, ptr nonnull @"main$string.4", i32 %a.len) #5
  %streq.memcmp.eq7 = icmp ne i32 %streq.memcmp6, 0
  br label %streq.next4

streq.next4:                                      ; preds = %streq.body3, %streq.next
  %streq.not = phi i1 [ true, %streq.next ], [ %streq.memcmp.eq7, %streq.body3 ]
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %2 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %a.len, i32 3)
  %strlt.memcmp = call i32 @memcmp(ptr %a.data, ptr nonnull @"main$string.5", i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %a.len, 3
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %3 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len8 = call i32 @llvm.umin.i32(i32 %a.len, i32 3)
  %strlt.memcmp9 = call i32 @memcmp(ptr nonnull @"main$string.6", ptr %a.data, i32 %strlt.min.len8) #5
  %strlt.memcmp.eq10 = icmp eq i32 %strlt.memcmp9, 0
  %strlt.len.lt11 = icmp ule i32 %a.len, 3
  %strlt.memcmp.lt12 = icmp sge i32 %strlt.memcmp9, 0
  %strlt.result13 = select i1 %strlt.memcmp.eq10, i1 %strlt.len.lt11, i1 %strlt.memcmp.lt12
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len15 = call i32 @llvm.umin.i32(i32 %a.len, i32 3)
  %strlt.memcmp17 = call i32 @memcmp(ptr nonnull @"main$string.7", ptr %a.data, i32 %strlt.min.len15) #5
  %strlt.memcmp.eq18 = icmp eq i32 %strlt.memcmp17, 0
  %strlt.len.lt19 = icmp ugt i32 %a.len, 3
  %strlt.memcmp.lt20 = icmp slt i32 %strlt.memcmp17, 0
  %strlt.result21 = select i1 %strlt.memcmp.eq18, i1 %strlt.len.lt19, i1 %strlt.memcmp.lt20
  %5 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len23 = call i32 @llvm.umin.i32(i32 %a.len, i32 3)
  %strlt.memcmp25 = call i32 @memcmp(ptr %a.data, ptr nonnull @"main$string.8", i32 %strlt.min.len23) #5
  %strlt.memcmp.eq26 = icmp eq i32 %strlt.memcmp25, 0
  %strlt.len.lt27 = icmp uge i32 %a.len, 3
  %strlt.memcmp.lt28 = icmp sge i32 %strlt.memcmp25, 0
  %strlt.result29 = select i1 %strlt.memcmp.eq26, i1 %strlt.len.lt27, i1 %strlt.memcmp.lt28
  store i1 %0, ptr %complit, align 1
  store i1 %streq.not, ptr %1, align 1
  store i1 %strlt.result, ptr %2, align 1
  store i1 %strlt.result13, ptr %3, align 1
  store i1 %strlt.result21, ptr %4, align 1
  store i1 %strlt.result29, ptr %5, align 1
  %6 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt31 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack32 = load i1, ptr %.elt31, align 1
  %7 = insertvalue [6 x i1] %6, i1 %.unpack32, 1
  %.elt33 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack34 = load i1, ptr %.elt33, align 1
  %8 = insertvalue [6 x i1] %7, i1 %.unpack34, 2
  %.elt35 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack36 = load i1, ptr %.elt35, align 1
  %9 = insertvalue [6 x i1] %8, i1 %.unpack36, 3
  %.elt37 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack38 = load i1, ptr %.elt37, align 1
  %10 = insertvalue [6 x i1] %9, i1 %.unpack38, 4
  %.elt39 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack40 = load i1, ptr %.elt39, align 1
  %11 = insertvalue [6 x i1] %10, i1 %.unpack40, 5
  ret [6 x i1] %11
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedLiteralByte(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, 3
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr noundef nonnull dereferenceable(3) @"main$string.9", ptr noundef nonnull dereferenceable(3) %a.data, i32 3) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq2 = icmp eq i32 %a.len, 3
  br i1 %streq.len.eq2, label %streq.body3, label %streq.next4

streq.body3:                                      ; preds = %streq.next
  %streq.memcmp6 = call i32 @memcmp(ptr noundef nonnull dereferenceable(3) @"main$string.10", ptr noundef nonnull dereferenceable(3) %a.data, i32 3) #5
  %streq.memcmp.eq7 = icmp ne i32 %streq.memcmp6, 0
  br label %streq.next4

streq.next4:                                      ; preds = %streq.body3, %streq.next
  %streq.not = phi i1 [ true, %streq.next ], [ %streq.memcmp.eq7, %streq.body3 ]
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %2 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %a.len, i32 3)
  %strlt.memcmp = call i32 @memcmp(ptr nonnull @"main$string.11", ptr %a.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ugt i32 %a.len, 3
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %3 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len8 = call i32 @llvm.umin.i32(i32 %a.len, i32 3)
  %strlt.memcmp9 = call i32 @memcmp(ptr %a.data, ptr nonnull @"main$string.12", i32 %strlt.min.len8) #5
  %strlt.memcmp.eq10 = icmp eq i32 %strlt.memcmp9, 0
  %strlt.len.lt11 = icmp uge i32 %a.len, 3
  %strlt.memcmp.lt12 = icmp sge i32 %strlt.memcmp9, 0
  %strlt.result13 = select i1 %strlt.memcmp.eq10, i1 %strlt.len.lt11, i1 %strlt.memcmp.lt12
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len15 = call i32 @llvm.umin.i32(i32 %a.len, i32 3)
  %strlt.memcmp17 = call i32 @memcmp(ptr %a.data, ptr nonnull @"main$string.13", i32 %strlt.min.len15) #5
  %strlt.memcmp.eq18 = icmp eq i32 %strlt.memcmp17, 0
  %strlt.len.lt19 = icmp ult i32 %a.len, 3
  %strlt.memcmp.lt20 = icmp slt i32 %strlt.memcmp17, 0
  %strlt.result21 = select i1 %strlt.memcmp.eq18, i1 %strlt.len.lt19, i1 %strlt.memcmp.lt20
  %5 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len23 = call i32 @llvm.umin.i32(i32 %a.len, i32 3)
  %strlt.memcmp25 = call i32 @memcmp(ptr nonnull @"main$string.14", ptr %a.data, i32 %strlt.min.len23) #5
  %strlt.memcmp.eq26 = icmp eq i32 %strlt.memcmp25, 0
  %strlt.len.lt27 = icmp ule i32 %a.len, 3
  %strlt.memcmp.lt28 = icmp sge i32 %strlt.memcmp25, 0
  %strlt.result29 = select i1 %strlt.memcmp.eq26, i1 %strlt.len.lt27, i1 %strlt.memcmp.lt28
  store i1 %0, ptr %complit, align 1
  store i1 %streq.not, ptr %1, align 1
  store i1 %strlt.result, ptr %2, align 1
  store i1 %strlt.result13, ptr %3, align 1
  store i1 %strlt.result21, ptr %4, align 1
  store i1 %strlt.result29, ptr %5, align 1
  %6 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt31 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack32 = load i1, ptr %.elt31, align 1
  %7 = insertvalue [6 x i1] %6, i1 %.unpack32, 1
  %.elt33 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack34 = load i1, ptr %.elt33, align 1
  %8 = insertvalue [6 x i1] %7, i1 %.unpack34, 2
  %.elt35 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack36 = load i1, ptr %.elt35, align 1
  %9 = insertvalue [6 x i1] %8, i1 %.unpack36, 3
  %.elt37 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack38 = load i1, ptr %.elt37, align 1
  %10 = insertvalue [6 x i1] %9, i1 %.unpack38, 4
  %.elt39 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack40 = load i1, ptr %.elt39, align 1
  %11 = insertvalue [6 x i1] %10, i1 %.unpack40, 5
  ret [6 x i1] %11
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedNamedByteString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq3 = icmp eq i32 %a.len, %s.len
  br i1 %streq.len.eq3, label %streq.body4, label %streq.next5

streq.body4:                                      ; preds = %streq.next
  %streq.memcmp8 = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %a.len) #5
  %streq.memcmp.eq9 = icmp ne i32 %streq.memcmp8, 0
  br label %streq.next5

streq.next5:                                      ; preds = %streq.body4, %streq.next
  %streq.not = phi i1 [ true, %streq.next ], [ %streq.memcmp.eq9, %streq.body4 ]
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %2 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %a.len, i32 %s.len)
  %strlt.memcmp = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %a.len, %s.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %3 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len12 = call i32 @llvm.umin.i32(i32 %s.len, i32 %a.len)
  %strlt.memcmp15 = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %strlt.min.len12) #5
  %strlt.memcmp.eq16 = icmp eq i32 %strlt.memcmp15, 0
  %strlt.len.lt17 = icmp uge i32 %s.len, %a.len
  %strlt.memcmp.lt18 = icmp sge i32 %strlt.memcmp15, 0
  %strlt.result19 = select i1 %strlt.memcmp.eq16, i1 %strlt.len.lt17, i1 %strlt.memcmp.lt18
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len22 = call i32 @llvm.umin.i32(i32 %s.len, i32 %a.len)
  %strlt.memcmp25 = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %strlt.min.len22) #5
  %strlt.memcmp.eq26 = icmp eq i32 %strlt.memcmp25, 0
  %strlt.len.lt27 = icmp ult i32 %s.len, %a.len
  %strlt.memcmp.lt28 = icmp slt i32 %strlt.memcmp25, 0
  %strlt.result29 = select i1 %strlt.memcmp.eq26, i1 %strlt.len.lt27, i1 %strlt.memcmp.lt28
  %5 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len32 = call i32 @llvm.umin.i32(i32 %a.len, i32 %s.len)
  %strlt.memcmp35 = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %strlt.min.len32) #5
  %strlt.memcmp.eq36 = icmp eq i32 %strlt.memcmp35, 0
  %strlt.len.lt37 = icmp uge i32 %a.len, %s.len
  %strlt.memcmp.lt38 = icmp sge i32 %strlt.memcmp35, 0
  %strlt.result39 = select i1 %strlt.memcmp.eq36, i1 %strlt.len.lt37, i1 %strlt.memcmp.lt38
  store i1 %0, ptr %complit, align 1
  store i1 %streq.not, ptr %1, align 1
  store i1 %strlt.result, ptr %2, align 1
  store i1 %strlt.result19, ptr %3, align 1
  store i1 %strlt.result29, ptr %4, align 1
  store i1 %strlt.result39, ptr %5, align 1
  %6 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt41 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack42 = load i1, ptr %.elt41, align 1
  %7 = insertvalue [6 x i1] %6, i1 %.unpack42, 1
  %.elt43 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack44 = load i1, ptr %.elt43, align 1
  %8 = insertvalue [6 x i1] %7, i1 %.unpack44, 2
  %.elt45 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack46 = load i1, ptr %.elt45, align 1
  %9 = insertvalue [6 x i1] %8, i1 %.unpack46, 3
  %.elt47 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack48 = load i1, ptr %.elt47, align 1
  %10 = insertvalue [6 x i1] %9, i1 %.unpack48, 4
  %.elt49 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack50 = load i1, ptr %.elt49, align 1
  %11 = insertvalue [6 x i1] %10, i1 %.unpack50, 5
  ret [6 x i1] %11
}

; Function Attrs: nounwind
define hidden [6 x i1] @main.mixedNamedStringByte(ptr readonly %s.data, i32 %s.len, ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 1 dereferenceable(6) ptr @runtime.alloc(i32 6, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %s.len, %a.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %s.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq3 = icmp eq i32 %s.len, %a.len
  br i1 %streq.len.eq3, label %streq.body4, label %streq.next5

streq.body4:                                      ; preds = %streq.next
  %streq.memcmp8 = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %s.len) #5
  %streq.memcmp.eq9 = icmp ne i32 %streq.memcmp8, 0
  br label %streq.next5

streq.next5:                                      ; preds = %streq.body4, %streq.next
  %streq.not = phi i1 [ true, %streq.next ], [ %streq.memcmp.eq9, %streq.body4 ]
  %1 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %2 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %s.len, i32 %a.len)
  %strlt.memcmp = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %s.len, %a.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %3 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len12 = call i32 @llvm.umin.i32(i32 %a.len, i32 %s.len)
  %strlt.memcmp15 = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %strlt.min.len12) #5
  %strlt.memcmp.eq16 = icmp eq i32 %strlt.memcmp15, 0
  %strlt.len.lt17 = icmp uge i32 %a.len, %s.len
  %strlt.memcmp.lt18 = icmp sge i32 %strlt.memcmp15, 0
  %strlt.result19 = select i1 %strlt.memcmp.eq16, i1 %strlt.len.lt17, i1 %strlt.memcmp.lt18
  %4 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len22 = call i32 @llvm.umin.i32(i32 %a.len, i32 %s.len)
  %strlt.memcmp25 = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %strlt.min.len22) #5
  %strlt.memcmp.eq26 = icmp eq i32 %strlt.memcmp25, 0
  %strlt.len.lt27 = icmp ult i32 %a.len, %s.len
  %strlt.memcmp.lt28 = icmp slt i32 %strlt.memcmp25, 0
  %strlt.result29 = select i1 %strlt.memcmp.eq26, i1 %strlt.len.lt27, i1 %strlt.memcmp.lt28
  %5 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len32 = call i32 @llvm.umin.i32(i32 %s.len, i32 %a.len)
  %strlt.memcmp35 = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %strlt.min.len32) #5
  %strlt.memcmp.eq36 = icmp eq i32 %strlt.memcmp35, 0
  %strlt.len.lt37 = icmp uge i32 %s.len, %a.len
  %strlt.memcmp.lt38 = icmp sge i32 %strlt.memcmp35, 0
  %strlt.result39 = select i1 %strlt.memcmp.eq36, i1 %strlt.len.lt37, i1 %strlt.memcmp.lt38
  store i1 %0, ptr %complit, align 1
  store i1 %streq.not, ptr %1, align 1
  store i1 %strlt.result, ptr %2, align 1
  store i1 %strlt.result19, ptr %3, align 1
  store i1 %strlt.result29, ptr %4, align 1
  store i1 %strlt.result39, ptr %5, align 1
  %6 = insertvalue [6 x i1] poison, i1 %0, 0
  %.elt41 = getelementptr inbounds nuw i8, ptr %complit, i32 1
  %.unpack42 = load i1, ptr %.elt41, align 1
  %7 = insertvalue [6 x i1] %6, i1 %.unpack42, 1
  %.elt43 = getelementptr inbounds nuw i8, ptr %complit, i32 2
  %.unpack44 = load i1, ptr %.elt43, align 1
  %8 = insertvalue [6 x i1] %7, i1 %.unpack44, 2
  %.elt45 = getelementptr inbounds nuw i8, ptr %complit, i32 3
  %.unpack46 = load i1, ptr %.elt45, align 1
  %9 = insertvalue [6 x i1] %8, i1 %.unpack46, 3
  %.elt47 = getelementptr inbounds nuw i8, ptr %complit, i32 4
  %.unpack48 = load i1, ptr %.elt47, align 1
  %10 = insertvalue [6 x i1] %9, i1 %.unpack48, 4
  %.elt49 = getelementptr inbounds nuw i8, ptr %complit, i32 5
  %.unpack50 = load i1, ptr %.elt49, align 1
  %11 = insertvalue [6 x i1] %10, i1 %.unpack50, 5
  ret [6 x i1] %11
}

; Function Attrs: nounwind
define hidden i1 @main.mixedLocalByteString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %0 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.mixedLocalStringByte(ptr readonly %s.data, i32 %s.len, ptr %a.data, i32 %a.len, i32 %a.cap, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %s.len, i32 %a.len)
  %strlt.memcmp = call i32 @memcmp(ptr %s.data, ptr %a.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %s.len, %a.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden i1 @main.mixedSliceByteString(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.highmax = icmp ult i32 %a.cap, 2
  br i1 %slice.highmax, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %s.len, 2
  br i1 %streq.len.eq, label %streq.body, label %streq.next

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef

streq.body:                                       ; preds = %slice.next
  %streq.memcmp = call i32 @memcmp(ptr noundef nonnull dereferenceable(2) %a.data, ptr noundef nonnull dereferenceable(2) %s.data, i32 2) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %slice.next
  %0 = phi i1 [ false, %slice.next ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.mixedPrecomputedStringSlice(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %slice.maxcap = icmp ult i32 %s.len, 2
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, 2
  br i1 %streq.len.eq, label %streq.body, label %streq.next

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef

streq.body:                                       ; preds = %slice.next
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %slice.next
  %0 = phi i1 [ false, %slice.next ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden { i1, i1 } @main.mixedByteStringReuse(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %s.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %2 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  %strlt.lhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %strlt.lhs.len, i32 %s.len)
  %strlt.lhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr %strlt.lhs.ptr, ptr %s.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %strlt.lhs.len, %s.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  %3 = insertvalue { i1, i1 } zeroinitializer, i1 %2, 0
  %4 = insertvalue { i1, i1 } %3, i1 %strlt.result, 1
  ret { i1, i1 } %4
}

; Function Attrs: nounwind
define hidden { i1, %runtime._string } @main.mixedByteStringEscape(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %s.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %2 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  %3 = insertvalue { i1, %runtime._string } zeroinitializer, i1 %2, 0
  %4 = insertvalue { i1, %runtime._string } %3, %runtime._string %0, 1
  ret { i1, %runtime._string } %4
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringStore(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr dereferenceable_or_null(8) %dst, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %s.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %2 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  %3 = icmp eq ptr %dst, null
  br i1 %3, label %store.throw, label %store.next

store.next:                                       ; preds = %streq.next
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %dst, align 4
  %dst.repack1 = getelementptr inbounds nuw i8, ptr %dst, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %dst.repack1, align 4
  ret i1 %2

store.throw:                                      ; preds = %streq.next
  call void @runtime.nilPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %store.throw
  ret i1 undef
}

; Function Attrs: nounwind
define hidden { i1, %runtime._interface } @main.mixedByteStringBox(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %s.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %2 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  %3 = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %3, ptr nonnull %stackalloc, ptr undef) #5
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %3, align 4
  %.repack1 = getelementptr inbounds nuw i8, ptr %3, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %.repack1, align 4
  %4 = insertvalue %runtime._interface { ptr @"reflect/types.type:basic:string", ptr undef }, ptr %3, 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %3, ptr nonnull %stackalloc, ptr undef) #5
  %5 = insertvalue { i1, %runtime._interface } zeroinitializer, i1 %2, 0
  %6 = insertvalue { i1, %runtime._interface } %5, %runtime._interface %4, 1
  ret { i1, %runtime._interface } %6
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringMutation(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %entry
  store i8 122, ptr %b.data, align 1
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

lookup.throw:                                     ; preds = %entry
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef

streq.body:                                       ; preds = %lookup.next
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %s.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %lookup.next
  %3 = phi i1 [ false, %lookup.next ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %3
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringCall(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = call { ptr, i32, i32 } @main.mutateBytes(ptr %b.data, i32 %b.len, i32 %b.cap, ptr undef)
  %3 = extractvalue { ptr, i32, i32 } %2, 0
  call void @runtime.trackPointer(ptr %3, ptr nonnull %stackalloc, ptr undef) #5
  %strlt.rhs.len = extractvalue %runtime._string %0, 1
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %s.len, i32 %strlt.rhs.len)
  %strlt.rhs.ptr = extractvalue %runtime._string %0, 0
  %strlt.memcmp = call i32 @memcmp(ptr %s.data, ptr %strlt.rhs.ptr, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %s.len, %strlt.rhs.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringAcrossBlock(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %b.data, i32 %b.len, i32 %b.cap, ptr readonly %s.data, i32 %s.len, i1 %mutate, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  br i1 %mutate, label %if.then, label %if.done

if.then:                                          ; preds = %entry
  %2 = icmp eq i32 %b.len, 0
  br i1 %2, label %lookup.throw, label %lookup.next

lookup.next:                                      ; preds = %if.then
  store i8 122, ptr %b.data, align 1
  br label %if.done

if.done:                                          ; preds = %lookup.next, %entry
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

lookup.throw:                                     ; preds = %if.then
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i1 undef

streq.body:                                       ; preds = %if.done
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %s.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %if.done
  %3 = phi i1 [ false, %if.done ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %3
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringLoad(ptr %a.data, i32 %a.len, i32 %a.cap, ptr dereferenceable_or_null(8) %s, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %0 = icmp eq ptr %s, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load ptr, ptr %s, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %s, i32 4
  %.unpack2 = load i32, ptr %.elt1, align 4
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %a.len, %.unpack2
  br i1 %streq.len.eq, label %streq.body, label %streq.next

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef

streq.body:                                       ; preds = %deref.next
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %.unpack, i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %deref.next
  %1 = phi i1 [ false, %deref.next ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %1
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
  call void @runtime.trackPointer(ptr %.unpack, ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %streq.len.eq = icmp eq i32 %.unpack2, %a.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %deref.throw
  ret i1 undef

streq.body:                                       ; preds = %deref.next
  %streq.memcmp = call i32 @memcmp(ptr %.unpack, ptr %a.data, i32 %.unpack2) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %deref.next
  %1 = phi i1 [ false, %deref.next ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %1
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringSlice(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %a.data, ptr nonnull %stackalloc, ptr undef) #5
  %slice.maxcap = icmp ult i32 %s.len, 2
  br i1 %slice.maxcap, label %slice.throw, label %slice.next

slice.next:                                       ; preds = %entry
  %streq.len.eq = icmp eq i32 %a.len, 2
  br i1 %streq.len.eq, label %streq.body, label %streq.next

slice.throw:                                      ; preds = %entry
  call void @runtime.slicePanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %slice.throw
  ret i1 undef

streq.body:                                       ; preds = %slice.next
  %streq.memcmp = call i32 @memcmp(ptr %a.data, ptr %s.data, i32 %a.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %slice.next
  %0 = phi i1 [ false, %slice.next ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringChangeType(ptr %a.data, i32 %a.len, i32 %a.cap, ptr readonly %s.data, i32 %s.len, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %streq.lhs.len = extractvalue %runtime._string %0, 1
  %streq.len.eq = icmp eq i32 %streq.lhs.len, %s.len
  br i1 %streq.len.eq, label %streq.body, label %streq.next

streq.body:                                       ; preds = %entry
  %streq.lhs.ptr = extractvalue %runtime._string %0, 0
  %streq.memcmp = call i32 @memcmp(ptr %streq.lhs.ptr, ptr %s.data, i32 %streq.lhs.len) #5
  %streq.memcmp.eq = icmp eq i32 %streq.memcmp, 0
  br label %streq.next

streq.next:                                       ; preds = %streq.body, %entry
  %2 = phi i1 [ false, %entry ], [ %streq.memcmp.eq, %streq.body ]
  ret i1 %2
}

; Function Attrs: nounwind
define hidden i1 @main.mixedByteStringInterface(ptr %a.data, i32 %a.len, i32 %a.cap, ptr %s.typecode, ptr %s.value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @runtime.stringFromBytes(ptr %a.data, i32 %a.len, i32 %a.cap, ptr undef) #5
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #5
  %2 = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 69 to ptr), ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %2, ptr nonnull %stackalloc, ptr undef) #5
  %.elt = extractvalue %runtime._string %0, 0
  store ptr %.elt, ptr %2, align 4
  %.repack1 = getelementptr inbounds nuw i8, ptr %2, i32 4
  %.elt2 = extractvalue %runtime._string %0, 1
  store i32 %.elt2, ptr %.repack1, align 4
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %stackalloc, ptr undef) #5
  call void @runtime.trackPointer(ptr nonnull %2, ptr nonnull %stackalloc, ptr undef) #5
  %3 = call i1 @runtime.interfaceEqual(ptr nonnull @"reflect/types.type:basic:string", ptr nonnull %2, ptr %s.typecode, ptr %s.value, ptr undef) #5
  ret i1 %3
}

declare i1 @runtime.interfaceEqual(ptr, ptr, ptr, ptr, ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.stringCompareLarger(ptr readonly %s1.data, i32 %s1.len, ptr readonly %s2.data, i32 %s2.len, ptr %context) unnamed_addr #1 {
entry:
  %strlt.min.len = call i32 @llvm.umin.i32(i32 %s2.len, i32 %s1.len)
  %strlt.memcmp = call i32 @memcmp(ptr %s2.data, ptr %s1.data, i32 %strlt.min.len) #5
  %strlt.memcmp.eq = icmp eq i32 %strlt.memcmp, 0
  %strlt.len.lt = icmp ult i32 %s2.len, %s1.len
  %strlt.memcmp.lt = icmp slt i32 %strlt.memcmp, 0
  %strlt.result = select i1 %strlt.memcmp.eq, i1 %strlt.len.lt, i1 %strlt.memcmp.lt
  ret i1 %strlt.result
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
  call void @runtime.lookupPanic(ptr undef) #5
  br label %unwind.return

unwind.return:                                    ; preds = %lookup.throw
  ret i8 undef
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { noinline nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #4 = { nocallback nofree nounwind speculatable willreturn memory(none) }
attributes #5 = { nounwind }
