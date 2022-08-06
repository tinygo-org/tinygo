; ModuleID = 'goasm.go'
source_filename = "goasm.go"
target datalayout = "e-m:e-i8:8:32-i16:16:32-i64:64-i128:128-n32:64-S128"
target triple = "aarch64-unknown-linux"

%runtime._string = type { ptr, i64 }
%main.AsmStruct = type { i64, i64 }

@main.asmGlobalExport = hidden global i32 0, align 4

@__GoABI0_main.asmGlobalExport = alias i32, ptr @main.asmGlobalExport

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i64, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden double @main.AsmSqrt(double %x, ptr %context) unnamed_addr #1 {
entry:
  %0 = call double asm sideeffect alignstack "sub sp, sp, #32\0A\09str d0, [sp, #8]\0A\09bl \22__GoABI0_main.AsmSqrt\22\0A\09ldr d0, [sp, #16]\0A\09add sp, sp, #32", "={d0},{d0},~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{lr},~{nzcv},~{ffr},~{vg},~{memory},~{x18},~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{d1},~{d2},~{d3},~{d4},~{d5},~{d6},~{d7},~{d8},~{d9},~{d10},~{d11},~{d12},~{d13},~{d14},~{d15},~{d16},~{d17},~{d18},~{d19},~{d20},~{d21},~{d22},~{d23},~{d24},~{d25},~{d26},~{d27},~{d28},~{d29},~{d30}"(double %x) #4
  ret double %0
}

; Function Attrs: nounwind
define hidden double @main.AsmAdd(double %x, double %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = call double asm sideeffect alignstack "sub sp, sp, #32\0A\09str d0, [sp, #8]\0A\09str d1, [sp, #16]\0A\09bl \22__GoABI0_main.AsmAdd\22\0A\09ldr d0, [sp, #24]\0A\09add sp, sp, #32", "={d0},{d0},{d1},~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{lr},~{nzcv},~{ffr},~{vg},~{memory},~{x18},~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{d1},~{d2},~{d3},~{d4},~{d5},~{d6},~{d7},~{d8},~{d9},~{d10},~{d11},~{d12},~{d13},~{d14},~{d15},~{d16},~{d17},~{d18},~{d19},~{d20},~{d21},~{d22},~{d23},~{d24},~{d25},~{d26},~{d27},~{d28},~{d29},~{d30}"(double %x, double %y) #4
  ret double %0
}

; Function Attrs: nounwind
define hidden { i64, double } @main.AsmFoo(double %x, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { i64, double } asm sideeffect alignstack "sub sp, sp, #32\0A\09str d0, [sp, #8]\0A\09bl \22__GoABI0_main.AsmFoo\22\0A\09ldr x0, [sp, #16]\0A\09ldr d0, [sp, #24]\0A\09add sp, sp, #32", "={x0},={d0},{d0},~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{lr},~{nzcv},~{ffr},~{vg},~{memory},~{x18},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{d1},~{d2},~{d3},~{d4},~{d5},~{d6},~{d7},~{d8},~{d9},~{d10},~{d11},~{d12},~{d13},~{d14},~{d15},~{d16},~{d17},~{d18},~{d19},~{d20},~{d21},~{d22},~{d23},~{d24},~{d25},~{d26},~{d27},~{d28},~{d29},~{d30}"(double %x) #4
  ret { i64, double } %0
}

; Function Attrs: nounwind
define hidden double @main.asmExport(double %x, ptr %context) unnamed_addr #1 {
entry:
  ret double 0.000000e+00
}

; Function Attrs: noinline nounwind
define internal void @"main.asmExport$goasmwrapper"(ptr %0) #2 {
entry:
  %1 = getelementptr i8, ptr %0, i64 8
  %2 = load double, ptr %1, align 8
  %result = call double @main.asmExport(double %2, ptr null)
  %3 = getelementptr i8, ptr %0, i64 16
  store double %result, ptr %3, align 8
  ret void
}

; Function Attrs: naked nounwind
define void @__GoABI0_main.asmExport() #3 {
entry:
  %sp = call ptr asm "mov x0, sp", "=r"() #4
  tail call void @"main.asmExport$goasmwrapper"(ptr %sp)
  ret void
}

; Function Attrs: nounwind
define hidden { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } @main.AsmAllIntTypes(i1 %b, i64 %i, i8 %i8, i16 %i16, i32 %i32, i64 %i64, i64 %u, i8 %u8, i16 %u16, i32 %u32, i64 %u64, i64 %uptr, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } asm sideeffect alignstack "sub sp, sp, #144\0A\09ldrsb w0, [sp, #8]\0A\09ldr x1, [sp, #16]\0A\09ldrsb w2, [sp, #24]\0A\09ldrsh w3, [sp, #26]\0A\09ldr w4, [sp, #28]\0A\09ldr x5, [sp, #32]\0A\09ldr x6, [sp, #40]\0A\09ldrsb w7, [sp, #48]\0A\09ldrsh w8, [sp, #50]\0A\09ldr w9, [sp, #52]\0A\09ldr x10, [sp, #56]\0A\09ldr x11, [sp, #64]\0A\09bl \22__GoABI0_main.AsmAllIntTypes\22\0A\09ldrb w0, [sp, #72]\0A\09ldr x1, [sp, #80]\0A\09ldrsb w2, [sp, #88]\0A\09ldrsh w3, [sp, #90]\0A\09ldr w4, [sp, #92]\0A\09ldr x5, [sp, #96]\0A\09ldr x6, [sp, #104]\0A\09ldrb w7, [sp, #112]\0A\09ldrh w8, [sp, #114]\0A\09ldr w9, [sp, #116]\0A\09ldr x10, [sp, #120]\0A\09ldr x11, [sp, #128]\0A\09add sp, sp, #144", "={w0},={x1},={w2},={w3},={w4},={x5},={x6},={w7},={w8},={w9},={x10},={x11},{w0},{x1},{w2},{w3},{w4},{x5},{x6},{w7},{w8},{w9},{x10},{x11},~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{lr},~{nzcv},~{ffr},~{vg},~{memory},~{x18},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{d0},~{d1},~{d2},~{d3},~{d4},~{d5},~{d6},~{d7},~{d8},~{d9},~{d10},~{d11},~{d12},~{d13},~{d14},~{d15},~{d16},~{d17},~{d18},~{d19},~{d20},~{d21},~{d22},~{d23},~{d24},~{d25},~{d26},~{d27},~{d28},~{d29},~{d30}"(i1 %b, i64 %i, i8 %i8, i16 %i16, i32 %i32, i64 %i64, i64 %u, i8 %u8, i16 %u16, i32 %u32, i64 %u64, i64 %uptr) #4
  %1 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 0
  %2 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 1
  %3 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 2
  %4 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 3
  %5 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 4
  %6 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 5
  %7 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 6
  %8 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 7
  %9 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 8
  %10 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 9
  %11 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 10
  %12 = extractvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %0, 11
  %13 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } zeroinitializer, i1 %1, 0
  %14 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %13, i64 %2, 1
  %15 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %14, i8 %3, 2
  %16 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %15, i16 %4, 3
  %17 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %16, i32 %5, 4
  %18 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %17, i64 %6, 5
  %19 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %18, i64 %7, 6
  %20 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %19, i8 %8, 7
  %21 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %20, i16 %9, 8
  %22 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %21, i32 %10, 9
  %23 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %22, i64 %11, 10
  %24 = insertvalue { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %23, i64 %12, 11
  ret { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } %24
}

; Function Attrs: nounwind
define hidden { float, double, { float, float }, { double, double } } @main.AsmAllFloatTypes(float %f32, double %f64, float %c64.r, float %c64.i, double %c128.r, double %c128.i, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { float, double, float, float, double, double } asm sideeffect alignstack "sub sp, sp, #96\0A\09str s0, [sp, #8]\0A\09str d1, [sp, #16]\0A\09str s2, [sp, #24]\0A\09str s3, [sp, #28]\0A\09str d4, [sp, #32]\0A\09str d5, [sp, #40]\0A\09bl \22__GoABI0_main.AsmAllFloatTypes\22\0A\09ldr s0, [sp, #48]\0A\09ldr d1, [sp, #56]\0A\09ldr s2, [sp, #64]\0A\09ldr s3, [sp, #68]\0A\09ldr d4, [sp, #72]\0A\09ldr d5, [sp, #80]\0A\09add sp, sp, #96", "={s0},={d1},={s2},={s3},={d4},={d5},{s0},{d1},{s2},{s3},{d4},{d5},~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{lr},~{nzcv},~{ffr},~{vg},~{memory},~{x18},~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{d6},~{d7},~{d8},~{d9},~{d10},~{d11},~{d12},~{d13},~{d14},~{d15},~{d16},~{d17},~{d18},~{d19},~{d20},~{d21},~{d22},~{d23},~{d24},~{d25},~{d26},~{d27},~{d28},~{d29},~{d30}"(float %f32, double %f64, float %c64.r, float %c64.i, double %c128.r, double %c128.i) #4
  %1 = extractvalue { float, double, float, float, double, double } %0, 0
  %2 = extractvalue { float, double, float, float, double, double } %0, 1
  %3 = extractvalue { float, double, float, float, double, double } %0, 2
  %4 = extractvalue { float, double, float, float, double, double } %0, 3
  %5 = extractvalue { float, double, float, float, double, double } %0, 4
  %6 = extractvalue { float, double, float, float, double, double } %0, 5
  %7 = insertvalue { float, float } zeroinitializer, float %3, 0
  %8 = insertvalue { float, float } %7, float %4, 1
  %9 = insertvalue { double, double } zeroinitializer, double %5, 0
  %10 = insertvalue { double, double } %9, double %6, 1
  %11 = insertvalue { float, double, { float, float }, { double, double } } zeroinitializer, float %1, 0
  %12 = insertvalue { float, double, { float, float }, { double, double } } %11, double %2, 1
  %13 = insertvalue { float, double, { float, float }, { double, double } } %12, { float, float } %8, 2
  %14 = insertvalue { float, double, { float, float }, { double, double } } %13, { double, double } %10, 3
  ret { float, double, { float, float }, { double, double } } %14
}

; Function Attrs: nounwind
define hidden { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } @main.AsmAllOtherTypes([2 x i64] %a, ptr dereferenceable_or_null(64) %c, ptr dereferenceable_or_null(80) %m, ptr dereferenceable_or_null(8) %ptr, ptr %slice.data, i64 %slice.len, i64 %slice.cap, ptr %str.data, i64 %str.len, i64 %stru.X, i64 %stru.Y, ptr %uptr, ptr %context) unnamed_addr #1 {
entry:
  %0 = extractvalue [2 x i64] %a, 0
  %1 = extractvalue [2 x i64] %a, 1
  %2 = call { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } asm sideeffect alignstack "sub sp, sp, #224\0A\09ldr x0, [sp, #8]\0A\09ldr x1, [sp, #16]\0A\09ldr x2, [sp, #24]\0A\09ldr x3, [sp, #32]\0A\09ldr x4, [sp, #40]\0A\09ldr x5, [sp, #48]\0A\09ldr x6, [sp, #56]\0A\09ldr x7, [sp, #64]\0A\09ldr x8, [sp, #72]\0A\09ldr x9, [sp, #80]\0A\09ldr x10, [sp, #88]\0A\09ldr x11, [sp, #96]\0A\09ldr x12, [sp, #104]\0A\09bl \22__GoABI0_main.AsmAllOtherTypes\22\0A\09ldr x0, [sp, #112]\0A\09ldr x1, [sp, #120]\0A\09ldr x2, [sp, #128]\0A\09ldr x3, [sp, #136]\0A\09ldr x4, [sp, #144]\0A\09ldr x5, [sp, #152]\0A\09ldr x6, [sp, #160]\0A\09ldr x7, [sp, #168]\0A\09ldr x8, [sp, #176]\0A\09ldr x9, [sp, #184]\0A\09ldr x10, [sp, #192]\0A\09ldr x11, [sp, #200]\0A\09ldr x12, [sp, #208]\0A\09add sp, sp, #224", "={x0},={x1},={x2},={x3},={x4},={x5},={x6},={x7},={x8},={x9},={x10},={x11},={x12},{x0},{x1},{x2},{x3},{x4},{x5},{x6},{x7},{x8},{x9},{x10},{x11},{x12},~{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{lr},~{nzcv},~{ffr},~{vg},~{memory},~{x18},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{d0},~{d1},~{d2},~{d3},~{d4},~{d5},~{d6},~{d7},~{d8},~{d9},~{d10},~{d11},~{d12},~{d13},~{d14},~{d15},~{d16},~{d17},~{d18},~{d19},~{d20},~{d21},~{d22},~{d23},~{d24},~{d25},~{d26},~{d27},~{d28},~{d29},~{d30}"(i64 %0, i64 %1, ptr %c, ptr %m, ptr %ptr, ptr %slice.data, i64 %slice.len, i64 %slice.len, ptr %str.data, i64 %str.len, i64 %stru.X, i64 %stru.Y, ptr %uptr) #4
  %3 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 0
  %4 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 1
  %5 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 2
  %6 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 3
  %7 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 4
  %8 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 5
  %9 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 6
  %10 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 7
  %11 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 8
  %12 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 9
  %13 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 10
  %14 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 11
  %15 = extractvalue { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } %2, 12
  %16 = insertvalue [2 x i64] zeroinitializer, i64 %3, 0
  %17 = insertvalue [2 x i64] %16, i64 %4, 1
  %18 = insertvalue { ptr, i64, i64 } zeroinitializer, ptr %8, 0
  %19 = insertvalue { ptr, i64, i64 } %18, i64 %9, 1
  %20 = insertvalue { ptr, i64, i64 } %19, i64 %10, 2
  %21 = insertvalue %runtime._string zeroinitializer, ptr %11, 0
  %22 = insertvalue %runtime._string %21, i64 %12, 1
  %23 = insertvalue %main.AsmStruct zeroinitializer, i64 %13, 0
  %24 = insertvalue %main.AsmStruct %23, i64 %14, 1
  %25 = insertvalue { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } zeroinitializer, [2 x i64] %17, 0
  %26 = insertvalue { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } %25, ptr %5, 1
  %27 = insertvalue { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } %26, ptr %6, 2
  %28 = insertvalue { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } %27, ptr %7, 3
  %29 = insertvalue { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } %28, { ptr, i64, i64 } %20, 4
  %30 = insertvalue { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } %29, %runtime._string %22, 5
  %31 = insertvalue { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } %30, %main.AsmStruct %24, 6
  %32 = insertvalue { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } %31, ptr %15, 7
  ret { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } %32
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+neon,-fmv" }
attributes #1 = { nounwind "target-features"="+neon,-fmv" }
attributes #2 = { noinline nounwind "target-features"="+neon,-fmv" }
attributes #3 = { naked nounwind "target-features"="+neon,-fmv" }
attributes #4 = { nounwind }
