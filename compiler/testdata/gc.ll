; ModuleID = 'gc.go'
source_filename = "gc.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.typecodeID = type { ptr, i32, ptr, ptr, i32 }
%runtime._interface = type { i32, ptr }

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
@main.slice1 = hidden global { ptr, i32, i32 } zeroinitializer, align 8
@main.slice2 = hidden global { ptr, i32, i32 } zeroinitializer, align 8
@main.slice3 = hidden global { ptr, i32, i32 } zeroinitializer, align 8
@"runtime/gc.layout:62-2000000000000001" = linkonce_odr unnamed_addr constant { i32, [8 x i8] } { i32 62, [8 x i8] c" \00\00\00\00\00\00\01" }
@"runtime/gc.layout:62-0001" = linkonce_odr unnamed_addr constant { i32, [8 x i8] } { i32 62, [8 x i8] c"\00\00\00\00\00\00\00\01" }
@"reflect/types.type:basic:complex128" = linkonce_odr constant %runtime.typecodeID { ptr null, i32 0, ptr null, ptr @"reflect/types.type:pointer:basic:complex128", i32 0 }
@"reflect/types.type:pointer:basic:complex128" = linkonce_odr constant %runtime.typecodeID { ptr @"reflect/types.type:basic:complex128", i32 0, ptr null, ptr null, i32 0 }

declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #0

declare void @runtime.trackPointer(ptr nocapture readonly, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.newScalar(ptr %context) unnamed_addr #1 {
entry:
  %new = call ptr @runtime.alloc(i32 1, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new, ptr undef) #2
  store ptr %new, ptr @main.scalar1, align 4
  %new1 = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new1, ptr undef) #2
  store ptr %new1, ptr @main.scalar2, align 4
  %new2 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new2, ptr undef) #2
  store ptr %new2, ptr @main.scalar3, align 4
  %new3 = call ptr @runtime.alloc(i32 4, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new3, ptr undef) #2
  store ptr %new3, ptr @main.scalar4, align 4
  ret void
}

; Function Attrs: nounwind
define hidden void @main.newArray(ptr %context) unnamed_addr #1 {
entry:
  %new = call ptr @runtime.alloc(i32 3, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new, ptr undef) #2
  store ptr %new, ptr @main.array1, align 4
  %new1 = call ptr @runtime.alloc(i32 71, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new1, ptr undef) #2
  store ptr %new1, ptr @main.array2, align 4
  %new2 = call ptr @runtime.alloc(i32 12, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new2, ptr undef) #2
  store ptr %new2, ptr @main.array3, align 4
  ret void
}

; Function Attrs: nounwind
define hidden void @main.newStruct(ptr %context) unnamed_addr #1 {
entry:
  %new = call ptr @runtime.alloc(i32 0, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new, ptr undef) #2
  store ptr %new, ptr @main.struct1, align 4
  %new1 = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new1, ptr undef) #2
  store ptr %new1, ptr @main.struct2, align 4
  %new2 = call ptr @runtime.alloc(i32 248, ptr nonnull @"runtime/gc.layout:62-2000000000000001", ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new2, ptr undef) #2
  store ptr %new2, ptr @main.struct3, align 4
  %new3 = call ptr @runtime.alloc(i32 248, ptr nonnull @"runtime/gc.layout:62-0001", ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new3, ptr undef) #2
  store ptr %new3, ptr @main.struct4, align 4
  ret void
}

; Function Attrs: nounwind
define hidden ptr @main.newFuncValue(ptr %context) unnamed_addr #1 {
entry:
  %new = call ptr @runtime.alloc(i32 8, ptr nonnull inttoptr (i32 197 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %new, ptr undef) #2
  ret ptr %new
}

; Function Attrs: nounwind
define hidden void @main.makeSlice(ptr %context) unnamed_addr #1 {
entry:
  %makeslice = call ptr @runtime.alloc(i32 5, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %makeslice, ptr undef) #2
  store ptr %makeslice, ptr @main.slice1, align 8
  store i32 5, ptr getelementptr inbounds ({ ptr, i32, i32 }, ptr @main.slice1, i32 0, i32 1), align 4
  store i32 5, ptr getelementptr inbounds ({ ptr, i32, i32 }, ptr @main.slice1, i32 0, i32 2), align 8
  %makeslice1 = call ptr @runtime.alloc(i32 20, ptr nonnull inttoptr (i32 67 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %makeslice1, ptr undef) #2
  store ptr %makeslice1, ptr @main.slice2, align 8
  store i32 5, ptr getelementptr inbounds ({ ptr, i32, i32 }, ptr @main.slice2, i32 0, i32 1), align 4
  store i32 5, ptr getelementptr inbounds ({ ptr, i32, i32 }, ptr @main.slice2, i32 0, i32 2), align 8
  %makeslice3 = call ptr @runtime.alloc(i32 60, ptr nonnull inttoptr (i32 71 to ptr), ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %makeslice3, ptr undef) #2
  store ptr %makeslice3, ptr @main.slice3, align 8
  store i32 5, ptr getelementptr inbounds ({ ptr, i32, i32 }, ptr @main.slice3, i32 0, i32 1), align 4
  store i32 5, ptr getelementptr inbounds ({ ptr, i32, i32 }, ptr @main.slice3, i32 0, i32 2), align 8
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.makeInterface(double %v.r, double %v.i, ptr %context) unnamed_addr #1 {
entry:
  %0 = call ptr @runtime.alloc(i32 16, ptr null, ptr undef) #2
  call void @runtime.trackPointer(ptr nonnull %0, ptr undef) #2
  store double %v.r, ptr %0, align 8
  %.repack1 = getelementptr inbounds { double, double }, ptr %0, i32 0, i32 1
  store double %v.i, ptr %.repack1, align 8
  %1 = insertvalue %runtime._interface { i32 ptrtoint (ptr @"reflect/types.type:basic:complex128" to i32), ptr undef }, ptr %0, 1
  call void @runtime.trackPointer(ptr nonnull %0, ptr undef) #2
  ret %runtime._interface %1
}

attributes #0 = { "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+nontrapping-fptoint,+sign-ext" }
attributes #2 = { nounwind }
