; ModuleID = 'inlineasm-riscv.go'
source_filename = "inlineasm-riscv.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32-S128"
target triple = "riscv32-unknown-none"

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.csrGet(ptr %context) unnamed_addr #0 {
entry:
  %0 = call i32 asm sideeffect "csrr $0, 768", "=r"() #1
  ret i32 %0
}

; Function Attrs: nounwind
define hidden void @main.csrSet(i32 %value, ptr %context) unnamed_addr #0 {
entry:
  call void asm sideeffect "csrw 768, $0", "r,~{memory}"(i32 %value) #1
  ret void
}

; Function Attrs: nounwind
define hidden i32 @main.csrSetBits(i32 %mask, ptr %context) unnamed_addr #0 {
entry:
  %0 = call i32 asm sideeffect "csrrs $0, 768, $1", "=r,r,~{memory}"(i32 %mask) #1
  ret i32 %0
}

; Function Attrs: nounwind
define hidden i32 @main.csrClearBits(i32 %mask, ptr %context) unnamed_addr #0 {
entry:
  %0 = call i32 asm sideeffect "csrrc $0, 768, $1", "=r,r,~{memory}"(i32 %mask) #1
  ret i32 %0
}

; Function Attrs: nounwind
define hidden void @main.main(ptr %context) unnamed_addr #0 {
entry:
  ret void
}

attributes #0 = { nounwind "target-features"="+32bit,+a,+c,+m,+zaamo,+zalrsc,+zihintpause,+zmmul,-relax" }
attributes #1 = { nounwind }

!llvm.module.flags = !{!0}

!0 = !{i32 1, !"target-abi", !"ilp32"}
