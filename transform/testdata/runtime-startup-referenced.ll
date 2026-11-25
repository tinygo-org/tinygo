@goroot = global ptr @runtime.GOROOT
@godebug = global ptr @"(*internal/godebug.Setting).Value"

declare i1 @runtime.gorootEnvEnabled(ptr)
declare i1 @runtime.godebugEnvEnabled(ptr)
declare ptr @runtime.GOROOT(ptr)
declare ptr @"(*internal/godebug.Setting).Value"(ptr, ptr)

define i1 @test.goroot() {
entry:
  %enabled = call i1 @runtime.gorootEnvEnabled(ptr undef)
  ret i1 %enabled
}

define i1 @test.godebug() {
entry:
  %enabled = call i1 @runtime.godebugEnvEnabled(ptr undef)
  ret i1 %enabled
}
