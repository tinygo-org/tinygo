target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32--wasi"

@pointerFree12 = local_unnamed_addr global i8* getelementptr inbounds ([12 x i8], [12 x i8]* @"main$alloc", i32 0, i32 0)
@pointerFree7 = local_unnamed_addr global i8* getelementptr inbounds ([7 x i8], [7 x i8]* @"main$alloc.1", i32 0, i32 0)
@pointerFree3 = local_unnamed_addr global i8* getelementptr inbounds ([3 x i8], [3 x i8]* @"main$alloc.2", i32 0, i32 0)
@pointerFree0 = local_unnamed_addr global i8* getelementptr inbounds ([0 x i8], [0 x i8]* @"main$alloc.3", i32 0, i32 0)
@layout1 = local_unnamed_addr global i8* bitcast ([3 x i8*]* @"main$alloc.4" to i8*)
@layout2 = local_unnamed_addr global i8* bitcast ([5 x { i8*, i32, i32 }]* @"main$alloc.5" to i8*)
@layout3 = local_unnamed_addr global i8* bitcast ({ i8*, i8*, i8*, i32, i32, i8*, i8*, i32, i32, i32, i32, i32, i32, i8*, i8*, i32, i32, i32, i8*, i8*, i32, i32, i8*, i32, i32, i8* }* @"main$alloc.6" to i8*)
@layout4 = local_unnamed_addr global i8* bitcast ([3 x { i8*, i8*, i8*, i32, i32, i8*, i8*, i32, i32, i32, i32, i32, i32, i8*, i8*, i32, i32, i32, i8*, i8*, i32, i32, i8*, i32, i32, i8* }]* @"main$alloc.7" to i8*)
@bigobj1 = local_unnamed_addr global i8* bitcast ({ i8*, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i8* }* @"main$alloc.8" to i8*)
@"main$alloc" = internal global [12 x i8] zeroinitializer, align 4
@"main$alloc.1" = internal global [7 x i8] zeroinitializer, align 4
@"main$alloc.2" = internal global [3 x i8] zeroinitializer, align 4
@"main$alloc.3" = internal global [0 x i8] zeroinitializer, align 4
@"main$alloc.4" = internal global [3 x i8*] zeroinitializer, align 4
@"main$alloc.5" = internal global [5 x { i8*, i32, i32 }] zeroinitializer, align 4
@"main$alloc.6" = internal global { i8*, i8*, i8*, i32, i32, i8*, i8*, i32, i32, i32, i32, i32, i32, i8*, i8*, i32, i32, i32, i8*, i8*, i32, i32, i8*, i32, i32, i8* } zeroinitializer, align 4
@"main$alloc.7" = internal global [3 x { i8*, i8*, i8*, i32, i32, i8*, i8*, i32, i32, i32, i32, i32, i32, i8*, i8*, i32, i32, i32, i8*, i8*, i32, i32, i8*, i32, i32, i8* }] zeroinitializer, align 4
@"main$alloc.8" = internal global { i8*, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i8* } zeroinitializer, align 4

define void @runtime.initAll() unnamed_addr {
  ret void
}
