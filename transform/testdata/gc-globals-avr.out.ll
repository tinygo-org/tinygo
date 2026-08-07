target datalayout = "e-P1-p:16:8-i8:8-i16:8-i32:8-i64:8-f32:8-f64:8-n8:16-a:8"
target triple = "avr-unknown-unknown"

%entry = type { i8, ptr }

@table = global [4 x %entry] zeroinitializer
@runtime.gcGlobalRoots = internal constant [4 x { ptr, i16 }] [{ ptr, i16 } { ptr getelementptr (i8, ptr @table, i16 1), i16 2 }, { ptr, i16 } { ptr getelementptr (i8, ptr @table, i16 4), i16 2 }, { ptr, i16 } { ptr getelementptr (i8, ptr @table, i16 7), i16 2 }, { ptr, i16 } { ptr getelementptr (i8, ptr @table, i16 10), i16 2 }]

define i16 @runtime.gcGlobalRootCount() addrspace(1) {
entry:
  ret i16 4
}

define ptr @runtime.gcGlobalRoot(i16 %0) addrspace(1) {
entry:
  %1 = getelementptr inbounds [4 x { ptr, i16 }], ptr @runtime.gcGlobalRoots, i32 0, i16 %0
  %2 = getelementptr inbounds nuw { ptr, i16 }, ptr %1, i32 0, i32 0
  %3 = load ptr, ptr %2, align 1
  ret ptr %3
}

define i16 @runtime.gcGlobalRootSize(i16 %0) addrspace(1) {
entry:
  %1 = getelementptr inbounds [4 x { ptr, i16 }], ptr @runtime.gcGlobalRoots, i32 0, i16 %0
  %2 = getelementptr inbounds nuw { ptr, i16 }, ptr %1, i32 0, i32 1
  %3 = load i16, ptr %2, align 1
  ret i16 %3
}
