
.section .text.tinygo_scanCurrentStack
.global tinygo_scanCurrentStack
tinygo_scanCurrentStack:
    // TODO: save callee saved registers on the stack
    j tinygo_scanstack
