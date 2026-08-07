target datalayout = "e-P1-p:16:8-i8:8-i16:8-i32:8-i64:8-f32:8-f64:8-n8:16-a:8"
target triple = "avr-unknown-unknown"

%entry = type { i8, ptr }

@table = global [4 x %entry] zeroinitializer

declare i16 @runtime.gcGlobalRootCount()

declare ptr @runtime.gcGlobalRoot(i16)

declare i16 @runtime.gcGlobalRootSize(i16)
