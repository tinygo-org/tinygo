; ModuleID = 'testdata/constdata-be.ll'
source_filename = "testdata/constdata-be.ll"
target datalayout = "E-m:m-p:32:32-i8:8:32-i16:16:32-i64:64-n32-S64"
target triple = "mips-unknown-linux-gnu"

@main.i8 = local_unnamed_addr global [4 x i8] c"zbcd"
@main.i16 = local_unnamed_addr global [3 x i16] [i16 265, i16 772, i16 1286]
@main.i32 = local_unnamed_addr global [2 x i32] [i32 16909060, i32 84477704]
@main.i64 = local_unnamed_addr global [2 x i64] [i64 72630456860149512, i64 1]
@main.f32 = local_unnamed_addr global [2 x float] [float 1.000000e+00, float 2.000000e+00]
@main.f64 = local_unnamed_addr global [2 x double] [double 1.000000e+00, double 2.000000e+00]
@main.b16 = local_unnamed_addr global i8 2
@main.b32 = local_unnamed_addr global i8 6
@main.b64 = local_unnamed_addr global i8 3
@main.bf32 = local_unnamed_addr global i8 0
@main.bf64 = local_unnamed_addr global i8 0
@main.w16 = local_unnamed_addr global i32 16909060

define void @runtime.initAll() unnamed_addr {
entry:
  ret void
}
