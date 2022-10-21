.globaltype __stack_pointer, i32

.global  tinygo_getCurrentStackPointer
.hidden  tinygo_getCurrentStackPointer
.type    tinygo_getCurrentStackPointer,@function
tinygo_getCurrentStackPointer: // func getCurrentStackPointer() uintptr
    .functype tinygo_getCurrentStackPointer() -> (i32)
    global.get __stack_pointer
    return
    end_function
