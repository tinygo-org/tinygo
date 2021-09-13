; ModuleID = 'interface.go'
source_filename = "interface.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-wasi"

%runtime.typecodeID = type { %runtime.typecodeID*, i32, %runtime.interfaceMethodInfo*, %runtime.typecodeID* }
%runtime.interfaceMethodInfo = type { i8*, i32 }
%runtime._interface = type { i32, i8* }
%runtime._string = type { i8*, i32 }

@"reflect/types.type:basic:int" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* null, i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* @"reflect/types.type:pointer:basic:int" }
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:int", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null }
@"reflect/types.type:pointer:named:error" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:named:error", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null }
@"reflect/types.type:named:error" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* @"reflect/types.type:pointer:named:error" }
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* bitcast ([1 x i8*]* @"reflect/types.interface:interface{Error() string}$interface" to %runtime.typecodeID*), i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* @"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" }
@"reflect/methods.Error() string" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{Error() string}$interface" = linkonce_odr constant [1 x i8*] [i8* @"reflect/methods.Error() string"]
@"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null }
@"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{String:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null }
@"reflect/types.type:interface:{String:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* bitcast ([1 x i8*]* @"reflect/types.interface:interface{String() string}$interface" to %runtime.typecodeID*), i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" }
@"reflect/methods.String() string" = linkonce_odr constant i8 0, align 1
@"reflect/types.interface:interface{String() string}$interface" = linkonce_odr constant [1 x i8*] [i8* @"reflect/methods.String() string"]
@"reflect/types.typeid:basic:int" = external constant i8
@"error$interface" = linkonce_odr constant [1 x i8*] [i8* @"reflect/methods.Error() string"]

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.simpleType(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:basic:int" to i32), i8* null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.pointerType(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:pointer:basic:int" to i32), i8* null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.interfaceType(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:pointer:named:error" to i32), i8* null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.anonymousInterfaceType(i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" to i32), i8* null }
}

; Function Attrs: nounwind
define hidden i1 @main.isInt(i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %typecode = call i1 @runtime.typeAssert(i32 %itf.typecode, i8* nonnull @"reflect/types.typeid:basic:int", i8* undef, i8* null) #0
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %typecode
}

declare i1 @runtime.typeAssert(i32, i8* dereferenceable_or_null(1), i8*, i8*)

; Function Attrs: nounwind
define hidden i1 @main.isError(i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = call i1 @runtime.interfaceImplements(i32 %itf.typecode, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"error$interface", i32 0, i32 0), i8* undef, i8* null) #0
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0
}

declare i1 @runtime.interfaceImplements(i32, i8** dereferenceable_or_null(4), i8*, i8*)

; Function Attrs: nounwind
define hidden i1 @main.isStringer(i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %0 = call i1 @runtime.interfaceImplements(i32 %itf.typecode, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"reflect/types.interface:interface{String() string}$interface", i32 0, i32 0), i8* undef, i8* null) #0
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0
}

; Function Attrs: nounwind
define hidden %runtime._string @main.callErrorMethod(i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) unnamed_addr #0 {
entry:
  %invoke.func = call i32 @runtime.interfaceMethod(i32 %itf.typecode, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"error$interface", i32 0, i32 0), i8* nonnull @"reflect/methods.Error() string", i8* undef, i8* null) #0
  %invoke.func.cast = inttoptr i32 %invoke.func to %runtime._string (i8*, i8*, i8*)*
  %0 = call %runtime._string %invoke.func.cast(i8* %itf.value, i8* undef, i8* undef) #0
  ret %runtime._string %0
}

declare i32 @runtime.interfaceMethod(i32, i8** dereferenceable_or_null(4), i8* dereferenceable_or_null(1), i8*, i8*)

attributes #0 = { nounwind }
