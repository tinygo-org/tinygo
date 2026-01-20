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
@"reflect/types.type:pointer:named:main.Foo$local:11:0:" = internal constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:named:main.Foo$local:11:0:" }, align 4
@"reflect/types.type:named:main.Foo$local:11:0:" = internal constant { i8, i16, ptr, ptr, ptr, [9 x i8] } { i8 -6, i16 0, ptr @"reflect/types.type:pointer:named:main.Foo$local:11:0:", ptr @"reflect/types.type:struct:{A:basic:int}", ptr @"reflect/types.type.pkgpath:main", [9 x i8] c"main.Foo\00" }, align 4
@"reflect/types.type.pkgpath:main" = linkonce_odr unnamed_addr constant [5 x i8] c"main\00", align 1
@"reflect/types.type:struct:{A:basic:int}" = linkonce_odr constant { i8, i16, ptr, ptr, i32, i16, [1 x %runtime.structField] } { i8 -38, i16 0, ptr @"reflect/types.type:pointer:struct:{A:basic:int}", ptr @"reflect/types.type.pkgpath:main", i32 4, i16 1, [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{A:basic:int}.A" }] }, align 4
@"reflect/types.type:pointer:struct:{A:basic:int}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:struct:{A:basic:int}" }, align 4
@"reflect/types.type:struct:{A:basic:int}.A" = internal unnamed_addr constant [4 x i8] c"\04\00A\00", align 1
@"reflect/types.typeid:pointer:named:main.Foo$local:11:0:" = external constant i8
@"reflect/types.type:pointer:named:main.Foo$local:12:0:" = internal constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:named:main.Foo$local:12:0:" }, align 4
@"reflect/types.type:named:main.Foo$local:12:0:" = internal constant { i8, i16, ptr, ptr, ptr, [9 x i8] } { i8 -6, i16 0, ptr @"reflect/types.type:pointer:named:main.Foo$local:12:0:", ptr @"reflect/types.type:struct:{A:pointer:basic:int}", ptr @"reflect/types.type.pkgpath:main", [9 x i8] c"main.Foo\00" }, align 4
@"reflect/types.type:struct:{A:pointer:basic:int}" = linkonce_odr constant { i8, i16, ptr, ptr, i32, i16, [1 x %runtime.structField] } { i8 -38, i16 0, ptr @"reflect/types.type:pointer:struct:{A:pointer:basic:int}", ptr @"reflect/types.type.pkgpath:main", i32 4, i16 1, [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:basic:int", ptr @"reflect/types.type:struct:{A:pointer:basic:int}.A" }] }, align 4
@"reflect/types.type:pointer:struct:{A:pointer:basic:int}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:struct:{A:pointer:basic:int}" }, align 4
@"reflect/types.type:struct:{A:pointer:basic:int}.A" = internal unnamed_addr constant [4 x i8] c"\04\00A\00", align 1
@"reflect/types.typeid:pointer:named:main.Foo$local:12:0:" = external constant i8
@"reflect/types.type:pointer:named:main.Foo$local:0:0:12:0:" = internal constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:named:main.Foo$local:0:0:12:0:" }, align 4
@"reflect/types.type:named:main.Foo$local:0:0:12:0:" = internal constant { i8, i16, ptr, ptr, ptr, [9 x i8] } { i8 -6, i16 0, ptr @"reflect/types.type:pointer:named:main.Foo$local:0:0:12:0:", ptr @"reflect/types.type:struct:{A:pointer:basic:uint8}", ptr @"reflect/types.type.pkgpath:main", [9 x i8] c"main.Foo\00" }, align 4
@"reflect/types.type:struct:{A:pointer:basic:uint8}" = linkonce_odr constant { i8, i16, ptr, ptr, i32, i16, [1 x %runtime.structField] } { i8 -38, i16 0, ptr @"reflect/types.type:pointer:struct:{A:pointer:basic:uint8}", ptr @"reflect/types.type.pkgpath:main", i32 4, i16 1, [1 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:pointer:basic:uint8", ptr @"reflect/types.type:struct:{A:pointer:basic:uint8}.A" }] }, align 4
@"reflect/types.type:pointer:struct:{A:pointer:basic:uint8}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:struct:{A:pointer:basic:uint8}" }, align 4
@"reflect/types.type:struct:{A:pointer:basic:uint8}.A" = internal unnamed_addr constant [4 x i8] c"\04\00A\00", align 1
@"reflect/types.type:pointer:basic:uint8" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:uint8" }, align 4
@"reflect/types.type:basic:uint8" = linkonce_odr constant { i8, ptr } { i8 -56, ptr @"reflect/types.type:pointer:basic:uint8" }, align 4
@"reflect/types.typeid:pointer:named:main.Foo$local:0:0:12:0:" = external constant i8

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.simpleType(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:int", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._interface { ptr @"reflect/types.type:basic:int", ptr null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.pointerType(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:pointer:basic:int", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._interface { ptr @"reflect/types.type:pointer:basic:int", ptr null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.interfaceType(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:pointer:named:error", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._interface { ptr @"reflect/types.type:pointer:named:error", ptr null }
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.anonymousInterfaceType(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._interface { ptr @"reflect/types.type:pointer:interface:{String:func:{}{basic:string}}", ptr null }
}

; Function Attrs: nounwind
define hidden i1 @main.isInt(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #2 {
entry:
  %typecode = call i1 @runtime.typeAssert(ptr %itf.typecode, ptr nonnull @"reflect/types.typeid:basic:int", ptr undef) #7
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  ret i1 %typecode

typeassert.ok:                                    ; preds = %entry
  br label %typeassert.next
}

declare i1 @runtime.typeAssert(ptr, ptr dereferenceable_or_null(1), ptr) #1

; Function Attrs: nounwind
define hidden i1 @main.isError(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #2 {
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
define hidden i1 @main.isStringer(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #2 {
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
define hidden i8 @main.callFooMethod(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i8 @"interface:{String:func:{}{basic:string},main.foo:func:{basic:int}{basic:uint8}}.foo$invoke"(ptr %itf.value, i32 3, ptr %itf.typecode, ptr undef) #7
  ret i8 %0
}

declare i8 @"interface:{String:func:{}{basic:string},main.foo:func:{basic:int}{basic:uint8}}.foo$invoke"(ptr, i32, ptr, ptr) #5

; Function Attrs: nounwind
define hidden %runtime._string @main.callErrorMethod(ptr %itf.typecode, ptr %itf.value, ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(ptr %itf.value, ptr %itf.typecode, ptr undef) #7
  %1 = extractvalue %runtime._string %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #7
  ret %runtime._string %0
}

declare %runtime._string @"interface:{Error:func:{}{basic:string}}.Error$invoke"(ptr, ptr, ptr) #6

; Function Attrs: nounwind
define hidden void @main.namedFoo(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 4 dereferenceable(4) ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #7
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:pointer:named:main.Foo$local:11:0:", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #7
  %0 = call %runtime._interface @main.copyOf(ptr nonnull @"reflect/types.type:pointer:named:main.Foo$local:11:0:", ptr nonnull %complit, ptr undef)
  %1 = extractvalue %runtime._interface %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #7
  %2 = extractvalue %runtime._interface %0, 1
  call void @runtime.trackPointer(ptr %2, ptr nonnull %stackalloc, ptr undef) #7
  %interface.type = extractvalue %runtime._interface %0, 0
  %typecode = call i1 @runtime.typeAssert(ptr %interface.type, ptr nonnull @"reflect/types.typeid:pointer:named:main.Foo$local:11:0:", ptr undef) #7
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  %typeassert.value = phi ptr [ null, %entry ], [ %typeassert.value.ptr, %typeassert.ok ]
  call void @runtime.interfaceTypeAssert(i1 %typecode, ptr undef) #7
  %3 = icmp eq ptr %typeassert.value, null
  br i1 %3, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %typeassert.next
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %4 = load i32, ptr %typeassert.value, align 4
  call void @runtime.printlock(ptr undef) #7
  call void @runtime.printint32(i32 %4, ptr undef) #7
  call void @runtime.printnl(ptr undef) #7
  call void @runtime.printunlock(ptr undef) #7
  ret void

typeassert.ok:                                    ; preds = %entry
  %typeassert.value.ptr = extractvalue %runtime._interface %0, 1
  br label %typeassert.next

gep.throw:                                        ; preds = %typeassert.next
  call void @runtime.nilPanic(ptr undef) #7
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.copyOf(ptr %src.typecode, ptr %src.value, ptr %context) unnamed_addr #2 {
entry:
  %0 = insertvalue %runtime._interface zeroinitializer, ptr %src.typecode, 0
  %1 = insertvalue %runtime._interface %0, ptr %src.value, 1
  ret %runtime._interface %1
}

declare void @runtime.interfaceTypeAssert(i1, ptr) #1

declare void @runtime.nilPanic(ptr) #1

declare void @runtime.printlock(ptr) #1

declare void @runtime.printint32(i32, ptr) #1

declare void @runtime.printnl(ptr) #1

declare void @runtime.printunlock(ptr) #1

; Function Attrs: nounwind
define hidden void @main.namedFoo2Nested(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %complit = call align 4 dereferenceable(4) ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #7
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:pointer:named:main.Foo$local:12:0:", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr nonnull %complit, ptr nonnull %stackalloc, ptr undef) #7
  %0 = call %runtime._interface @main.copyOf(ptr nonnull @"reflect/types.type:pointer:named:main.Foo$local:12:0:", ptr nonnull %complit, ptr undef)
  %1 = extractvalue %runtime._interface %0, 0
  call void @runtime.trackPointer(ptr %1, ptr nonnull %stackalloc, ptr undef) #7
  %2 = extractvalue %runtime._interface %0, 1
  call void @runtime.trackPointer(ptr %2, ptr nonnull %stackalloc, ptr undef) #7
  %interface.type = extractvalue %runtime._interface %0, 0
  %typecode = call i1 @runtime.typeAssert(ptr %interface.type, ptr nonnull @"reflect/types.typeid:pointer:named:main.Foo$local:12:0:", ptr undef) #7
  br i1 %typecode, label %typeassert.ok, label %typeassert.next

typeassert.next:                                  ; preds = %typeassert.ok, %entry
  %typeassert.value = phi ptr [ null, %entry ], [ %typeassert.value.ptr, %typeassert.ok ]
  call void @runtime.interfaceTypeAssert(i1 %typecode, ptr undef) #7
  %3 = icmp eq ptr %typeassert.value, null
  br i1 %3, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %typeassert.next
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %gep.next
  %4 = load ptr, ptr %typeassert.value, align 4
  call void @runtime.trackPointer(ptr %4, ptr nonnull %stackalloc, ptr undef) #7
  %5 = icmp eq ptr %4, null
  call void @runtime.printlock(ptr undef) #7
  call void @runtime.printbool(i1 %5, ptr undef) #7
  call void @runtime.printnl(ptr undef) #7
  call void @runtime.printunlock(ptr undef) #7
  br i1 false, label %gep.throw1, label %gep.next2

gep.next2:                                        ; preds = %deref.next
  br i1 false, label %deref.throw3, label %deref.next4

deref.next4:                                      ; preds = %gep.next2
  %6 = load ptr, ptr %typeassert.value, align 4
  call void @runtime.trackPointer(ptr %6, ptr nonnull %stackalloc, ptr undef) #7
  %7 = icmp eq ptr %6, null
  br i1 %7, label %if.then, label %if.done

typeassert.ok:                                    ; preds = %entry
  %typeassert.value.ptr = extractvalue %runtime._interface %0, 1
  br label %typeassert.next

if.then:                                          ; preds = %deref.next4
  %complit5 = call align 4 dereferenceable(4) ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #7
  call void @runtime.trackPointer(ptr nonnull %complit5, ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:pointer:named:main.Foo$local:0:0:12:0:", ptr nonnull %stackalloc, ptr undef) #7
  call void @runtime.trackPointer(ptr nonnull %complit5, ptr nonnull %stackalloc, ptr undef) #7
  %8 = call %runtime._interface @main.copyOf(ptr nonnull @"reflect/types.type:pointer:named:main.Foo$local:0:0:12:0:", ptr nonnull %complit5, ptr undef)
  %9 = extractvalue %runtime._interface %8, 0
  call void @runtime.trackPointer(ptr %9, ptr nonnull %stackalloc, ptr undef) #7
  %10 = extractvalue %runtime._interface %8, 1
  call void @runtime.trackPointer(ptr %10, ptr nonnull %stackalloc, ptr undef) #7
  %interface.type6 = extractvalue %runtime._interface %8, 0
  %typecode7 = call i1 @runtime.typeAssert(ptr %interface.type6, ptr nonnull @"reflect/types.typeid:pointer:named:main.Foo$local:0:0:12:0:", ptr undef) #7
  br i1 %typecode7, label %typeassert.ok8, label %typeassert.next9

typeassert.next9:                                 ; preds = %typeassert.ok8, %if.then
  %typeassert.value11 = phi ptr [ null, %if.then ], [ %typeassert.value.ptr10, %typeassert.ok8 ]
  call void @runtime.interfaceTypeAssert(i1 %typecode7, ptr undef) #7
  %11 = icmp eq ptr %typeassert.value11, null
  br i1 %11, label %gep.throw12, label %gep.next13

gep.next13:                                       ; preds = %typeassert.next9
  br i1 false, label %deref.throw14, label %deref.next15

deref.next15:                                     ; preds = %gep.next13
  %12 = load ptr, ptr %typeassert.value11, align 4
  call void @runtime.trackPointer(ptr %12, ptr nonnull %stackalloc, ptr undef) #7
  %13 = icmp eq ptr %12, null
  call void @runtime.printlock(ptr undef) #7
  call void @runtime.printbool(i1 %13, ptr undef) #7
  call void @runtime.printnl(ptr undef) #7
  call void @runtime.printunlock(ptr undef) #7
  br label %if.done

typeassert.ok8:                                   ; preds = %if.then
  %typeassert.value.ptr10 = extractvalue %runtime._interface %8, 1
  br label %typeassert.next9

if.done:                                          ; preds = %deref.next15, %deref.next4
  ret void

gep.throw:                                        ; preds = %typeassert.next
  call void @runtime.nilPanic(ptr undef) #7
  unreachable

deref.throw:                                      ; preds = %gep.next
  unreachable

gep.throw1:                                       ; preds = %deref.next
  unreachable

deref.throw3:                                     ; preds = %gep.next2
  unreachable

gep.throw12:                                      ; preds = %typeassert.next9
  call void @runtime.nilPanic(ptr undef) #7
  unreachable

deref.throw14:                                    ; preds = %gep.next13
  unreachable
}

declare void @runtime.printbool(i1, ptr) #1

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-methods"="reflect/methods.Error() string" }
attributes #4 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-methods"="reflect/methods.String() string" }
attributes #5 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-invoke"="main.$methods.foo(int) uint8" "tinygo-methods"="reflect/methods.String() string; main.$methods.foo(int) uint8" }
attributes #6 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-invoke"="reflect/methods.Error() string" "tinygo-methods"="reflect/methods.Error() string" }
attributes #7 = { nounwind }
