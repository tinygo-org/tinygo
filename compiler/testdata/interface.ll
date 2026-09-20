; ModuleID = 'interface.go'
source_filename = "interface.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.structField = type { ptr, ptr }
%runtime._interface = type { ptr, ptr }
%runtime._string = type { ptr, i32 }

@"reflect/types.type:basic:int" = linkonce_odr constant { i8, ptr } { i8 -62, ptr @"reflect/types.type:pointer:basic:int" }, align 4
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:int" }, align 4
@"reflect/types.type:pointer:named:error" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:named:error" }, align 4
@"reflect/types.signature:Error:func:{}{basic:string}" = linkonce_odr constant i8 0, align 1
@"reflect/types.type:named:error" = linkonce_odr constant { i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [7 x i8] } { i8 116, i16 -32767, ptr @"reflect/types.type:pointer:named:error", ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}", ptr @"reflect/types.type.pkgpath.empty", { i32, [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/types.signature:Error:func:{}{basic:string}"] }, [7 x i8] c".error\00" }, align 4
@"reflect/types.type.pkgpath.empty" = linkonce_odr unnamed_addr constant [1 x i8] zeroinitializer, align 1
@"reflect/types.type:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant { i8, ptr, { i32, [1 x ptr] } } { i8 84, ptr @"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}", { i32, [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/types.signature:Error:func:{}{basic:string}"] } }, align 4
@"reflect/types.type:pointer:interface:{Error:func:{}{basic:string}}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:interface:{Error:func:{}{basic:string}}" }, align 4
@"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:interface:{String:func:{}{basic:string}}" }, align 4
@"reflect/types.signature:String:func:{}{basic:string}" = linkonce_odr constant i8 0, align 1
@"reflect/types.type:interface:{String:func:{}{basic:string}}" = linkonce_odr constant { i8, ptr, { i32, [1 x ptr] } } { i8 84, ptr @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}", { i32, [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/types.signature:String:func:{}{basic:string}"] } }, align 4
@"reflect/types.typeid:basic:int" = external constant i8
@"reflect/types.type:slice:named:main.assertionElem" = linkonce_odr constant { i8, i16, ptr, ptr } { i8 22, i16 0, ptr @"reflect/types.type:pointer:slice:named:main.assertionElem", ptr @"reflect/types.type:named:main.assertionElem" }, align 4
@"reflect/types.type:pointer:slice:named:main.assertionElem" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:slice:named:main.assertionElem" }, align 4
@"reflect/types.type:named:main.assertionElem" = linkonce_odr constant { i8, i16, ptr, ptr, ptr, [19 x i8] } { i8 -22, i16 0, ptr @"reflect/types.type:pointer:named:main.assertionElem", ptr @"reflect/types.type:basic:uint32", ptr @"reflect/types.type.pkgpath:main", [19 x i8] c"main.assertionElem\00" }, align 4
@"reflect/types.type.pkgpath:main" = linkonce_odr unnamed_addr constant [5 x i8] c"main\00", align 1
@"reflect/types.type:pointer:named:main.assertionElem" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:named:main.assertionElem" }, align 4
@"reflect/types.type:basic:uint32" = linkonce_odr constant { i8, ptr } { i8 -54, ptr @"reflect/types.type:pointer:basic:uint32" }, align 4
@"reflect/types.type:pointer:basic:uint32" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:uint32" }, align 4
@"reflect/types.type:array:3:named:main.assertionElem" = linkonce_odr constant { i8, i16, ptr, ptr, i32, ptr, ptr } { i8 -41, i16 0, ptr @"reflect/types.type:pointer:array:3:named:main.assertionElem", ptr @"reflect/types.type:named:main.assertionElem", i32 3, ptr @"reflect/types.type:slice:named:main.assertionElem", ptr inttoptr (i32 3 to ptr) }, align 4
@"reflect/types.type:pointer:array:3:named:main.assertionElem" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:array:3:named:main.assertionElem" }, align 4
@"reflect/types.type:map:{named:main.assertionElem,named:main.assertionElem}" = linkonce_odr constant { i8, i16, ptr, ptr, ptr, ptr } { i8 25, i16 0, ptr @"reflect/types.type:pointer:map:{named:main.assertionElem,named:main.assertionElem}", ptr @"reflect/types.type:named:main.assertionElem", ptr @"reflect/types.type:named:main.assertionElem", ptr @"runtime.hashmapType:uint32:uint32" }, align 4
@"reflect/types.type:pointer:map:{named:main.assertionElem,named:main.assertionElem}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:map:{named:main.assertionElem,named:main.assertionElem}" }, align 4
@"runtime.hashmapType:uint32:uint32" = linkonce_odr unnamed_addr constant { ptr, ptr, ptr } { ptr inttoptr (i32 3 to ptr), ptr inttoptr (i32 3 to ptr), ptr inttoptr (i32 297 to ptr) }
@"reflect/types.type:chan:sr:named:main.assertionElem" = linkonce_odr constant { i8, i16, ptr, ptr } { i8 83, i16 3, ptr @"reflect/types.type:pointer:chan:sr:named:main.assertionElem", ptr @"reflect/types.type:named:main.assertionElem" }, align 4
@"reflect/types.type:pointer:chan:sr:named:main.assertionElem" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:chan:sr:named:main.assertionElem" }, align 4
@"reflect/types.type:struct:{Value:named:main.assertionElem}" = linkonce_odr constant { i8, i16, ptr, ptr, i32, i16, ptr, [1 x %runtime.structField] } { i8 90, i16 0, ptr @"reflect/types.type:pointer:struct:{Value:named:main.assertionElem}", ptr @"reflect/types.type.pkgpath:main", i32 4, i16 1, ptr inttoptr (i32 3 to ptr), [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:named:main.assertionElem", ptr @"reflect/types.type:struct:{Value:named:main.assertionElem}.Value" }] }, align 4
@"reflect/types.type:pointer:struct:{Value:named:main.assertionElem}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:struct:{Value:named:main.assertionElem}" }, align 4
@"reflect/types.type:struct:{Value:named:main.assertionElem}.Value" = internal unnamed_addr constant [8 x i8] c"\04\00Value\00", align 1

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
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:int", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._interface { ptr @"reflect/types.type:basic:int", ptr null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.pointerType(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:pointer:basic:int", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._interface { ptr @"reflect/types.type:pointer:basic:int", ptr null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.interfaceType(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:pointer:named:error", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._interface { ptr @"reflect/types.type:pointer:named:error", ptr null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.anonymousInterfaceType(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._interface { ptr @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}", ptr null }
}

; Function Attrs: nounwind
define hidden i1 @main.isInt(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %typecode = call i1 @runtime.typeAssert(ptr %itf.typecode, ptr nonnull @"reflect/types.typeid:basic:int", ptr undef) #7
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %typecode

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next
}

declare i1 @runtime.typeAssert(ptr, ptr dereferenceable_or_null(1), ptr) #0

; Function Attrs: nounwind
define hidden i1 @main.assertionOnlyTypes(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %typecode = icmp eq ptr %itf.typecode, @"reflect/types.type:slice:named:main.assertionElem"
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  br i1 %typecode, label %typeswitch.body, label %typeswitch.next

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next

typeswitch.body:                                  ; preds = %typeassert.next33, %typeassert.next27, %typeassert.next21, %typeassert.next15, %typeassert.next9, %typeassert.next
  ret i1 true

typeswitch.next:                                  ; preds = %typeassert.next
  %typecode7 = icmp eq ptr %itf.typecode, @"reflect/types.type:array:3:named:main.assertionElem"
  br i1 %typecode7, label %typeassert.ok8, label %typeassert.next9

typeassert.next9:                                 ; preds = %typeassert.ok8, %typeswitch.next
  br i1 %typecode7, label %typeswitch.body, label %typeswitch.next1

typeassert.ok8:                                   ; preds = %typeswitch.next
  br label %typeassert.next9

typeswitch.next1:                                 ; preds = %typeassert.next9
  %typecode13 = icmp eq ptr %itf.typecode, @"reflect/types.type:pointer:array:3:named:main.assertionElem"
  br i1 %typecode13, label %typeassert.ok14, label %typeassert.next15

typeassert.next15:                                ; preds = %typeassert.ok14, %typeswitch.next1
  br i1 %typecode13, label %typeswitch.body, label %typeswitch.next2

typeassert.ok14:                                  ; preds = %typeswitch.next1
  br label %typeassert.next15

typeswitch.next2:                                 ; preds = %typeassert.next15
  %typecode19 = icmp eq ptr %itf.typecode, @"reflect/types.type:map:{named:main.assertionElem,named:main.assertionElem}"
  br i1 %typecode19, label %typeassert.ok20, label %typeassert.next21

typeassert.next21:                                ; preds = %typeassert.ok20, %typeswitch.next2
  br i1 %typecode19, label %typeswitch.body, label %typeswitch.next3

typeassert.ok20:                                  ; preds = %typeswitch.next2
  br label %typeassert.next21

typeswitch.next3:                                 ; preds = %typeassert.next21
  %typecode25 = icmp eq ptr %itf.typecode, @"reflect/types.type:chan:sr:named:main.assertionElem"
  br i1 %typecode25, label %typeassert.ok26, label %typeassert.next27

typeassert.next27:                                ; preds = %typeassert.ok26, %typeswitch.next3
  br i1 %typecode25, label %typeswitch.body, label %typeswitch.next4

typeassert.ok26:                                  ; preds = %typeswitch.next3
  br label %typeassert.next27

typeswitch.next4:                                 ; preds = %typeassert.next27
  %typecode31 = icmp eq ptr %itf.typecode, @"reflect/types.type:struct:{Value:named:main.assertionElem}"
  br i1 %typecode31, label %typeassert.ok32, label %typeassert.next33

typeassert.next33:                                ; preds = %typeassert.ok32, %typeswitch.next4
  br i1 %typecode31, label %typeswitch.body, label %typeswitch.next5

typeassert.ok32:                                  ; preds = %typeswitch.next4
  br label %typeassert.next33

typeswitch.next5:                                 ; preds = %typeassert.next33
  ret i1 false
}

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.start.p0(ptr nocapture) #2

; Function Attrs: nocallback nofree nosync nounwind willreturn memory(argmem: readwrite)
declare void @llvm.lifetime.end.p0(ptr nocapture) #2

; Function Attrs: nounwind
define hidden i1 @main.isError(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(ptr %itf.typecode) #7
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next
}

declare i1 @"interface:{Error:func:{}{basic:string}}.$typeassert"(ptr) #3

; Function Attrs: nounwind
define hidden i1 @main.isStringer(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i1 @"interface:{String:func:{}{basic:string}}.$typeassert"(ptr %itf.typecode) #7
  br i1 %0, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %0

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next
}

declare i1 @"interface:{String:func:{}{basic:string}}.$typeassert"(ptr) #4

; Function Attrs: nounwind
define hidden i8 @main.callFooMethod(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %0 = call i8 @"interface:{String:func:{}{basic:string},main.foo:func:{basic:int}{basic:uint8}}.foo$invoke"(ptr %itf.value, i32 3, ptr %itf.typecode, ptr undef) #7
  ret i8 %0
}

declare i8 @"interface:{String:func:{}{basic:string},main.foo:func:{basic:int}{basic:uint8}}.foo$invoke"(ptr, i32, ptr, ptr) #5

; Function Attrs: nounwind
define hidden %runtime._string @main.callErrorMethod(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(ptr %itf.value, ptr %itf.typecode, ptr undef) #7
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._string %0
}

declare %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(ptr, ptr, ptr) #6

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nocallback nofree nosync nounwind willreturn memory(argmem: readwrite) }
attributes #3 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-methods"="reflect/methods.Error:func:{}{basic:string}" }
attributes #4 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-methods"="reflect/methods.String:func:{}{basic:string}" }
attributes #5 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-invoke"="main.$methods.foo:func:{basic:int}{basic:uint8}" "tinygo-methods"="reflect/methods.String:func:{}{basic:string}; main.$methods.foo:func:{basic:int}{basic:uint8}" }
attributes #6 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-invoke"="reflect/methods.Error:func:{}{basic:string}" "tinygo-methods"="reflect/methods.Error:func:{}{basic:string}" }
attributes #7 = { nounwind }
