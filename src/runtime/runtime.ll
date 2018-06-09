source_filename = "runtime/runtime.ll"

%interface = type { i32, i8* }

declare void @runtime.initAll()
declare void @main.main()
declare i8* @main.main$async(i8*)
declare void @runtime.scheduler(i8*)

; Will be changed to true if there are 'go' statements in the compiled program.
@has_scheduler = private unnamed_addr constant i1 false

; Will be changed by the compiler to the first type number with methods.
@first_interface_num = private unnamed_addr constant i32 0

; Will be filled by the compiler with runtime type information.
%interface_tuple = type { i32, i32 } ; { index, len }
@interface_tuples = external global [0 x %interface_tuple]
@interface_signatures = external global [0 x i32] ; array of method IDs
@interface_functions = external global [0 x i8*] ; array of function pointers

define i32 @main() {
    call void @runtime.initAll()
    %has_scheduler = load i1, i1* @has_scheduler
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

; Get the function pointer for the method on the interface.
; This function only reads constant global data and it's own arguments so it can
; be 'readnone' (a pure function).
define i8* @itfmethod(%interface %itf, i32 %method) noinline readnone {
entry:
    ; Calculate the index in @interface_tuples
    %concrete_type_num = extractvalue %interface %itf, 0
    %first_interface_num = load i32, i32* @first_interface_num
    %index = sub i32 %concrete_type_num, %first_interface_num

    ; Calculate the index for @interface_signatures and @interface_functions
    %itf_index_ptr = getelementptr inbounds [0 x %interface_tuple], [0 x %interface_tuple]* @interface_tuples, i32 0, i32 %index, i32 0
    %itf_index = load i32, i32* %itf_index_ptr
    br label %find_method

    ; This is a while loop until the method has been found.
    ; It must be in here, so avoid checking the length.
find_method:
    %itf_index.phi = phi i32 [ %itf_index, %entry], [ %itf_index.phi.next, %find_method]
    %m_ptr = getelementptr inbounds [0 x i32], [0 x i32]* @interface_signatures, i32 0, i32 %itf_index.phi
    %m = load i32, i32* %m_ptr
    %found = icmp eq i32 %m, %method
    %itf_index.phi.next = add i32 %itf_index.phi, 1
    br i1 %found, label %found_method, label %find_method

found_method:
    %fp_ptr = getelementptr inbounds [0 x i8*], [0 x i8*]* @interface_functions, i32 0, i32 %itf_index.phi
    %fp = load i8*, i8** %fp_ptr
    ret i8* %fp
}
