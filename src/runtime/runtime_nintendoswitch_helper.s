.section .text.armGetSystemTick, "ax", %progbits
.global  armGetSystemTick
.type    armGetSystemTick, %function
.align 2
armGetSystemTick:
    mrs x0, cntpct_el0
    ret

.section .text.armGetSystemTickFreq, "ax", %progbits
.global  armGetSystemTickFreq
.type    armGetSystemTickFreq, %function
.align 2
armGetSystemTickFreq:
    mrs x0, cntfrq_el0
    ret
