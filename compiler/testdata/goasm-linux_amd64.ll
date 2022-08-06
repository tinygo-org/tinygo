; ModuleID = 'goasm.go'
source_filename = "goasm.go"
target datalayout = "e-m:e-p270:32:32-p271:32:32-p272:64:64-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux"

%runtime._string = type { ptr, i64 }
%main.AsmStruct = type { i64, i64 }

@main.asmGlobalExport = hidden global i32 0, align 4

@__GoABI0_main.asmGlobalExport = alias i32, ptr @main.asmGlobalExport

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i64, ptr, ptr) #0

; Function Attrs: nounwind uwtable(sync)
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind uwtable(sync)
define hidden double @main.AsmSqrt(double %x, ptr %context) unnamed_addr #1 {
entry:
  %0 = call double asm sideeffect alignstack "subq $$16, %rsp\0A\09movsd %xmm0, 0(%rsp)\0A\09callq \22__GoABI0_main.AsmSqrt\22\0A\09movsd 8(%rsp), %xmm0\0A\09addq $$16, %rsp", "={xmm0},{xmm0},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory},~{rdi},~{rsi},~{rdx},~{rcx},~{r8},~{r9},~{rax},~{rbx},~{rbp},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31}"(double %x) #4
  ret double %0
}

; Function Attrs: nounwind uwtable(sync)
define hidden double @main.AsmAdd(double %x, double %y, ptr %context) unnamed_addr #1 {
entry:
  %0 = call double asm sideeffect alignstack "subq $$32, %rsp\0A\09movsd %xmm0, 0(%rsp)\0A\09movsd %xmm1, 8(%rsp)\0A\09callq \22__GoABI0_main.AsmAdd\22\0A\09movsd 16(%rsp), %xmm0\0A\09addq $$32, %rsp", "={xmm0},{xmm0},{xmm1},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory},~{rdi},~{rsi},~{rdx},~{rcx},~{r8},~{r9},~{rax},~{rbx},~{rbp},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31}"(double %x, double %y) #4
  ret double %0
}

; Function Attrs: nounwind uwtable(sync)
define hidden { i64, double } @main.AsmFoo(double %x, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { i64, double } asm sideeffect alignstack "subq $$32, %rsp\0A\09movsd %xmm0, 0(%rsp)\0A\09callq \22__GoABI0_main.AsmFoo\22\0A\09movq 8(%rsp), %rdi\0A\09movsd 16(%rsp), %xmm0\0A\09addq $$32, %rsp", "={rdi},={xmm0},{xmm0},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory},~{rsi},~{rdx},~{rcx},~{r8},~{r9},~{rax},~{rbx},~{rbp},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31}"(double %x) #4
  ret { i64, double } %0
}

; Function Attrs: nounwind uwtable(sync)
define hidden double @main.asmExport(double %x, ptr %context) unnamed_addr #1 {
entry:
  ret double 0.000000e+00
}

; Function Attrs: noinline nounwind alignstack(16) uwtable(sync)
define internal void @"main.asmExport$goasmwrapper"(ptr %0) #2 {
entry:
  %1 = getelementptr i8, ptr %0, i64 8
  %2 = load double, ptr %1, align 8
  %result = call double @main.asmExport(double %2, ptr null)
  %3 = getelementptr i8, ptr %0, i64 16
  store double %result, ptr %3, align 8
  ret void
}

; Function Attrs: naked nounwind uwtable(sync)
define void @__GoABI0_main.asmExport() #3 {
entry:
  %sp = call ptr asm "mov %rsp, $0", "=r"() #4
  tail call void @"main.asmExport$goasmwrapper"(ptr %sp)
  ret void
}

; Function Attrs: nounwind uwtable(sync)
define hidden { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } @main.AsmAllIntTypes(i1 %b, i64 %i, i8 %i8, i16 %i16, i32 %i32, i64 %i64, i64 %u, i8 %u8, i16 %u16, i32 %u32, i64 %u64, i64 %uptr, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { i1, i64, i8, i16, i32, i64, i64, i8, i16, i32, i64, i64 } asm sideeffect alignstack "subq $$128, %rsp\0A\09movb %dil, 0(%rsp)\0A\09movq %rsi, 8(%rsp)\0A\09movb %dl, 16(%rsp)\0A\09movw %cx, 18(%rsp)\0A\09movl %r8d, 20(%rsp)\0A\09movq %r9, 24(%rsp)\0A\09movq %rax, 32(%rsp)\0A\09movb %bl, 40(%rsp)\0A\09movw %bp, 42(%rsp)\0A\09movl %r10d, 44(%rsp)\0A\09movq %r11, 48(%rsp)\0A\09movq %r12, 56(%rsp)\0A\09callq \22__GoABI0_main.AsmAllIntTypes\22\0A\09movzbl 64(%rsp), %edi\0A\09movq 72(%rsp), %rsi\0A\09movsbl 80(%rsp), %edx\0A\09movswl 82(%rsp), %ecx\0A\09movl 84(%rsp), %r8d\0A\09movq 88(%rsp), %r9\0A\09movq 96(%rsp), %rax\0A\09movzbl 104(%rsp), %ebx\0A\09movzwl 106(%rsp), %ebp\0A\09movl 108(%rsp), %r10d\0A\09movq 112(%rsp), %r11\0A\09movq 120(%rsp), %r12\0A\09addq $$128, %rsp", "={edi},={rsi},={edx},={ecx},={r8d},={r9},={rax},={ebx},={ebp},={r10d},={r11},={r12},{dil},{rsi},{dl},{cx},{r8d},{r9},{rax},{bl},{bp},{r10d},{r11},{r12},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory},~{r13},~{r14},~{r15},~{xmm0},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31}"(i1 %b, i64 %i, i8 %i8, i16 %i16, i32 %i32, i64 %i64, i64 %u, i8 %u8, i16 %u16, i32 %u32, i64 %u64, i64 %uptr) #4
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

; Function Attrs: nounwind uwtable(sync)
define hidden { float, double, { float, float }, { double, double } } @main.AsmAllFloatTypes(float %f32, double %f64, float %c64.r, float %c64.i, double %c128.r, double %c128.i, ptr %context) unnamed_addr #1 {
entry:
  %0 = call { float, double, float, float, double, double } asm sideeffect alignstack "subq $$80, %rsp\0A\09movss %xmm0, 0(%rsp)\0A\09movsd %xmm1, 8(%rsp)\0A\09movss %xmm2, 16(%rsp)\0A\09movss %xmm3, 20(%rsp)\0A\09movsd %xmm4, 24(%rsp)\0A\09movsd %xmm5, 32(%rsp)\0A\09callq \22__GoABI0_main.AsmAllFloatTypes\22\0A\09movss 40(%rsp), %xmm0\0A\09movsd 48(%rsp), %xmm1\0A\09movss 56(%rsp), %xmm2\0A\09movss 60(%rsp), %xmm3\0A\09movsd 64(%rsp), %xmm4\0A\09movsd 72(%rsp), %xmm5\0A\09addq $$80, %rsp", "={xmm0},={xmm1},={xmm2},={xmm3},={xmm4},={xmm5},{xmm0},{xmm1},{xmm2},{xmm3},{xmm4},{xmm5},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory},~{rdi},~{rsi},~{rdx},~{rcx},~{r8},~{r9},~{rax},~{rbx},~{rbp},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31}"(float %f32, double %f64, float %c64.r, float %c64.i, double %c128.r, double %c128.i) #4
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

; Function Attrs: nounwind uwtable(sync)
define hidden { [2 x i64], ptr, ptr, ptr, { ptr, i64, i64 }, %runtime._string, %main.AsmStruct, ptr } @main.AsmAllOtherTypes([2 x i64] %a, ptr dereferenceable_or_null(64) %c, ptr dereferenceable_or_null(80) %m, ptr dereferenceable_or_null(8) %ptr, ptr %slice.data, i64 %slice.len, i64 %slice.cap, ptr %str.data, i64 %str.len, i64 %stru.X, i64 %stru.Y, ptr %uptr, ptr %context) unnamed_addr #1 {
entry:
  %0 = extractvalue [2 x i64] %a, 0
  %1 = extractvalue [2 x i64] %a, 1
  %2 = call { i64, i64, ptr, ptr, ptr, ptr, i64, i64, ptr, i64, i64, i64, ptr } asm sideeffect alignstack "subq $$208, %rsp\0A\09movq %rdi, 0(%rsp)\0A\09movq %rsi, 8(%rsp)\0A\09movq %rdx, 16(%rsp)\0A\09movq %rcx, 24(%rsp)\0A\09movq %r8, 32(%rsp)\0A\09movq %r9, 40(%rsp)\0A\09movq %rax, 48(%rsp)\0A\09movq %rbx, 56(%rsp)\0A\09movq %rbp, 64(%rsp)\0A\09movq %r10, 72(%rsp)\0A\09movq %r11, 80(%rsp)\0A\09movq %r12, 88(%rsp)\0A\09movq %r13, 96(%rsp)\0A\09callq \22__GoABI0_main.AsmAllOtherTypes\22\0A\09movq 104(%rsp), %rdi\0A\09movq 112(%rsp), %rsi\0A\09movq 120(%rsp), %rdx\0A\09movq 128(%rsp), %rcx\0A\09movq 136(%rsp), %r8\0A\09movq 144(%rsp), %r9\0A\09movq 152(%rsp), %rax\0A\09movq 160(%rsp), %rbx\0A\09movq 168(%rsp), %rbp\0A\09movq 176(%rsp), %r10\0A\09movq 184(%rsp), %r11\0A\09movq 192(%rsp), %r12\0A\09movq 200(%rsp), %r13\0A\09addq $$208, %rsp", "={rdi},={rsi},={rdx},={rcx},={r8},={r9},={rax},={rbx},={rbp},={r10},={r11},={r12},={r13},{rdi},{rsi},{rdx},{rcx},{r8},{r9},{rax},{rbx},{rbp},{r10},{r11},{r12},{r13},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory},~{r14},~{r15},~{xmm0},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31}"(i64 %0, i64 %1, ptr %c, ptr %m, ptr %ptr, ptr %slice.data, i64 %slice.len, i64 %slice.len, ptr %str.data, i64 %str.len, i64 %stru.X, i64 %stru.Y, ptr %uptr) #4
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

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+cx8,+fxsr,+mmx,+sse,+sse2,+x87" }
attributes #1 = { nounwind uwtable(sync) "target-features"="+cx8,+fxsr,+mmx,+sse,+sse2,+x87" }
attributes #2 = { noinline nounwind alignstack=16 uwtable(sync) "target-features"="+cx8,+fxsr,+mmx,+sse,+sse2,+x87" }
attributes #3 = { naked nounwind uwtable(sync) "target-features"="+cx8,+fxsr,+mmx,+sse,+sse2,+x87" }
attributes #4 = { nounwind }
