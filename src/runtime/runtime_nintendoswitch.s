.section .text.armGetSystemTick, "ax", %progbits
.global  armGetSystemTick
.type    armGetSystemTick, %function
.align 2
armGetSystemTick:
    mrs x0, cntpct_el0
    ret

.section .text.nxOutputString, "ax", %progbits
.global nxOutputString
.type nxOutputString, %function
.align 2
.cfi_startproc
nxOutputString:
	svc 0x27
	ret
.cfi_endproc

.section .text.exit, "ax", %progbits
.global exit
.type exit, %function
.align 2
exit:
	svc 0x7
	ret

.section .text.setHeapSize, "ax", %progbits
.global setHeapSize
.type setHeapSize, %function
.align 2
setHeapSize:
	str x0, [sp, #-16]!
	svc 0x1
	ldr x2, [sp], #16
	str x1, [x2]
	ret


.section .text.sleepThread, "ax", %progbits
.global sleepThread
.type sleepThread, %function
.align 2
sleepThread:
	svc 0xB
	ret
