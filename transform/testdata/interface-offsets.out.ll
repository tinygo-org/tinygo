target datalayout = "e-p:32:32-i64:64"

@external = external global i8
@symbolic = global ptr getelementptr (i8, ptr @"reflect/types.type:named:Symbolic", i32 ptrtoint (ptr @external to i32))
@"reflect/types.type:named:Symbolic" = internal constant { ptr, i8, [11 x i8] } { ptr null, i8 34, [11 x i8] zeroinitializer }, align 4

define i1 @narrow() {
  ret i1 true
}

define i1 @typed() {
  ret i1 true
}
