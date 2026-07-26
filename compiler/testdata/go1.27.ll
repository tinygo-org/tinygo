; ModuleID = 'go1.27.go'
source_filename = "go1.27.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.structField = type { ptr, ptr }
%runtime._interface = type { ptr, ptr }

@"reflect/types.signature:Regular:func:{basic:int}{basic:int}" = linkonce_odr constant i8 0, align 1
@"reflect/types.type:named:main.genericMethod" = linkonce_odr constant { ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [19 x i8] } { ptr @"named:main.genericMethod$methodset", i8 122, i16 -32767, ptr getelementptr ({ ptr, i8, i16, ptr, { i32, [1 x ptr] } }, ptr @"reflect/types.type:pointer:named:main.genericMethod", i32 0, i32 1), ptr @"reflect/types.type:struct:{}", ptr @"reflect/types.type.pkgpath:main", { i32, [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/types.signature:Regular:func:{basic:int}{basic:int}"] }, [19 x i8] c"main.genericMethod\00" }, align 4
@"reflect/types.type.pkgpath:main" = linkonce_odr unnamed_addr constant [5 x i8] c"main\00", align 1
@"reflect/types.type:pointer:named:main.genericMethod" = linkonce_odr constant { ptr, i8, i16, ptr, { i32, [1 x ptr] } } { ptr @"pointer:named:main.genericMethod$methodset", i8 -43, i16 -32767, ptr getelementptr ({ ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [19 x i8] }, ptr @"reflect/types.type:named:main.genericMethod", i32 0, i32 1), { i32, [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/types.signature:Regular:func:{basic:int}{basic:int}"] } }, align 4
@"reflect/methods.Regular:func:{basic:int}{basic:int}" = linkonce_odr constant i8 0, align 1
@"main$string" = internal unnamed_addr constant [18 x i8] c"main.genericMethod", align 1
@"main$string.1" = internal unnamed_addr constant [7 x i8] c"Regular", align 1
@"pointer:named:main.genericMethod$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"reflect/methods.Regular:func:{basic:int}{basic:int}"], { ptr } { ptr @"(*main.genericMethod).Regular" } }
@"reflect/types.type:struct:{}" = linkonce_odr constant { i8, i16, ptr, ptr, i32, i16, [0 x %runtime.structField] } { i8 90, i16 0, ptr @"reflect/types.type:pointer:struct:{}", ptr @"reflect/types.type.pkgpath.empty", i32 0, i16 0, [0 x %runtime.structField] zeroinitializer }, align 4
@"reflect/types.type.pkgpath.empty" = linkonce_odr unnamed_addr constant [1 x i8] zeroinitializer, align 1
@"reflect/types.type:pointer:struct:{}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:struct:{}" }, align 4
@"named:main.genericMethod$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"reflect/methods.Regular:func:{basic:int}{basic:int}"], { ptr } { ptr @"(main.genericMethod).Regular$invoke" } }
@"reflect/types.type:named:main.onlyGenericMethod" = linkonce_odr constant { i8, i16, ptr, ptr, ptr, [23 x i8] } { i8 122, i16 0, ptr @"reflect/types.type:pointer:named:main.onlyGenericMethod", ptr @"reflect/types.type:struct:{}", ptr @"reflect/types.type.pkgpath:main", [23 x i8] c"main.onlyGenericMethod\00" }, align 4
@"reflect/types.type:pointer:named:main.onlyGenericMethod" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:named:main.onlyGenericMethod" }, align 4

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @"(main.genericMethod).Regular"(i32 %n, ptr %context) unnamed_addr #1 {
entry:
  ret i32 %n
}

; Function Attrs: nounwind
define hidden { %runtime._interface, %runtime._interface } @main.useGenericMethods(ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr nonnull getelementptr inbounds nuw (i8, ptr @"reflect/types.type:named:main.genericMethod", i32 4), ptr nonnull %stackalloc, ptr undef) #2
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:named:main.onlyGenericMethod", ptr nonnull %stackalloc, ptr undef) #2
  call void @runtime.trackPointer(ptr null, ptr nonnull %stackalloc, ptr undef) #2
  ret { %runtime._interface, %runtime._interface } { %runtime._interface { ptr getelementptr ({ ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [19 x i8] }, ptr @"reflect/types.type:named:main.genericMethod", i32 0, i32 1), ptr null }, %runtime._interface { ptr @"reflect/types.type:named:main.onlyGenericMethod", ptr null } }
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(*main.genericMethod).Regular"(ptr %t, i32 %n, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %t, ptr nonnull %stackalloc, ptr undef) #2
  %0 = call i32 @"(main.genericMethod).Regular"(i32 %n, ptr undef)
  ret i32 %0
}

; Function Attrs: nounwind
define linkonce_odr i32 @"(main.genericMethod).Regular$invoke"(ptr %0, i32 %1, ptr %2) unnamed_addr #1 {
entry:
  %ret = call i32 @"(main.genericMethod).Regular"(i32 %1, ptr %2)
  ret i32 %ret
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind }
