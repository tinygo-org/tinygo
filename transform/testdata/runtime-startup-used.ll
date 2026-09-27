declare i1 @runtime.gorootEnvEnabled(ptr)
declare i1 @runtime.godebugEnvEnabled(ptr)
declare void @tinygo.runtime.feature.goroot()
declare void @tinygo.runtime.feature.godebug()

define i1 @test.goroot() {
entry:
  call void @tinygo.runtime.feature.goroot()
  %enabled = call i1 @runtime.gorootEnvEnabled(ptr undef)
  ret i1 %enabled
}

define i1 @test.godebug() {
entry:
  call void @tinygo.runtime.feature.godebug()
  %enabled = call i1 @runtime.godebugEnvEnabled(ptr undef)
  ret i1 %enabled
}
