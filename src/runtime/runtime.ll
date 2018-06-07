source_filename = "runtime/runtime.ll"

declare void @runtime.initAll()
declare void @main.main()
declare i8* @main.main$async(i8*)
declare void @runtime.scheduler(i8*)

; Will be changed to true if there are 'go' statements in the compiled program.
@.has_scheduler = private unnamed_addr constant i1 false

define i32 @main() {
	call void @runtime.initAll()
	%has_scheduler = load i1, i1* @.has_scheduler
	; This branch will be optimized away. Only one of the targets will remain.
	br i1 %has_scheduler, label %with_scheduler, label %without_scheduler

with_scheduler:
	; Initialize main and run the scheduler.
	%main = call i8* @main.main$async(i8* null)
	call void @runtime.scheduler(i8* %main)
	ret i32 0

without_scheduler:
	; No scheduler is necessary. Call main directly.
	call void @main.main()
	ret i32 0
}
