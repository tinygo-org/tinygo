target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv6m-none-eabi"

%runtime.hashmap = type { %runtime.hashmap*, i8*, i32, i8, i8, i8 }
%runtime._string = type { i8*, i32 }

@main.m = local_unnamed_addr global %runtime.hashmap* @"main$map"
@main.init.string = internal unnamed_addr constant [7 x i8] c"CONNECT"
@"main$mapbucket" = internal unnamed_addr global { [8 x i8], i8*, [8 x i8], [8 x %runtime._string] } { [8 x i8] c"\04\00\00\00\00\00\00\00", i8* null, [8 x i8] c"\01\00\00\00\00\00\00\00", [8 x %runtime._string] [%runtime._string { i8* getelementptr inbounds ([7 x i8], [7 x i8]* @main.init.string, i32 0, i32 0), i32 7 }, %runtime._string zeroinitializer, %runtime._string zeroinitializer, %runtime._string zeroinitializer, %runtime._string zeroinitializer, %runtime._string zeroinitializer, %runtime._string zeroinitializer, %runtime._string zeroinitializer] }
@"main$map" = internal unnamed_addr global %runtime.hashmap { %runtime.hashmap* null, i8* getelementptr inbounds ({ [8 x i8], i8*, [8 x i8], [8 x %runtime._string] }, { [8 x i8], i8*, [8 x i8], [8 x %runtime._string] }* @"main$mapbucket", i32 0, i32 0, i32 0), i32 1, i8 1, i8 8, i8 0 }

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
