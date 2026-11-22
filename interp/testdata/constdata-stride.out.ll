; ModuleID = 'testdata/constdata-stride.ll'
source_filename = "testdata/constdata-stride.ll"
target datalayout = "e-m:e-i16:32-i64:64-n8:16:32:64-S128"
target triple = "x86_64--linux"

@main.src = local_unnamed_addr global [2 x i16] [i16 258, i16 772]
@main.dst = local_unnamed_addr global [2 x i16] [i16 0, i16 9]
@main.read = local_unnamed_addr global i8 4

define void @runtime.initAll() local_unnamed_addr {
  ret void
}
