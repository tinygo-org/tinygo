declare i1 @runtime.gorootEnvEnabled(ptr)
declare i1 @runtime.godebugEnvEnabled(ptr)
declare void @tinygo.runtime.feature.goroot()
declare void @tinygo.runtime.feature.godebug()
declare void @"internal/godebug.setUpdate"(ptr, ptr)
declare ptr @runtime.GOROOT(ptr)
declare ptr @"(*internal/godebug.Setting).Value"(ptr, ptr)

define internal void @test.noop(ptr %callback) {
entry:
  ret void
}

define internal void @test.unusedCallback() {
entry:
  call void @tinygo.runtime.feature.goroot()
  %goroot = call ptr @runtime.GOROOT(ptr undef)
  call void @tinygo.runtime.feature.godebug()
  %godebug = call ptr @"(*internal/godebug.Setting).Value"(ptr null, ptr undef)
  ret void
}

define i1 @test.goroot() {
entry:
  call void @test.noop(ptr @test.unusedCallback)
  %enabled = call i1 @runtime.gorootEnvEnabled(ptr undef)
  ret i1 %enabled
}

define i1 @test.godebug() {
entry:
  call void @"internal/godebug.setUpdate"(ptr null, ptr undef)
  %enabled = call i1 @runtime.godebugEnvEnabled(ptr undef)
  ret i1 %enabled
}
