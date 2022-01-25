target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

%runtime.typecodeID = type { %runtime.typecodeID*, i32, %runtime.interfaceMethodInfo*, %runtime.typecodeID*, i32 }
%runtime.interfaceMethodInfo = type { i8*, i32 }

@"reflect/types.type:named:error" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null, i32 ptrtoint (i1 (i32)* @"error.$typeassert" to i32) }
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* bitcast ([1 x i8*]* @"reflect/types.interface:interface{Error() string}$interface" to %runtime.typecodeID*), i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null, i32 ptrtoint (i1 (i32)* @"error.$typeassert" to i32) }
@"reflect/methods.Error() string" = linkonce_odr constant i8 0
@"reflect/types.interface:interface{Error() string}$interface" = linkonce_odr constant [1 x i8*] [i8* @"reflect/methods.Error() string"]
@"reflect/methods.Align() int" = linkonce_odr constant i8 0
@"reflect/methods.Implements(reflect.Type) bool" = linkonce_odr constant i8 0
@"reflect.Type$interface" = linkonce_odr constant [2 x i8*] [i8* @"reflect/methods.Align() int", i8* @"reflect/methods.Implements(reflect.Type) bool"]
@"reflect/types.type:named:reflect.rawType" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:uintptr", i32 0, %runtime.interfaceMethodInfo* getelementptr inbounds ([20 x %runtime.interfaceMethodInfo], [20 x %runtime.interfaceMethodInfo]* @"reflect.rawType$methodset", i32 0, i32 0), %runtime.typecodeID* null, i32 0 }
@"reflect.rawType$methodset" = linkonce_odr constant [20 x %runtime.interfaceMethodInfo] zeroinitializer
@"reflect/types.type:basic:uintptr" = linkonce_odr constant %runtime.typecodeID zeroinitializer

define i1 @main.isError(i32 %typ.typecode, i8* %typ.value, i8* %context) {
entry:
  %0 = ptrtoint i8* %typ.value to i32
  %1 = call i1 @"error.$typeassert"(i32 %0)
  ret i1 %1
}

define i1 @main.isUnknown(i32 %typ.typecode, i8* %typ.value, i32 %itf.typecode, i8* %itf.value, i8* %context) {
entry:
  %result = call i1 @"reflect.Type.Implements$invoke"(i8* %typ.value, i32 %itf.typecode, i8* %itf.value, i32 %typ.typecode, i8* undef)
  ret i1 %result
}

declare i1 @"reflect.Type.Implements$invoke"(i8*, i32, i8*, i32, i8*) #0

declare i1 @"error.$typeassert"(i32) #1

attributes #0 = { "tinygo-invoke"="reflect/methods.Implements(reflect.Type) bool" "tinygo-methods"="reflect/methods.Align() int; reflect/methods.Implements(reflect.Type) bool" }
attributes #1 = { "tinygo-methods"="reflect/methods.Error() string" }
