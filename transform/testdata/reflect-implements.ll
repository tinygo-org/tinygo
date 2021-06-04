target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

%runtime.typecodeID = type { %runtime.typecodeID*, i32, %runtime.interfaceMethodInfo* }
%runtime.interfaceMethodInfo = type { i8*, i32 }

@"reflect/types.type:named:error" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null }
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* bitcast ([1 x i8*]* @"reflect/types.interface:interface{Error() string}$interface" to %runtime.typecodeID*), i32 0, %runtime.interfaceMethodInfo* null }
@"reflect/methods.Error() string" = linkonce_odr constant i8 0
@"reflect/types.interface:interface{Error() string}$interface" = linkonce_odr constant [1 x i8*] [i8* @"reflect/methods.Error() string"]
@"reflect/methods.Align() int" = linkonce_odr constant i8 0
@"reflect/methods.Implements(reflect.Type) bool" = linkonce_odr constant i8 0
@"reflect.Type$interface" = linkonce_odr constant [2 x i8*] [i8* @"reflect/methods.Align() int", i8* @"reflect/methods.Implements(reflect.Type) bool"]
@"reflect/types.type:named:reflect.rawType" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:uintptr", i32 0, %runtime.interfaceMethodInfo* getelementptr inbounds ([20 x %runtime.interfaceMethodInfo], [20 x %runtime.interfaceMethodInfo]* @"reflect.rawType$methodset", i32 0, i32 0) }
@"reflect.rawType$methodset" = linkonce_odr constant [20 x %runtime.interfaceMethodInfo] zeroinitializer
@"reflect/types.type:basic:uintptr" = linkonce_odr constant %runtime.typecodeID zeroinitializer

declare i1 @runtime.interfaceImplements(i32, i8**, i8*, i8*)

declare i32 @runtime.interfaceMethod(i32, i8**, i8*, i8*, i8*)

; var errorType = reflect.TypeOf((*error)(nil)).Elem()
; func isError(typ reflect.Type) bool {
;   return typ.Implements(errorType)
; }
; The type itself is stored in %typ.value, %typ.typecode just refers to the
; type of reflect.Type. This function can be optimized because errorType is
; known at compile time (after the interp pass has run).
define i1 @main.isError(i32 %typ.typecode, i8* %typ.value, i8* %context, i8* %parentHandle) {
entry:
  %invoke.func = call i32 @runtime.interfaceMethod(i32 %typ.typecode, i8** getelementptr inbounds ([2 x i8*], [2 x i8*]* @"reflect.Type$interface", i32 0, i32 0), i8* nonnull @"reflect/methods.Implements(reflect.Type) bool", i8* undef, i8* null)
  %invoke.func.cast = inttoptr i32 %invoke.func to i1 (i8*, i32, i8*, i8*, i8*)*
  %result = call i1 %invoke.func.cast(i8* %typ.value, i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:named:reflect.rawType" to i32), i8* bitcast (%runtime.typecodeID* @"reflect/types.type:named:error" to i8*), i8* undef, i8* undef)
  ret i1 %result
}

; This Implements method call can not be optimized because itf is not known at
; compile time.
; func isUnknown(typ, itf reflect.Type) bool {
;   return typ.Implements(itf)
; }
define i1 @main.isUnknown(i32 %typ.typecode, i8* %typ.value, i32 %itf.typecode, i8* %itf.value, i8* %context, i8* %parentHandle) {
entry:
  %invoke.func = call i32 @runtime.interfaceMethod(i32 %typ.typecode, i8** getelementptr inbounds ([2 x i8*], [2 x i8*]* @"reflect.Type$interface", i32 0, i32 0), i8* nonnull @"reflect/methods.Implements(reflect.Type) bool", i8* undef, i8* null)
  %invoke.func.cast = inttoptr i32 %invoke.func to i1 (i8*, i32, i8*, i8*, i8*)*
  %result = call i1 %invoke.func.cast(i8* %typ.value, i32 %itf.typecode, i8* %itf.value, i8* undef, i8* undef)
  ret i1 %result
}
