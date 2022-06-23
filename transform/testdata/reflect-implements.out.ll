target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

@"reflect/types.type:named:error" = internal constant { i8, i8*, i8* } { i8 52, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:pointer:named:error", i32 0, i32 0), i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, i32 0) }, align 4
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = internal constant { i8, i8* } { i8 20, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}", i32 0, i32 0) }, align 4
@"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" = internal constant { i8, i8* } { i8 21, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:interface:{Error:func:{}{basic:string}}", i32 0, i32 0) }, align 4
@"reflect/types.type:pointer:named:error" = internal constant { i8, i8* } { i8 21, i8* getelementptr inbounds ({ i8, i8*, i8* }, { i8, i8*, i8* }* @"reflect/types.type:named:error", i32 0, i32 0) }, align 4
@"reflect/types.type:pointer:named:reflect.rawType" = internal constant { i8*, i8, i8* } { i8* null, i8 21, i8* null }, align 4
@"reflect/methods.Implements(reflect.Type) bool" = internal constant i8 0, align 1

define i1 @main.isError(i8* %typ.typecode, i8* %typ.value, i8* %context) {
entry:
  %0 = call i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(i8* %typ.value)
  ret i1 %0
}

define i1 @main.isUnknown(i8* %typ.typecode, i8* %typ.value, i8* %itf.typecode, i8* %itf.value, i8* %context) {
entry:
  %result = call i1 @"reflect.Type.Implements$invoke"(i8* %typ.value, i8* %itf.typecode, i8* %itf.value, i8* %typ.typecode, i8* undef)
  ret i1 %result
}

declare i1 @"reflect.Type.Implements$invoke"(i8*, i8*, i8*, i8*, i8*) #0

declare i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(i8*) #1

attributes #0 = { "tinygo-invoke"="reflect/methods.Implements(reflect.Type) bool" "tinygo-methods"="reflect/methods.Align() int; reflect/methods.Implements(reflect.Type) bool" }
attributes #1 = { "tinygo-methods"="reflect/methods.Error() string" }
