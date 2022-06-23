target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = global i1 0
@main.v2 = global i1 0
@"reflect/types.type:named:main.foo" = private constant { i8, i8*, i8* } { i8 34, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:pointer:named:main.foo", i32 0, i32 0), i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:basic:int", i32 0, i32 0) }, align 4
@"reflect/types.type:pointer:named:main.foo" = external constant { i8, i8* }
@"reflect/types.typeid:named:main.foo" = external constant i8
@"reflect/types.type:basic:int" = private constant { i8, i8* } { i8 2, i8* getelementptr inbounds ({ i8, i8* }, { i8, i8* }* @"reflect/types.type:pointer:basic:int", i32 0, i32 0) }, align 4
@"reflect/types.type:pointer:basic:int" = external constant { i8, i8* }


declare i1 @runtime.typeAssert(i8*, i8*, i8*, i8*)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
entry:
  ; Test type asserts.
  %typecode = call i1 @runtime.typeAssert(i8* getelementptr inbounds ({ i8, i8*, i8* }, { i8, i8*, i8* }* @"reflect/types.type:named:main.foo", i32 0, i32 0), i8* @"reflect/types.typeid:named:main.foo", i8* undef, i8* null)
  store i1 %typecode, i1* @main.v1
  %typecode2 = call i1 @runtime.typeAssert(i8* null, i8* @"reflect/types.typeid:named:main.foo", i8* undef, i8* null)
  store i1 %typecode2, i1* @main.v2
  ret void
}
