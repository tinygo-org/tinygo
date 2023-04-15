target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32--wasi"

@pointerFree12 = local_unnamed_addr global ptr @"main$alloc"
@pointerFree7 = local_unnamed_addr global ptr @"main$alloc.1"
@pointerFree3 = local_unnamed_addr global ptr @"main$alloc.2"
@pointerFree0 = local_unnamed_addr global ptr @"main$alloc.3"
@layout1 = local_unnamed_addr global ptr @"main$alloc.4"
@layout2 = local_unnamed_addr global ptr @"main$alloc.5"
@layout3 = local_unnamed_addr global ptr @"main$alloc.6"
@layout4 = local_unnamed_addr global ptr @"main$alloc.7"
@bigobj1 = local_unnamed_addr global ptr @"main$alloc.8"
@"main$alloc" = internal global [12 x i8] zeroinitializer, align 4
@"main$alloc.1" = internal global [7 x i8] zeroinitializer, align 4
@"main$alloc.2" = internal global [3 x i8] zeroinitializer, align 4
@"main$alloc.3" = internal global [0 x i8] zeroinitializer, align 4
@"main$alloc.4" = internal global [3 x ptr] zeroinitializer, align 4
@"main$alloc.5" = internal global [5 x { ptr, i32, i32 }] zeroinitializer, align 4
@"main$alloc.6" = internal global { ptr, ptr, ptr, i32, i32, ptr, ptr, i32, i32, i32, i32, i32, i32, ptr, ptr, i32, i32, i32, ptr, ptr, i32, i32, ptr, i32, i32, ptr } zeroinitializer, align 4
@"main$alloc.7" = internal global [3 x { ptr, ptr, ptr, i32, i32, ptr, ptr, i32, i32, i32, i32, i32, i32, ptr, ptr, i32, i32, i32, ptr, ptr, i32, i32, ptr, i32, i32, ptr }] zeroinitializer, align 4
@"main$alloc.8" = internal global { ptr, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, ptr } zeroinitializer, align 4

define void @runtime.initAll() unnamed_addr {
  ret void
}
