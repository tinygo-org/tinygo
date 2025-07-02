; ModuleID = 'gc.go'
source_filename = "gc.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime._interface = type { ptr, ptr }

@main.scalar1 = hidden global ptr null, align 4
@main.scalar2 = hidden global ptr null, align 4
@main.scalar3 = hidden global ptr null, align 4
@main.scalar4 = hidden global ptr null, align 4
@main.array1 = hidden global ptr null, align 4
@main.array2 = hidden global ptr null, align 4
@main.array3 = hidden global ptr null, align 4
@main.struct1 = hidden global ptr null, align 4
@main.struct2 = hidden global ptr null, align 4
@main.struct3 = hidden global ptr null, align 4
@main.struct4 = hidden global ptr null, align 4
@main.slice1 = hidden global { ptr, i32, i32 } zeroinitializer, align 4
@main.slice2 = hidden global { ptr, i32, i32 } zeroinitializer, align 4
@main.slice3 = hidden global { ptr, i32, i32 } zeroinitializer, align 4
@"runtime/gc.layout:62-2000000000000001" = linkonce_odr unnamed_addr constant { i32, [8 x i8] } { i32 62, [8 x i8] c"\01\00\00\00\00\00\00 " }
@"runtime/gc.layout:62-0001" = linkonce_odr unnamed_addr constant { i32, [8 x i8] } { i32 62, [8 x i8] c"\01\00\00\00\00\00\00\00" }
@"reflect/types.type:basic:complex128" = linkonce_odr constant { i8, ptr } { i8 80, ptr @"reflect/types.type:pointer:basic:complex128" }, align 4
@"reflect/types.type:pointer:basic:complex128" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:complex128" }, align 4

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #1

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #2 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.newScalar(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %new = call align 1 dereferenceable(1) ptr @runtime.alloc(i32 1, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new, ptr @main.scalar1, align 4
  %new1 = call align 4 dereferenceable(4) ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new1, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new1, ptr @main.scalar2, align 4
  %new2 = call align 8 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new2, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new2, ptr @main.scalar3, align 4
  %new3 = call align 4 dereferenceable(4) ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new3, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new3, ptr @main.scalar4, align 4
  ret void
}

; Function Attrs: nounwind
define hidden void @main.newArray(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %new = call align 1 dereferenceable(3) ptr @runtime.alloc(i32 3, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new, ptr @main.array1, align 4
  %new1 = call align 1 dereferenceable(71) ptr @runtime.alloc(i32 71, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new1, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new1, ptr @main.array2, align 4
  %new2 = call align 4 dereferenceable(12) ptr @runtime.alloc(i32 12, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new2, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new2, ptr @main.array3, align 4
  ret void
}

; Function Attrs: nounwind
define hidden void @main.newStruct(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %new = call align 1 ptr @runtime.alloc(i32 0, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new, ptr @main.struct1, align 4
  %new1 = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new1, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new1, ptr @main.struct2, align 4
  %new2 = call align 4 dereferenceable(248) ptr @runtime.alloc(i32 248, ptr nonnull @"runtime/gc.layout:62-2000000000000001", ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new2, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new2, ptr @main.struct3, align 4
  %new3 = call align 4 dereferenceable(248) ptr @runtime.alloc(i32 248, ptr nonnull @"runtime/gc.layout:62-0001", ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new3, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %new3, ptr @main.struct4, align 4
  ret void
}

; Function Attrs: nounwind
define hidden ptr @main.newFuncValue(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %new = call align 4 dereferenceable(8) ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 197 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %new, ptr nonnull %stackalloc, ptr undef) #3
  ret ptr %new
}

; Function Attrs: nounwind
define hidden void @main.makeSlice(ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %makeslice = call align 1 dereferenceable(5) ptr @runtime.alloc(i32 5, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %makeslice, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %makeslice, ptr @main.slice1, align 4
  store i32 5, ptr getelementptr inbounds (i8, ptr @main.slice1, i32 4), align 4
  store i32 5, ptr getelementptr inbounds (i8, ptr @main.slice1, i32 8), align 4
  %makeslice1 = call align 4 dereferenceable(20) ptr @runtime.alloc(i32 20, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %makeslice1, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %makeslice1, ptr @main.slice2, align 4
  store i32 5, ptr getelementptr inbounds (i8, ptr @main.slice2, i32 4), align 4
  store i32 5, ptr getelementptr inbounds (i8, ptr @main.slice2, i32 8), align 4
  %makeslice3 = call align 4 dereferenceable(60) ptr @runtime.alloc(i32 60, ptr nonnull inttoptr (i32 71 to ptr), ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %makeslice3, ptr nonnull %stackalloc, ptr undef) #3
  store ptr %makeslice3, ptr @main.slice3, align 4
  store i32 5, ptr getelementptr inbounds (i8, ptr @main.slice3, i32 4), align 4
  store i32 5, ptr getelementptr inbounds (i8, ptr @main.slice3, i32 8), align 4
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.makeInterface(double %v.r, double %v.i, ptr %context) unnamed_addr #2 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call align 8 dereferenceable(16) ptr @runtime.alloc(i32 16, ptr null, ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #3
  store double %v.r, ptr %0, align 8
  %.repack1 = getelementptr inbounds i8, ptr %0, i32 8
  store double %v.i, ptr %.repack1, align 8
  %1 = insertvalue %runtime._interface { ptr @"reflect/types.type:basic:complex128", ptr undef }, ptr %0, 1
  call void @runtime.trackPointer(ptr nonnull @"reflect/types.type:basic:complex128", ptr nonnull %stackalloc, ptr undef) #3
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #3
  ret %runtime._interface %1
}

attributes #0 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { nounwind "target-features"="+bulk-memory,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nounwind }
