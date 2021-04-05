; ModuleID = 'interface.go'
source_filename = "interface.go"
target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

%runtime.typecodeID = type { %runtime.typecodeID*, i32, %runtime.interfaceMethodInfo* }
%runtime.interfaceMethodInfo = type { i8*, i32 }
%runtime._interface = type { i32, i8* }
%runtime._string = type { i8*, i32 }

@"reflect/types.type:basic:int" = linkonce_odr constant %runtime.typecodeID zeroinitializer
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:int", i32 0, %runtime.interfaceMethodInfo* null }
@"reflect/types.type:pointer:named:error" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:named:error", i32 0, %runtime.interfaceMethodInfo* null }
@"reflect/types.type:named:error" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null }
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* bitcast ([1 x i8*]* @"reflect/types.interface:interface{Error() string}$interface" to %runtime.typecodeID*), i32 0, %runtime.interfaceMethodInfo* null }
@"func Error() string" = external constant i8
@"reflect/types.interface:interface{Error() string}$interface" = linkonce_odr constant [1 x i8*] [i8* @"func Error() string"]
@"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{String:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null }
@"reflect/types.type:interface:{String:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* bitcast ([1 x i8*]* @"reflect/types.interface:interface{String() string}$interface" to %runtime.typecodeID*), i32 0, %runtime.interfaceMethodInfo* null }
@"func String() string" = external constant i8
@"reflect/types.interface:interface{String() string}$interface" = linkonce_odr constant [1 x i8*] [i8* @"func String() string"]
@"reflect/types.typeid:basic:int" = external constant i8
@"error$interface" = linkonce_odr constant [1 x i8*] [i8* @"func Error() string"]

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

define hidden void @main.init(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret void
}

define hidden %runtime._interface @main.simpleType(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:basic:int" to i32), i8* null }
}

define hidden %runtime._interface @main.pointerType(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:pointer:basic:int" to i32), i8* null }
}

define hidden %runtime._interface @main.interfaceType(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:pointer:named:error" to i32), i8* null }
}

define hidden %runtime._interface @main.anonymousInterfaceType(i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  ret %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" to i32), i8* null }
}

define hidden i1 @main.isInt(i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %typecode = call i1 @runtime.typeAssert(i32 %itf.typecode, i8* nonnull @"reflect/types.typeid:basic:int", i8* undef, i8* null)
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %typecode
}

declare i1 @runtime.typeAssert(i32, i8* dereferenceable_or_null(1), i8*, i8*)

define hidden i1 @main.isError(i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = call i1 @runtime.interfaceImplements(i32 %itf.typecode, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"error$interface", i32 0, i32 0), i8* undef, i8* null)
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0
}

declare i1 @runtime.interfaceImplements(i32, i8** dereferenceable_or_null(4), i8*, i8*)

define hidden i1 @main.isStringer(i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %0 = call i1 @runtime.interfaceImplements(i32 %itf.typecode, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"reflect/types.interface:interface{String() string}$interface", i32 0, i32 0), i8* undef, i8* null)
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0
}

define hidden %runtime._string @main.callErrorMethod(i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) unnamed_addr {
entry:
  %invoke.func = call i32 @runtime.interfaceMethod(i32 %itf.typecode, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"error$interface", i32 0, i32 0), i8* nonnull @"func Error() string", i8* undef, i8* null)
  %invoke.func.cast = inttoptr i32 %invoke.func to %runtime._string (i8*, i8*, i8*)*
  %0 = call %runtime._string %invoke.func.cast(i8* %itf.value, i8* undef, i8* undef)
  ret %runtime._string %0
}

declare i32 @runtime.interfaceMethod(i32, i8** dereferenceable_or_null(4), i8* dereferenceable_or_null(1), i8*, i8*)
