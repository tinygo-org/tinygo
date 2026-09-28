target datalayout = "e-m:e-p:32:32-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv7m-none-eabi"

%inner = type { ptr, i32 }
%outer = type { i32, %inner }
%deep = type { i32, { [2 x %inner], i32 } }
%mixed = type { ptr, { i32, i32 }, [2 x i32], [0 x ptr] }

@sink = global %inner zeroinitializer
@pointerSink = global ptr null

declare ptr @runtime.alloc(i32, ptr)
declare void @capture(%inner)
declare void @nocapture(ptr nocapture)

define %outer @wrap(ptr %p) {
  %inner = insertvalue %inner { ptr null, i32 7 }, ptr %p, 0
  %outer = insertvalue %outer { i32 3, %inner undef }, %inner %inner, 1
  ret %outer %outer
}

define %deep @wrapDeep(ptr %p) {
  %inner = insertvalue %inner { ptr null, i32 7 }, ptr %p, 0
  %array = insertvalue [2 x %inner] zeroinitializer, %inner %inner, 1
  %middle = insertvalue { [2 x %inner], i32 } zeroinitializer, [2 x %inner] %array, 0
  %outer = insertvalue %deep zeroinitializer, { [2 x %inner], i32 } %middle, 1
  ret %deep %outer
}

define { i32, [2 x [2 x ptr]] } @wrapArrays(ptr %p) {
  %inner = insertvalue [2 x ptr] zeroinitializer, ptr %p, 1
  %array = insertvalue [2 x [2 x ptr]] zeroinitializer, [2 x ptr] %inner, 1
  %outer = insertvalue { i32, [2 x [2 x ptr]] } zeroinitializer, [2 x [2 x ptr]] %array, 1
  ret { i32, [2 x [2 x ptr]] } %outer
}

define [2 x %inner] @wrapSingleArray(ptr %p) {
  %inner = insertvalue %inner zeroinitializer, ptr %p, 0
  %array = insertvalue [2 x %inner] zeroinitializer, %inner %inner, 1
  ret [2 x %inner] %array
}

define %inner @forward(ptr %p) {
  %outer = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %outer, 1
  ret %inner %inner
}

define void @captureInner(%inner %value) {
  store %inner %value, ptr @sink
  ret void
}

define %outer @recursive(ptr %p, i1 %again) {
entry:
  br i1 %again, label %recurse, label %end
recurse:
  %result = call %outer @recursive(ptr %p, i1 false)
  ret %outer %result
end:
  %wrapped = call %outer @wrap(ptr %p)
  ret %outer %wrapped
}

define void @captureSecond(%inner %ignored, %inner %captured) {
  store %inner %captured, ptr @sink
  ret void
}

define { [2 x %inner], i32 } @nestedStructReturn() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %deep @wrapDeep(ptr %p)
  %middle = extractvalue %deep %result, 1
  ret { [2 x %inner], i32 } %middle
}

define [2 x ptr] @nestedArrayReturn() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call { i32, [2 x [2 x ptr]] } @wrapArrays(ptr %p)
  %array = extractvalue { i32, [2 x [2 x ptr]] } %result, 1
  %inner = extractvalue [2 x [2 x ptr]] %array, 1
  ret [2 x ptr] %inner
}

define %inner @arrayOfStructReturn() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %deep @wrapDeep(ptr %p)
  %middle = extractvalue %deep %result, 1
  %array = extractvalue { [2 x %inner], i32 } %middle, 0
  %inner = extractvalue [2 x %inner] %array, 1
  ret %inner %inner
}

define [2 x %inner] @structOfArrayReturn() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %deep @wrapDeep(ptr %p)
  %array = extractvalue %deep %result, 1, 0
  ret [2 x %inner] %array
}

define %inner @singleArrayReturn() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %array = call [2 x %inner] @wrapSingleArray(ptr %p)
  %inner = extractvalue [2 x %inner] %array, 1
  ret %inner %inner
}

define void @aggregateStore() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  store %inner %inner, ptr @sink
  ret void
}

define void @pointerStore() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  %ptr = extractvalue %inner %inner, 0
  store ptr %ptr, ptr @pointerSink
  ret void
}

define void @aggregateCall() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  call void @captureInner(%inner %inner)
  ret void
}

define void @unknownCall() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  call void @capture(%inner %inner)
  ret void
}

define void @indirectCall(ptr %fn) {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  call void %fn(%inner %inner)
  ret void
}

define %outer @repackReturn() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  %new = insertvalue %outer zeroinitializer, %inner %inner, 1
  ret %outer %new
}

define %inner @forwardReturn() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %inner = call %inner @forward(ptr %p)
  ret %inner %inner
}

define i32 @forwardLoad() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %inner = call %inner @forward(ptr %p)
  %ptr = extractvalue %inner %inner, 0
  %value = load i32, ptr %ptr
  ret i32 %value
}

define ptr @gepReturn() {
  %p = call ptr @runtime.alloc(i32 8, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  %ptr = extractvalue %inner %inner, 0
  %element = getelementptr i32, ptr %ptr, i32 1
  ret ptr %element
}

define i32 @gepLoad() {
  %p = call ptr @runtime.alloc(i32 8, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  %ptr = extractvalue %inner %inner, 0
  %element = getelementptr i32, ptr %ptr, i32 1
  store i32 42, ptr %element
  %value = load i32, ptr %element
  ret i32 %value
}

define %inner @aggregatePhi(i1 %cond) {
entry:
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  br i1 %cond, label %yes, label %no
yes:
  br label %end
no:
  br label %end
end:
  %value = phi %inner [ %inner, %yes ], [ zeroinitializer, %no ]
  ret %inner %value
}

define %inner @aggregateSelect(i1 %cond) {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  %value = select i1 %cond, %inner %inner, %inner zeroinitializer
  ret %inner %value
}

define %inner @recursiveReturn() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @recursive(ptr %p, i1 true)
  %inner = extractvalue %outer %result, 1
  ret %inner %inner
}

define void @duplicateArguments() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  call void @captureSecond(%inner %inner, %inner %inner)
  ret void
}

define i32 @nonEscapingScalar() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %deep @wrapDeep(ptr %p)
  %middle = extractvalue %deep %result, 1
  %array = extractvalue { [2 x %inner], i32 } %middle, 0
  %inner = extractvalue [2 x %inner] %array, 1
  %value = extractvalue %inner %inner, 1
  ret i32 %value
}

define i32 @nonEscapingLoad() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  %ptr = extractvalue %inner %inner, 0
  %value = load i32, ptr %ptr
  ret i32 %value
}

define i1 @nonEscapingNilCheck() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  %ptr = extractvalue %inner %inner, 0
  %value = icmp eq ptr %ptr, null
  ret i1 %value
}

define void @nonEscapingDiscard() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  ret void
}

define void @nonEscapingCall() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %outer @wrap(ptr %p)
  %inner = extractvalue %outer %result, 1
  %ptr = extractvalue %inner %inner, 0
  call void @nocapture(ptr %ptr)
  ret void
}

define %mixed @wrapMixed(ptr %p) {
  %result = insertvalue %mixed zeroinitializer, ptr %p, 0
  ret %mixed %result
}

define { i32, i32 } @pointerFreeStruct() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %mixed @wrapMixed(ptr %p)
  %scalar = extractvalue %mixed %result, 1
  ret { i32, i32 } %scalar
}

define [2 x i32] @pointerFreeArray() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %mixed @wrapMixed(ptr %p)
  %scalar = extractvalue %mixed %result, 2
  ret [2 x i32] %scalar
}

define [0 x ptr] @emptyPointerArray() {
  %p = call ptr @runtime.alloc(i32 4, ptr null)
  %result = call %mixed @wrapMixed(ptr %p)
  %empty = extractvalue %mixed %result, 3
  ret [0 x ptr] %empty
}
