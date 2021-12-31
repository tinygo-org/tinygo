; ModuleID = 'interface.go'
source_filename = "interface.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.typecodeID = type { %runtime.typecodeID*, i32, %runtime.interfaceMethodInfo*, %runtime.typecodeID*, i32 }
%runtime.interfaceMethodInfo = type { i8*, i32 }
%runtime._interface = type { i32, i8* }
%runtime._string = type { i8*, i32 }

@"reflect/types.type:basic:int" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* null, i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* @"reflect/types.type:pointer:basic:int", i32 0 }
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:int", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null, i32 0 }
@"reflect/types.type:pointer:named:error" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:named:error", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null, i32 0 }
@"reflect/types.type:named:error" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* @"reflect/types.type:pointer:named:error", i32 ptrtoint (i1 (i32)* @"interface:{Error:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* bitcast ([1 x i8*]* @"reflect/types.interface:interface{Error() string}$interface" to %runtime.typecodeID*), i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* @"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}", i32 ptrtoint (i1 (i32)* @"interface:{Error:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/methods.Error() string" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{Error() string}$interface" = linkonce_odr constant [1 x i8*] [i8* @"reflect/methods.Error() string"]
@"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null, i32 0 }
@"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{String:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null, i32 0 }
@"reflect/types.type:interface:{String:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* bitcast ([1 x i8*]* @"reflect/types.interface:interface{String() string}$interface" to %runtime.typecodeID*), i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}", i32 ptrtoint (i1 (i32)* @"interface:{String:func:{}{basic:string}}.$typeassert" to i32) }
@"reflect/methods.String() string" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{String() string}$interface" = linkonce_odr constant [1 x i8*] [i8* @"reflect/methods.String() string"]
@"reflect/types.typeid:basic:int" = external constant i8

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

declare void @runtime.trackPointer(i8* nocapture readonly, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.simpleType(i8* %context) unnamed_addr #0 {
entry:
  call void @runtime.trackPointer(i8* null, i8* undef) #0
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:basic:int" to i32), i8* null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.pointerType(i8* %context) unnamed_addr #0 {
entry:
  call void @runtime.trackPointer(i8* null, i8* undef) #0
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:pointer:basic:int" to i32), i8* null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.interfaceType(i8* %context) unnamed_addr #0 {
entry:
  call void @runtime.trackPointer(i8* null, i8* undef) #0
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:pointer:named:error" to i32), i8* null }
}

declare i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(i32) #1

; Function Attrs: nounwind
define hidden %runtime._interface @main.anonymousInterfaceType(i8* %context) unnamed_addr #0 {
entry:
  call void @runtime.trackPointer(i8* null, i8* undef) #0
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" to i32), i8* null }
}

declare i1 @"interface:{String:func:{}{basic:string}}.$typeassert"(i32) #2

; Function Attrs: nounwind
define hidden i1 @main.isInt(i32 %itf.typecode, i8* %itf.value, i8* %context) unnamed_addr #0 {
entry:
  %typecode = call i1 @runtime.typeAssert(i32 %itf.typecode, i8* nonnull @"reflect/types.typeid:basic:int", i8* undef) #0
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %typecode
}

declare i1 @runtime.typeAssert(i32, i8* dereferenceable_or_null(1), i8*)

; Function Attrs: nounwind
define hidden i1 @main.isError(i32 %itf.typecode, i8* %itf.value, i8* %context) unnamed_addr #0 {
entry:
  %0 = call i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(i32 %itf.typecode) #0
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i1 @main.isStringer(i32 %itf.typecode, i8* %itf.value, i8* %context) unnamed_addr #0 {
entry:
  %0 = call i1 @"interface:{String:func:{}{basic:string}}.$typeassert"(i32 %itf.typecode) #0
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0
}

; Function Attrs: nounwind
define hidden i8 @main.callFooMethod(i32 %itf.typecode, i8* %itf.value, i8* %context) unnamed_addr #0 {
entry:
  %0 = call i8 @"interface:{String:func:{}{basic:string},main.foo:func:{basic:int}{basic:uint8}}.foo$invoke"(i8* %itf.value, i32 3, i32 %itf.typecode, i8* undef) #0
  ret i8 %0
}

declare i8 @"interface:{String:func:{}{basic:string},main.foo:func:{basic:int}{basic:uint8}}.foo$invoke"(i8*, i32, i32, i8*) #3

; Function Attrs: nounwind
define hidden %runtime._string @main.callErrorMethod(i32 %itf.typecode, i8* %itf.value, i8* %context) unnamed_addr #0 {
entry:
  %0 = call %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(i8* %itf.value, i32 %itf.typecode, i8* undef) #0
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(i8* %1, i8* undef) #0
  ret %runtime._string %0
}

declare %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(i8*, i32, i8*) #4

attributes #0 = { nounwind }
attributes #1 = { "tinygo-methods"="reflect/methods.Error() string" }
attributes #2 = { "tinygo-methods"="reflect/methods.String() string" }
attributes #3 = { "tinygo-invoke"="main.$methods.foo(int) byte" "tinygo-methods"="reflect/methods.String() string; main.$methods.foo(int) byte" }
attributes #4 = { "tinygo-invoke"="reflect/methods.Error() string" "tinygo-methods"="reflect/methods.Error() string" }
