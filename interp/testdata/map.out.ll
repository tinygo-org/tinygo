target datalayout = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
target triple = "armv6m-none-eabi"

%runtime.hashmap = type { %runtime.hashmap*, i8*, i32, i8, i8, i8 }

@main.m = local_unnamed_addr global %runtime.hashmap* @"main$map"
@main.binaryMap = local_unnamed_addr global %runtime.hashmap* @"main$map.1"
@main.stringMap = local_unnamed_addr global %runtime.hashmap* @"main$map.3"
@main.init.string = internal unnamed_addr constant [7 x i8] c"CONNECT"
@"main$map" = internal global %runtime.hashmap { %runtime.hashmap* null, i8* getelementptr inbounds ({ [8 x i8], i8*, { i8, [7 x i8] }, { { [7 x i8]*, [4 x i8] }, [56 x i8] } }, { [8 x i8], i8*, { i8, [7 x i8] }, { { [7 x i8]*, [4 x i8] }, [56 x i8] } }* @"main$mapbucket", i32 0, i32 0, i32 0), i32 1, i8 1, i8 8, i8 0 }
@"main$mapbucket" = internal unnamed_addr global { [8 x i8], i8*, { i8, [7 x i8] }, { { [7 x i8]*, [4 x i8] }, [56 x i8] } } { [8 x i8] c"\04\00\00\00\00\00\00\00", i8* null, { i8, [7 x i8] } { i8 1, [7 x i8] zeroinitializer }, { { [7 x i8]*, [4 x i8] }, [56 x i8] } { { [7 x i8]*, [4 x i8] } { [7 x i8]* @main.init.string, [4 x i8] c"\07\00\00\00" }, [56 x i8] zeroinitializer } }
@"main$map.1" = internal global %runtime.hashmap { %runtime.hashmap* null, i8* getelementptr inbounds ({ [8 x i8], i8*, { i8, [7 x i8] }, { i8, [7 x i8] } }, { [8 x i8], i8*, { i8, [7 x i8] }, { i8, [7 x i8] } }* @"main$mapbucket.2", i32 0, i32 0, i32 0), i32 1, i8 1, i8 1, i8 0 }
@"main$mapbucket.2" = internal unnamed_addr global { [8 x i8], i8*, { i8, [7 x i8] }, { i8, [7 x i8] } } { [8 x i8] c"\04\00\00\00\00\00\00\00", i8* null, { i8, [7 x i8] } { i8 1, [7 x i8] zeroinitializer }, { i8, [7 x i8] } { i8 2, [7 x i8] zeroinitializer } }
@"main$map.3" = internal global %runtime.hashmap { %runtime.hashmap* null, i8* getelementptr inbounds ({ [8 x i8], i8*, { { [7 x i8]*, [4 x i8] }, [56 x i8] }, { i8, [7 x i8] } }, { [8 x i8], i8*, { { [7 x i8]*, [4 x i8] }, [56 x i8] }, { i8, [7 x i8] } }* @"main$mapbucket.4", i32 0, i32 0, i32 0), i32 1, i8 8, i8 1, i8 0 }
@"main$mapbucket.4" = internal unnamed_addr global { [8 x i8], i8*, { { [7 x i8]*, [4 x i8] }, [56 x i8] }, { i8, [7 x i8] } } { [8 x i8] c"x\00\00\00\00\00\00\00", i8* null, { { [7 x i8]*, [4 x i8] }, [56 x i8] } { { [7 x i8]*, [4 x i8] } { [7 x i8]* @main.init.string, [4 x i8] c"\07\00\00\00" }, [56 x i8] zeroinitializer }, { i8, [7 x i8] } { i8 2, [7 x i8] zeroinitializer } }

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
