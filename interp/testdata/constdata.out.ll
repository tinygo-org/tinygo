; ModuleID = 'testdata/constdata.ll'
source_filename = "testdata/constdata.ll"
target datalayout = "e-m:e-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.i8 = local_unnamed_addr global [4 x i8] c"zbcd"
@main.i16 = local_unnamed_addr global [3 x i16] [i16 2306, i16 772, i16 1286]
@main.i32 = local_unnamed_addr global [2 x i32] [i32 16909060, i32 84281608]
@main.i64 = local_unnamed_addr global [2 x i64] [i64 72623859790579464, i64 1]
@main.f32 = local_unnamed_addr global [2 x float] [float 1.000000e+00, float 2.000000e+00]
@main.f64 = local_unnamed_addr global [2 x double] [double 1.000000e+00, double 2.000000e+00]
@main.b16 = local_unnamed_addr global i8 1
@main.b32 = local_unnamed_addr global i8 7
@main.b64 = local_unnamed_addr global i8 6
@main.bf32 = local_unnamed_addr global i8 64
@main.bf64 = local_unnamed_addr global i8 64
@main.w16 = local_unnamed_addr global i32 50594050

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
