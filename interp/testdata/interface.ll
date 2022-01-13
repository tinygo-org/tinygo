target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

%runtime.typecodeID = type { %runtime.typecodeID*, i64, %runtime.interfaceMethodInfo* }
%runtime.interfaceMethodInfo = type { i8*, i64 }

@main.v1 = global i1 0
@main.v2 = global i1 0
@"reflect/types.type:named:main.foo" = private constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:int", i64 0, %runtime.interfaceMethodInfo* null }
@"reflect/types.typeid:named:main.foo" = external constant i8
@"reflect/types.type:basic:int" = external constant %runtime.typecodeID


declare i1 @runtime.typeAssert(i64, i8*, i8*, i8*)

define void @runtime.initAll() unnamed_addr {
entry:
  call void @main.init()
  ret void
}

define internal void @main.init() unnamed_addr {
entry:
  ; Test type asserts.
  %typecode = call i1 @runtime.typeAssert(i64 ptrtoint (%runtime.typecodeID* @"reflect/types.type:named:main.foo" to i64), i8* @"reflect/types.typeid:named:main.foo", i8* undef, i8* null)
  store i1 %typecode, i1* @main.v1
  %typecode2 = call i1 @runtime.typeAssert(i64 0, i8* @"reflect/types.typeid:named:main.foo", i8* undef, i8* null)
  store i1 %typecode2, i1* @main.v2
  ret void
}
