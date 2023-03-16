target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

@"reflect/types.type:named:error" = internal constant { i8, ptr, ptr } { i8 52, ptr @"reflect/types.type:pointer:named:error", ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}" }, align 4
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = internal constant { i8, ptr } { i8 20, ptr @"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" }, align 4
@"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" = internal constant { i8, ptr } { i8 21, ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}" }, align 4
@"reflect/types.type:pointer:named:error" = internal constant { i8, ptr } { i8 21, ptr @"reflect/types.type:named:error" }, align 4
@"reflect/types.type:pointer:named:reflect.rawType" = internal constant { ptr, i8, ptr } { ptr null, i8 21, ptr null }, align 4
@"reflect/methods.Implements(reflect.Type) bool" = internal constant i8 0, align 1

define i1 @main.isError(ptr %typ.typecode, ptr %typ.value, ptr %context) {
entry:
  %0 = call i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(ptr %typ.value)
  ret i1 %0
}

define i1 @main.isUnknown(ptr %typ.typecode, ptr %typ.value, ptr %itf.typecode, ptr %itf.value, ptr %context) {
entry:
  %result = call i1 @"reflect.Type.Implements$invoke"(ptr %typ.value, ptr %itf.typecode, ptr %itf.value, ptr %typ.typecode, ptr undef)
  ret i1 %result
}

declare i1 @"reflect.Type.Implements$invoke"(ptr, ptr, ptr, ptr, ptr) #0

declare i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(ptr) #1

attributes #0 = { "tinygo-invoke"="reflect/methods.Implements(reflect.Type) bool" "tinygo-methods"="reflect/methods.Align() int; reflect/methods.Implements(reflect.Type) bool" }
attributes #1 = { "tinygo-methods"="reflect/methods.Error() string" }
