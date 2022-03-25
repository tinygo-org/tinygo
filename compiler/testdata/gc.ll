; ModuleID = 'gc.go'
source_filename = "gc.go"
target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.typecodeID = type { %runtime.typecodeID*, i32, %runtime.interfaceMethodInfo*, %runtime.typecodeID*, i32, %runtime.typecodeID* }
%runtime.interfaceMethodInfo = type { i8*, i32 }
%runtime._interface = type { i32, i8* }

@main.scalar1 = hidden global i8* null, align 4
@main.scalar2 = hidden global i32* null, align 4
@main.scalar3 = hidden global i64* null, align 4
@main.scalar4 = hidden global float* null, align 4
@main.array1 = hidden global [3 x i8]* null, align 4
@main.array2 = hidden global [71 x i8]* null, align 4
@main.array3 = hidden global [3 x i8*]* null, align 4
@main.struct1 = hidden global {}* null, align 4
@main.struct2 = hidden global { i32, i32 }* null, align 4
@main.struct3 = hidden global { i8*, [60 x i32], i8* }* null, align 4
@main.struct4 = hidden global { i8*, [61 x i32] }* null, align 4
@main.slice1 = hidden global { i8*, i32, i32 } zeroinitializer, align 8
@main.slice2 = hidden global { i32**, i32, i32 } zeroinitializer, align 8
@main.slice3 = hidden global { { i8*, i32, i32 }*, i32, i32 } zeroinitializer, align 8
@"runtime/gc.layout:62-2000000000000001" = linkonce_odr unnamed_addr constant { i32, [8 x i8] } { i32 62, [8 x i8] c" \00\00\00\00\00\00\01" }
@"runtime/gc.layout:62-0001" = linkonce_odr unnamed_addr constant { i32, [8 x i8] } { i32 62, [8 x i8] c"\00\00\00\00\00\00\00\01" }
@"reflect/types.type:basic:complex128" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* null, i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* @"reflect/types.type:pointer:basic:complex128", i32 0, %runtime.typecodeID* null }
@"reflect/types.type:pointer:basic:complex128" = linkonce_odr constant %runtime.typecodeID { %runtime.typecodeID* @"reflect/types.type:basic:complex128", i32 0, %runtime.interfaceMethodInfo* null, %runtime.typecodeID* null, i32 0, %runtime.typecodeID* null }

declare noalias nonnull i8* @runtime.alloc(i32, i8*, i8*)

declare void @runtime.trackPointer(i8* nocapture readonly, i8*)

; Function Attrs: nounwind
define hidden void @main.init(i8* %context) unnamed_addr #0 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden void @main.newScalar(i8* %context) unnamed_addr #0 {
entry:
  %new = call i8* @runtime.alloc(i32 1, i8* nonnull inttoptr (i32 3 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new, i8* undef) #0
  store i8* %new, i8** @main.scalar1, align 4
  %new1 = call i8* @runtime.alloc(i32 4, i8* nonnull inttoptr (i32 3 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new1, i8* undef) #0
  store i8* %new1, i8** bitcast (i32** @main.scalar2 to i8**), align 4
  %new2 = call i8* @runtime.alloc(i32 8, i8* nonnull inttoptr (i32 3 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new2, i8* undef) #0
  store i8* %new2, i8** bitcast (i64** @main.scalar3 to i8**), align 4
  %new3 = call i8* @runtime.alloc(i32 4, i8* nonnull inttoptr (i32 3 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new3, i8* undef) #0
  store i8* %new3, i8** bitcast (float** @main.scalar4 to i8**), align 4
  ret void
}

; Function Attrs: nounwind
define hidden void @main.newArray(i8* %context) unnamed_addr #0 {
entry:
  %new = call i8* @runtime.alloc(i32 3, i8* nonnull inttoptr (i32 3 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new, i8* undef) #0
  store i8* %new, i8** bitcast ([3 x i8]** @main.array1 to i8**), align 4
  %new1 = call i8* @runtime.alloc(i32 71, i8* nonnull inttoptr (i32 3 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new1, i8* undef) #0
  store i8* %new1, i8** bitcast ([71 x i8]** @main.array2 to i8**), align 4
  %new2 = call i8* @runtime.alloc(i32 12, i8* nonnull inttoptr (i32 67 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new2, i8* undef) #0
  store i8* %new2, i8** bitcast ([3 x i8*]** @main.array3 to i8**), align 4
  ret void
}

; Function Attrs: nounwind
define hidden void @main.newStruct(i8* %context) unnamed_addr #0 {
entry:
  %new = call i8* @runtime.alloc(i32 0, i8* nonnull inttoptr (i32 3 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new, i8* undef) #0
  store i8* %new, i8** bitcast ({}** @main.struct1 to i8**), align 4
  %new1 = call i8* @runtime.alloc(i32 8, i8* nonnull inttoptr (i32 3 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new1, i8* undef) #0
  store i8* %new1, i8** bitcast ({ i32, i32 }** @main.struct2 to i8**), align 4
  %new2 = call i8* @runtime.alloc(i32 248, i8* bitcast ({ i32, [8 x i8] }* @"runtime/gc.layout:62-2000000000000001" to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new2, i8* undef) #0
  store i8* %new2, i8** bitcast ({ i8*, [60 x i32], i8* }** @main.struct3 to i8**), align 4
  %new3 = call i8* @runtime.alloc(i32 248, i8* bitcast ({ i32, [8 x i8] }* @"runtime/gc.layout:62-0001" to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %new3, i8* undef) #0
  store i8* %new3, i8** bitcast ({ i8*, [61 x i32] }** @main.struct4 to i8**), align 4
  ret void
}

; Function Attrs: nounwind
define hidden { i8*, void ()* }* @main.newFuncValue(i8* %context) unnamed_addr #0 {
entry:
  %new = call i8* @runtime.alloc(i32 8, i8* nonnull inttoptr (i32 197 to i8*), i8* undef) #0
  %0 = bitcast i8* %new to { i8*, void ()* }*
  call void @runtime.trackPointer(i8* nonnull %new, i8* undef) #0
  ret { i8*, void ()* }* %0
}

; Function Attrs: nounwind
define hidden void @main.makeSlice(i8* %context) unnamed_addr #0 {
entry:
  %makeslice = call i8* @runtime.alloc(i32 5, i8* nonnull inttoptr (i32 3 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %makeslice, i8* undef) #0
  store i8* %makeslice, i8** getelementptr inbounds ({ i8*, i32, i32 }, { i8*, i32, i32 }* @main.slice1, i32 0, i32 0), align 8
  store i32 5, i32* getelementptr inbounds ({ i8*, i32, i32 }, { i8*, i32, i32 }* @main.slice1, i32 0, i32 1), align 4
  store i32 5, i32* getelementptr inbounds ({ i8*, i32, i32 }, { i8*, i32, i32 }* @main.slice1, i32 0, i32 2), align 8
  %makeslice1 = call i8* @runtime.alloc(i32 20, i8* nonnull inttoptr (i32 67 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %makeslice1, i8* undef) #0
  store i8* %makeslice1, i8** bitcast ({ i32**, i32, i32 }* @main.slice2 to i8**), align 8
  store i32 5, i32* getelementptr inbounds ({ i32**, i32, i32 }, { i32**, i32, i32 }* @main.slice2, i32 0, i32 1), align 4
  store i32 5, i32* getelementptr inbounds ({ i32**, i32, i32 }, { i32**, i32, i32 }* @main.slice2, i32 0, i32 2), align 8
  %makeslice3 = call i8* @runtime.alloc(i32 60, i8* nonnull inttoptr (i32 71 to i8*), i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %makeslice3, i8* undef) #0
  store i8* %makeslice3, i8** bitcast ({ { i8*, i32, i32 }*, i32, i32 }* @main.slice3 to i8**), align 8
  store i32 5, i32* getelementptr inbounds ({ { i8*, i32, i32 }*, i32, i32 }, { { i8*, i32, i32 }*, i32, i32 }* @main.slice3, i32 0, i32 1), align 4
  store i32 5, i32* getelementptr inbounds ({ { i8*, i32, i32 }*, i32, i32 }, { { i8*, i32, i32 }*, i32, i32 }* @main.slice3, i32 0, i32 2), align 8
  ret void
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.makeInterface(double %v.r, double %v.i, i8* %context) unnamed_addr #0 {
entry:
  %0 = call i8* @runtime.alloc(i32 16, i8* null, i8* undef) #0
  call void @runtime.trackPointer(i8* nonnull %0, i8* undef) #0
  %.repack = bitcast i8* %0 to double*
  store double %v.r, double* %.repack, align 8
  %.repack1 = getelementptr inbounds i8, i8* %0, i32 8
  %1 = bitcast i8* %.repack1 to double*
  store double %v.i, double* %1, align 8
  %2 = insertvalue %runtime._interface { i32 ptrtoint (%runtime.typecodeID* @"reflect/types.type:basic:complex128" to i32), i8* undef }, i8* %0, 1
  call void @runtime.trackPointer(i8* nonnull %0, i8* undef) #0
  ret %runtime._interface %2
}

attributes #0 = { nounwind }
