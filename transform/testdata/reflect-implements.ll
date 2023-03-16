target datalayout = "e-m:e-p:32:32-p270:32:32-p271:32:32-p272:64:64-f64:32:64-f80:32-n8:16:32-S128"
target triple = "i686--linux"

%runtime._interface = type { ptr, ptr }

@"reflect/types.type:named:error" = internal constant { i8, ptr, ptr } { i8 52, ptr @"reflect/types.type:pointer:named:error", ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}" }, align 4
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = internal constant { i8, ptr } { i8 20, ptr @"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" }, align 4
@"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" = internal constant { i8, ptr } { i8 21, ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}" }, align 4
@"reflect/types.type:pointer:named:error" = internal constant { i8, ptr } { i8 21, ptr @"reflect/types.type:named:error" }, align 4
@"reflect/types.type:pointer:named:reflect.rawType" = internal constant { ptr, i8, ptr } { ptr null, i8 21, ptr null }, align 4
@"reflect/methods.Implements(reflect.Type) bool" = internal constant i8 0, align 1

; var errorType = reflect.TypeOf((*error)(nil)).Elem()
; func isError(typ reflect.Type) bool {
;   return typ.Implements(errorType)
; }
; The type itself is stored in %typ.value, %typ.typecode just refers to the
; type of reflect.Type. This function can be optimized because errorType is
; known at compile time (after the interp pass has run).
define i1 @main.isError(ptr %typ.typecode, ptr %typ.value, ptr %context) {
entry:
  %result = call i1 @"reflect.Type.Implements$invoke"(ptr %typ.value, ptr getelementptr inbounds ({ ptr, i8, ptr }, ptr @"reflect/types.type:pointer:named:reflect.rawType", i32 0, i32 1), ptr @"reflect/types.type:named:error", ptr %typ.typecode, ptr undef)
  ret i1 %result
}

; This Implements method call can not be optimized because itf is not known at
; compile time.
; func isUnknown(typ, itf reflect.Type) bool {
;   return typ.Implements(itf)
; }
define i1 @main.isUnknown(ptr %typ.typecode, ptr %typ.value, ptr %itf.typecode, ptr %itf.value, ptr %context) {
entry:
  %result = call i1 @"reflect.Type.Implements$invoke"(ptr %typ.value, ptr %itf.typecode, ptr %itf.value, ptr %typ.typecode, ptr undef)
  ret i1 %result
}

declare i1 @"reflect.Type.Implements$invoke"(ptr, ptr, ptr, ptr, ptr) #0
declare i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(ptr %0) #1

attributes #0 = { "tinygo-invoke"="reflect/methods.Implements(reflect.Type) bool" "tinygo-methods"="reflect/methods.Align() int; reflect/methods.Implements(reflect.Type) bool" }
attributes #1 = { "tinygo-methods"="reflect/methods.Error() string" }
