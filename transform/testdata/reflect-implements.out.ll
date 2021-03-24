target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

%runtime.typecodeID = type { %runtime.typecodeID*, i32, %runtime.interfaceMethodInfo* }
%runtime.interfaceMethodInfo = type { i8*, i32 }

@"reflect/types.type:named:error" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null }
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* bitcast ([1 x i8*]* @"reflect/types.interface:interface{Error() string}$interface" to %runtime.typecodeID*), i32 0, %runtime.interfaceMethodInfo* null }
@"func Error() string" = external constant i8
@"reflect/types.interface:interface{Error() string}$interface" = linkonce_odr constant [1 x i8*] [i8* @"func Error() string"]
@"func Align() int" = external constant i8
@"func Implements(reflect.Type) bool" = external constant i8
@"reflect.Type$interface" = linkonce_odr constant [2 x i8*] [i8* @"func Align() int", i8* @"func Implements(reflect.Type) bool"]
@"reflect/types.type:named:reflect.rawType" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:uintptr", i32 0, %runtime.interfaceMethodInfo* getelementptr inbounds ([20 x %runtime.interfaceMethodInfo], [20 x %runtime.interfaceMethodInfo]* @"reflect.rawType$methodset", i32 0, i32 0) }
@"reflect.rawType$methodset" = linkonce_odr constant [20 x %runtime.interfaceMethodInfo] zeroinitializer
@"reflect/types.type:basic:uintptr" = linkonce_odr constant %runtime.typecodeID zeroinitializer

declare i1 @runtime.interfaceImplements(i32, i8**, i8*, i8*)

declare i32 @runtime.interfaceMethod(i32, i8**, i8*, i8*, i8*)

define i1 @main.isError(i32 %typ.typecode, i8* %typ.value, i8* %context, i8* %parentHandle) {
entry:
  %0 = ptrtoint i8* %typ.value to i32
  %1 = call i1 @runtime.interfaceImplements(i32 %0, i8** getelementptr inbounds ([1 x i8*], [1 x i8*]* @"reflect/types.interface:interface{Error() string}$interface", i32 0, i32 0), i8* undef, i8* undef)
  ret i1 %1
}

define i1 @main.isUnknown(i32 %typ.typecode, i8* %typ.value, i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) {
entry:
  %invoke.func = call i32 @runtime.interfaceMethod(i32 %typ.typecode, i8** getelementptr inbounds ([2 x i8*], [2 x i8*]* @"reflect.Type$interface", i32 0, i32 0), i8* nonnull @"func Implements(reflect.Type) bool", i8* undef, i8* null)
  %invoke.func.cast = inttoptr i32 %invoke.func to i1 (i8*, i32, i8*, i8*, i8*)*
  %result = call i1 %invoke.func.cast(i8* %typ.value, i32 %itf.typecode, i8* %itf.value, i8* undef, i8* undef)
  ret i1 %result
}
