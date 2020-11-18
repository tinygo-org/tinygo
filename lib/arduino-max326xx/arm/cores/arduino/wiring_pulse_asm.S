/*
  Copyright (c) 2015 Arduino LLC.  All right reserved.
  This library is free software; you can redistribute it and/or
  modify it under the terms of the GNU Lesser General Public
  License as published by the Free Software Foundation; either
  version 2.1 of the License, or (at your option) any later version.
  This library is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
  See the GNU Lesser General Public License for more details.
  You should have received a copy of the GNU Lesser General Public
  License along with this library; if not, write to the Free Software
  Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
  
  Modified 2017 by Maxim Integrated for MAX326xx
  
*/

/*
The following function has been compiled to ASM with gcc
    unsigned long countPulseASM(const volatile uint32_t *port, uint32_t bit, uint32_t stateMask, unsigned long maxloops)
    {
      unsigned long width = 0;
      // wait for any previous pulse to end
      while ((*port & bit) == stateMask)
        if (--maxloops == 0)
          return 0;
      // wait for the pulse to start
      while ((*port & bit) != stateMask)
        if (--maxloops == 0)
          return 0;
      // wait for the pulse to stop
      while ((*port & bit) == stateMask) {
        if (++width == maxloops)
          return 0;
      }
      return width;
    }

using the command line:

	arm-none-eabi-gcc -mcpu=cortex-m3 -mthumb -c -O2 -W -ffunction-sections -fdata-sections -nostdlib \
		countPulseASM.c -Wa,-ahlmsd=output.lst -dp -fverbose-asm -S \
		-I.arduino15/packages/arduino/hardware/sam/1.6.3/cores/arduino \
		-I.arduino15/packages/arduino/hardware/sam/1.6.3/system/CMSIS/CMSIS/Include \
		-I.arduino15/packages/arduino/hardware/sam/1.6.3/system/CMSIS/Device/ATMEL \
		-I.arduino15/packages/arduino/hardware/sam/1.6.3/system/libsam/include \
		-I.arduino15/packages/arduino/hardware/sam/1.6.3/variants/arduino_due_x

The result has been slightly edited to increase readability.
*/

	.syntax unified
	.cpu cortex-m3
	.fpu softvfp
	.eabi_attribute 20, 1	@ Tag_ABI_FP_denormal
	.eabi_attribute 21, 1	@ Tag_ABI_FP_exceptions
	.eabi_attribute 23, 3	@ Tag_ABI_FP_number_model
	.eabi_attribute 24, 1	@ Tag_ABI_align8_needed
	.eabi_attribute 25, 1	@ Tag_ABI_align8_preserved
	.eabi_attribute 26, 1	@ Tag_ABI_enum_size
	.eabi_attribute 30, 2	@ Tag_ABI_optimization_goals
	.eabi_attribute 34, 1	@ Tag_CPU_unaligned_access
	.eabi_attribute 18, 4	@ Tag_ABI_PCS_wchar_t
@	.file	"countPulseASM.c"
    
@ GNU C (GNU Tools for ARM Embedded Processors) version 4.9.3 20150303 (release) [ARM/embedded-4_9-branch revision 221220] (arm-none-eabi)
@	compiled by GNU C version 4.7.4, GMP version 4.3.2, MPFR version 2.4.2, MPC version 0.8.1
@ GGC heuristics: --param ggc-min-expand=100 --param ggc-min-heapsize=131072
@ options passed: 
@ -I .arduino15/packages/arduino/hardware/sam/1.6.3/cores/arduino
@ -I .arduino15/packages/arduino/hardware/sam/1.6.3/system/CMSIS/CMSIS/Include
@ -I .arduino15/packages/arduino/hardware/sam/1.6.3/system/CMSIS/Device/ATMEL
@ -I .arduino15/packages/arduino/hardware/sam/1.6.3/system/libsam/include
@ -I .arduino15/packages/arduino/hardware/sam/1.6.3/variants/arduino_due_x
@ -imultilib armv7-m -iprefix /usr/bin/../lib/gcc/arm-none-eabi/4.9.3/
@ -isysroot /usr/bin/../arm-none-eabi -D__USES_INITFINI__ countPulseASM.c
@ -mcpu=cortex-m3 -mthumb -O2 -Wextra -ffunction-sections -fdata-sections
@ -fverbose-asm
@ options enabled:  -faggressive-loop-optimizations -fauto-inc-dec
@ -fbranch-count-reg -fcaller-saves -fcombine-stack-adjustments -fcommon
@ -fcompare-elim -fcprop-registers -fcrossjumping -fcse-follow-jumps
@ -fdata-sections -fdefer-pop -fdelete-null-pointer-checks -fdevirtualize
@ -fdevirtualize-speculatively -fdwarf2-cfi-asm -fearly-inlining
@ -feliminate-unused-debug-types -fexpensive-optimizations
@ -fforward-propagate -ffunction-cse -ffunction-sections -fgcse -fgcse-lm
@ -fgnu-runtime -fgnu-unique -fguess-branch-probability
@ -fhoist-adjacent-loads -fident -fif-conversion -fif-conversion2
@ -findirect-inlining -finline -finline-atomics
@ -finline-functions-called-once -finline-small-functions -fipa-cp
@ -fipa-profile -fipa-pure-const -fipa-reference -fipa-sra
@ -fira-hoist-pressure -fira-share-save-slots -fira-share-spill-slots
@ -fisolate-erroneous-paths-dereference -fivopts -fkeep-static-consts
@ -fleading-underscore -flifetime-dse -fmath-errno -fmerge-constants
@ -fmerge-debug-strings -fmove-loop-invariants -fomit-frame-pointer
@ -foptimize-sibling-calls -foptimize-strlen -fpartial-inlining -fpeephole
@ -fpeephole2 -fprefetch-loop-arrays -freg-struct-return -freorder-blocks
@ -freorder-functions -frerun-cse-after-loop
@ -fsched-critical-path-heuristic -fsched-dep-count-heuristic
@ -fsched-group-heuristic -fsched-interblock -fsched-last-insn-heuristic
@ -fsched-pressure -fsched-rank-heuristic -fsched-spec
@ -fsched-spec-insn-heuristic -fsched-stalled-insns-dep -fschedule-insns
@ -fschedule-insns2 -fsection-anchors -fshow-column -fshrink-wrap
@ -fsigned-zeros -fsplit-ivs-in-unroller -fsplit-wide-types
@ -fstrict-aliasing -fstrict-overflow -fstrict-volatile-bitfields
@ -fsync-libcalls -fthread-jumps -ftoplevel-reorder -ftrapping-math
@ -ftree-bit-ccp -ftree-builtin-call-dce -ftree-ccp -ftree-ch
@ -ftree-coalesce-vars -ftree-copy-prop -ftree-copyrename -ftree-cselim
@ -ftree-dce -ftree-dominator-opts -ftree-dse -ftree-forwprop -ftree-fre
@ -ftree-loop-if-convert -ftree-loop-im -ftree-loop-ivcanon
@ -ftree-loop-optimize -ftree-parallelize-loops= -ftree-phiprop -ftree-pre
@ -ftree-pta -ftree-reassoc -ftree-scev-cprop -ftree-sink -ftree-slsr
@ -ftree-sra -ftree-switch-conversion -ftree-tail-merge -ftree-ter
@ -ftree-vrp -funit-at-a-time -fverbose-asm -fzero-initialized-in-bss
@ -mfix-cortex-m3-ldrd -mlittle-endian -mlra -mpic-data-is-text-relative
@ -msched-prolog -mthumb -munaligned-access -mvectorize-with-neon-quad

	.section	.text.countPulseASM,"ax",%progbits
	.align	2
	.global	countPulseASM
	.thumb
	.thumb_func
	.type	countPulseASM, %function
countPulseASM:
	@ args = 0, pretend = 0, frame = 0
	@ frame_needed = 0, uses_anonymous_args = 0
	@ link register save eliminated.
	push	{r4, r5}	@						@ 132	*push_multi						[length = 2]
	b		.L2			@						@ 178	*arm_jump						[length = 2]
.L4:
	subs	r3, r3, #1	@ maxloops, maxloops,	@ 18	thumb2_addsi3_compare0/1		[length = 2]
	beq		.L12		@,						@ 19	arm_cond_branch					[length = 2]
.L2:
	ldr		r4, [r0]	@ D.4169, *port_7(D)	@ 22	*thumb2_movsi_insn/6			[length = 4]
	ands	r4, r4, r1	@, D.4169, D.4169, bit	@ 24	*thumb2_alusi3_short			[length = 2]
	cmp		r4, r2		@ D.4169, stateMask		@ 25	*arm_cmpsi_insn/2				[length = 2]
	beq		.L4			@,						@ 26	arm_cond_branch					[length = 2]
	b		.L6			@						@ 181	*arm_jump						[length = 2]
.L7:
	subs	r3, r3, #1	@ maxloops, maxloops,	@ 32	thumb2_addsi3_compare0/1		[length = 2]
	beq		.L12		@,						@ 33	arm_cond_branch					[length = 2]
.L6:
	ldr		r4, [r0]	@ D.4169, *port_7(D)	@ 41	*thumb2_movsi_insn/6			[length = 4]
	ands	r4, r4, r1	@, D.4169, D.4169, bit	@ 43	*thumb2_alusi3_short			[length = 2]
	cmp		r4, r2		@ D.4169, stateMask		@ 44	*arm_cmpsi_insn/2				[length = 2]
	bne		.L7			@,						@ 45	arm_cond_branch					[length = 2]
	movs	r5, #0		@ width,				@ 7		*thumb2_movsi_shortim			[length = 2]
	b		.L9			@						@ 183	*arm_jump						[length = 2]
.L10:
	adds	r5, r5, #1	@ width, width,			@ 50	*thumb2_addsi_short/1			[length = 2]
	cmp		r3, r5		@ maxloops, width		@ 51	*arm_cmpsi_insn/2				[length = 2]
	beq		.L22		@,						@ 52	arm_cond_branch					[length = 2]
.L9:
	ldr		r4, [r0]	@ D.4169, *port_7(D)	@ 60	*thumb2_movsi_insn/6			[length = 4]
	ands	r4, r4, r1	@, D.4169, D.4169, bit	@ 62	*thumb2_alusi3_short			[length = 2]
	cmp		r4, r2		@ D.4169, stateMask		@ 63	*arm_cmpsi_insn/2				[length = 2]
	beq		.L10		@,						@ 64	arm_cond_branch					[length = 2]
	mov		r0, r5		@ D.4169, width			@ 9		*thumb2_movsi_insn/1			[length = 2]
	pop		{r4, r5}	@						@ 165	*load_multiple_with_writeback	[length = 4]
	bx		lr			@						@ 166	*thumb2_return					[length = 4]
.L12:
	mov		r0, r3		@ D.4169, maxloops		@ 8		*thumb2_movsi_insn/1			[length = 2]
	pop		{r4, r5}	@						@ 137	*load_multiple_with_writeback	[length = 4]
	bx		lr			@						@ 138	*thumb2_return					[length = 4]
.L22:
	movs	r0, #0		@ D.4169,				@ 11	*thumb2_movsi_shortim			[length = 2]
	pop		{r4, r5}	@						@ 173	*load_multiple_with_writeback	[length = 4]
	bx		lr			@						@ 174	*thumb2_return					[length = 4]
	.size	countPulseASM, .-countPulseASM
	.ident	"GCC: (GNU Tools for ARM Embedded Processors) 4.9.3 20150303 (release) [ARM/embedded-4_9-branch revision 221220]"
