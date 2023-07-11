target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.v1 = global i1 0
@main.v2 = global i1 0
@"reflect/types.type:named:main.foo" = private constant { i8, ptr, ptr } { i8 34, ptr @"reflect/types.type:pointer:named:main.foo", ptr @"reflect/types.type:basic:int" }, align 4
@"reflect/types.type:pointer:named:main.foo" = external constant { i8, ptr }
@"reflect/types.typeid:named:main.foo" = external constant i8
@"reflect/types.type:basic:int" = private constant { i8, ptr } { i8 2, ptr @"reflect/types.type:pointer:basic:int" }, align 4
@"reflect/types.type:pointer:basic:int" = external constant { i8, ptr }


declare i1 @runtime.typeAssert(ptr, ptr, ptr, ptr)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
entry:
  ; Test type asserts.
  %typecode = call i1 @runtime.typeAssert(ptr @"reflect/types.type:named:main.foo", ptr @"reflect/types.typeid:named:main.foo", ptr undef, ptr null)
  store i1 %typecode, ptr @main.v1
  %typecode2 = call i1 @runtime.typeAssert(ptr null, ptr @"reflect/types.typeid:named:main.foo", ptr undef, ptr null)
  store i1 %typecode2, ptr @main.v2
  ret void
}
