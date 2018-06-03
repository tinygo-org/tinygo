source_filename = "runtime/runtime.ll"

declare void @runtime.initAll()
declare void @main.main()

define i32 @main() {
	call void @runtime.initAll()
	call void @main.main()
	ret i32 0
}
