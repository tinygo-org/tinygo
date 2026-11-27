target datalayout = "e-p:32:32-i64:64"

%descriptor = type { ptr, i8, [11 x i8] }
@"Number$methodset" = internal constant { i32, [0 x ptr], [0 x ptr] } zeroinitializer
@"reflect/types.type:named:Narrow" = internal constant %descriptor { ptr @"Number$methodset", i8 34, [11 x i8] zeroinitializer }, align 4
@"reflect/types.type:named:Typed" = internal constant %descriptor { ptr @"Number$methodset", i8 34, [11 x i8] zeroinitializer }, align 4
@"reflect/types.type:named:Symbolic" = internal constant %descriptor { ptr @"Number$methodset", i8 34, [11 x i8] zeroinitializer }, align 4
@external = external global i8
@symbolic = global ptr getelementptr (i8, ptr @"reflect/types.type:named:Symbolic", i32 ptrtoint (ptr @external to i32))

define i1 @narrow() {
  %want = getelementptr i8, ptr getelementptr (%descriptor, ptr @"reflect/types.type:named:Narrow", i32 0, i32 1), i32 4
  %ok = icmp eq ptr getelementptr (i8, ptr getelementptr (i8, ptr @"reflect/types.type:named:Narrow", i3 -4), i32 12), %want
  ret i1 %ok
}

define i1 @typed() {
  %want = getelementptr i8, ptr getelementptr (%descriptor, ptr @"reflect/types.type:named:Typed", i32 0, i32 1), i32 4
  %ok = icmp eq ptr getelementptr inbounds ({ i64, i8 }, ptr @"reflect/types.type:named:Typed", i32 0, i32 1), %want
  ret i1 %ok
}
