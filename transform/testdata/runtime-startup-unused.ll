declare i1 @runtime.gorootEnvEnabled(ptr)
declare i1 @runtime.godebugEnvEnabled(ptr)
declare void @tinygo.runtime.feature.goroot()
declare void @tinygo.runtime.feature.godebug()
declare void @"internal/godebug.setUpdate"(ptr, ptr)

define i1 @test.goroot() {
entry:
  %enabled = call i1 @runtime.gorootEnvEnabled(ptr undef)
  ret i1 %enabled
}

define i1 @test.godebug() {
entry:
  call void @"internal/godebug.setUpdate"(ptr null, ptr undef)
  %enabled = call i1 @runtime.godebugEnvEnabled(ptr undef)
  ret i1 %enabled
}
