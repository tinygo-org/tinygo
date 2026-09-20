@"internal/reflectlite.makeFuncTypes" = global ptr null
@"internal/reflectlite.makeFuncAdapters" = global ptr null
@"internal/reflectlite.makeFuncLinksLen" = global i32 0
@"reflect/types.type:named:unusedMakeFuncType" = internal constant { i8 } zeroinitializer
@"reflect/types.type:func:{}{}" = constant { [2 x i32], i8, i8, i16, ptr, [0 x ptr] } { [2 x i32] [i32 ptrtoint (ptr @"reflect/call:func:{}{}" to i32), i32 ptrtoint (ptr @"reflect/makefunc:func:{}{}" to i32)], i8 24, i8 0, i16 0, ptr null, [0 x ptr] zeroinitializer }

define weak_odr void @"reflect/call:func:{}{}"() {
entry:
  ret void
}

define internal void @unused() #0 {
entry:
  ret void
}

define weak_odr void @"reflect/makefunc:func:{}{}"() {
entry:
  call void @"internal/reflectlite.makeFuncCall"()
  ret void
}

define internal void @"internal/reflectlite.makeFuncCall"() {
entry:
  %value = load i8, ptr @"reflect/types.type:named:unusedMakeFuncType"
  ret void
}

attributes #0 = { "tinygo-reflect-makefunc"="" }
