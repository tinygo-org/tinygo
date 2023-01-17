; ModuleID = 'interface.go'
source_filename = "interface.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.typecodeID = type { ptr, i32, ptr, ptr, i32 }
%runtime._interface = type { i32, ptr }
%runtime._string = type { ptr, i32 }

@"reflect/types.type:basic:int" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:int", i32 0 }
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:int", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:named:error" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:named:error", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:named:error" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, ptr null, ptr @"reflect/types.type:pointer:named:error", i32 ptrtoint (ptr @"interface:{Error:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.interface:interface{Error() string}$interface", i32 0, ptr null, ptr @"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}", i32 ptrtoint (ptr @"interface:{Error:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/methods.Error() string" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{Error() string}$interface" = linkonce_odr constant [1 x ptr] [ptr @"reflect/methods.Error() string"]
@"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:interface:{String:func:{}{basic:string}}", i32 0, ptr null, ptr null, i32 0 }
@"reflect/types.type:interface:{String:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.interface:interface{String() string}$interface", i32 0, ptr null, ptr @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}", i32 ptrtoint (ptr @"interface:{String:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/methods.String() string" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{String() string}$interface" = linkonce_odr constant [1 x ptr] [ptr @"reflect/methods.String() string"]
@"reflect/types.typeid:basic:int" = external constant i8

declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.simpleType(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #6
  ret %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:basic:int" to i32), ptr null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.pointerType(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #6
  ret %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:pointer:basic:int" to i32), ptr null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.interfaceType(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #6
  ret %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:pointer:named:error" to i32), ptr null }
}

declare i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(i32) #2

; Function Attrs: nounwind
define hidden %runtime._interface @main.anonymousInterfaceType(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #6
  ret %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" to i32), ptr null }
}

declare i1 @"interface:{String:func:{}{basic:string}}.$typeassert"(i32) #3

; Function Attrs: nounwind
define hidden i1 @main.isInt(i32 %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %typecode = call i1 @runtime.typeAssert(i32 %itf.typecode, ptr nonnull @"reflect/types.typeid:basic:int", ptr undef) #6
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %typecode

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next
}

declare i1 @runtime.typeAssert(i32, ptr dereferenceable_or_null(1), ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.isError(i32 %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(i32 %itf.typecode) #6
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next
}

; Function Attrs: nounwind
define hidden i1 @main.isStringer(i32 %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @"interface:{String:func:{}{basic:string}}.$typeassert"(i32 %itf.typecode) #6
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next
}

; Function Attrs: nounwind
define hidden i8 @main.callFooMethod(i32 %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i8 @"interface:{String:func:{}{basic:string},main.foo:func:{basic:int}{basic:uint8}}.foo$invoke"(ptr %itf.value, i32 3, i32 %itf.typecode, ptr undef) #6
  ret i8 %0
}

declare i8 @"interface:{String:func:{}{basic:string},main.foo:func:{basic:int}{basic:uint8}}.foo$invoke"(ptr, i32, i32, ptr) #4

; Function Attrs: nounwind
define hidden %runtime._string @main.callErrorMethod(i32 %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(ptr %itf.value, i32 %itf.typecode, ptr undef) #6
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #6
  ret %runtime._string %0
}

declare %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(ptr, i32, ptr) #5

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-methods"="reflect/methods.Error() string" }
attributes #3 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-methods"="reflect/methods.String() string" }
attributes #4 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-invoke"="main.$methods.foo(int) uint8" "tinygo-methods"="reflect/methods.String() string; main.$methods.foo(int) uint8" }
attributes #5 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" "tinygo-invoke"="reflect/methods.Error() string" "tinygo-methods"="reflect/methods.Error() string" }
attributes #6 = { nounwind }
