; ModuleID = 'paramspill.go'
source_filename = "paramspill.go"
target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-i128:128-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

%runtime.structField = type { ptr, ptr }
%main.edge = type { i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32 }
%main.big = type { i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32 }
%runtime._interface = type { ptr, ptr }

@main.sink = hidden global i32 0, align 4
@"reflect/types.signature:main.sum:func:{}{basic:int}" = linkonce_odr constant i8 0, align 1
@"reflect/types.type:named:main.big" = linkonce_odr constant { ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [9 x i8] } { ptr @"named:main.big$methodset", i8 122, i16 -32768, ptr getelementptr ({ ptr, i8, i16, ptr, { i32, [1 x ptr] } }, ptr @"reflect/types.type:pointer:named:main.big", i32 0, i32 1), ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}", ptr @"reflect/types.type.pkgpath:main", { i32, [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/types.signature:main.sum:func:{}{basic:int}"] }, [9 x i8] c"main.big\00" }, align 4
@"reflect/types.type.pkgpath:main" = linkonce_odr unnamed_addr constant [5 x i8] c"main\00", align 1
@"reflect/types.type:pointer:named:main.big" = linkonce_odr constant { ptr, i8, i16, ptr, { i32, [1 x ptr] } } { ptr @"pointer:named:main.big$methodset", i8 -43, i16 -32768, ptr getelementptr ({ ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [9 x i8] }, ptr @"reflect/types.type:named:main.big", i32 0, i32 1), { i32, [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/types.signature:main.sum:func:{}{basic:int}"] } }, align 4
@"main.$methods.sum:func:{}{basic:int}" = linkonce_odr constant i8 0, align 1
@"main$string" = internal unnamed_addr constant [8 x i8] c"main.big", align 1
@"main$string.1" = internal unnamed_addr constant [3 x i8] c"sum", align 1
@"pointer:named:main.big$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"main.$methods.sum:func:{}{basic:int}"], { ptr } { ptr @"(*main.big).sum" } }
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}" = linkonce_odr constant { i8, i16, ptr, ptr, i32, i16, [17 x %runtime.structField] } { i8 90, i16 0, ptr @"reflect/types.type:pointer:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}", ptr @"reflect/types.type.pkgpath:main", i32 68, i16 17, [17 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.a" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.b" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.c" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.d" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.e" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.f" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.g" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.h" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.i" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.j" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.k" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.l" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.m" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.n" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.o" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.p" }, %runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.q" }] }, align 4
@"reflect/types.type:pointer:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}" }, align 4
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.a" = internal unnamed_addr constant [4 x i8] c"\00\00a\00", align 1
@"reflect/types.type:basic:int" = linkonce_odr constant { i8, ptr } { i8 -62, ptr @"reflect/types.type:pointer:basic:int" }, align 4
@"reflect/types.type:pointer:basic:int" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:int" }, align 4
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.b" = internal unnamed_addr constant [4 x i8] c"\00\04b\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.c" = internal unnamed_addr constant [4 x i8] c"\00\08c\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.d" = internal unnamed_addr constant [4 x i8] c"\00\0Cd\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.e" = internal unnamed_addr constant [4 x i8] c"\00\10e\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.f" = internal unnamed_addr constant [4 x i8] c"\00\14f\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.g" = internal unnamed_addr constant [4 x i8] c"\00\18g\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.h" = internal unnamed_addr constant [4 x i8] c"\00\1Ch\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.i" = internal unnamed_addr constant [4 x i8] c"\00 i\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.j" = internal unnamed_addr constant [4 x i8] c"\00$j\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.k" = internal unnamed_addr constant [4 x i8] c"\00(k\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.l" = internal unnamed_addr constant [4 x i8] c"\00,l\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.m" = internal unnamed_addr constant [4 x i8] c"\000m\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.n" = internal unnamed_addr constant [4 x i8] c"\004n\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.o" = internal unnamed_addr constant [4 x i8] c"\008o\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.p" = internal unnamed_addr constant [4 x i8] c"\00<p\00", align 1
@"reflect/types.type:struct:{a:basic:int,b:basic:int,c:basic:int,d:basic:int,e:basic:int,f:basic:int,g:basic:int,h:basic:int,i:basic:int,j:basic:int,k:basic:int,l:basic:int,m:basic:int,n:basic:int,o:basic:int,p:basic:int,q:basic:int}.q" = internal unnamed_addr constant [4 x i8] c"\00@q\00", align 1
@"named:main.big$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"main.$methods.sum:func:{}{basic:int}"], { ptr } { ptr @"(main.big).sum$invoke" } }
@llvm.used = appending global [10 x ptr] [ptr @"(main.big).sum", ptr @sumWithArrayC, ptr @main.takeBig, ptr @main.takeArray, ptr @main.takeWithArray, ptr @takeBigC, ptr @main.spawnBig, ptr @main.callEverything, ptr @main.makeInterface, ptr @main.makeInterfaceWithArray]
@"reflect/types.type:named:main.withArray" = linkonce_odr constant { ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [15 x i8] } { ptr @"named:main.withArray$methodset", i8 122, i16 -32768, ptr getelementptr ({ ptr, i8, i16, ptr, { i32, [1 x ptr] } }, ptr @"reflect/types.type:pointer:named:main.withArray", i32 0, i32 1), ptr @"reflect/types.type:struct:{tag:basic:int,buf:array:16:basic:int32}", ptr @"reflect/types.type.pkgpath:main", { i32, [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/types.signature:main.sum:func:{}{basic:int}"] }, [15 x i8] c"main.withArray\00" }, align 4
@"reflect/types.type:pointer:named:main.withArray" = linkonce_odr constant { ptr, i8, i16, ptr, { i32, [1 x ptr] } } { ptr @"pointer:named:main.withArray$methodset", i8 -43, i16 -32768, ptr getelementptr ({ ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [15 x i8] }, ptr @"reflect/types.type:named:main.withArray", i32 0, i32 1), { i32, [1 x ptr] } { i32 1, [1 x ptr] [ptr @"reflect/types.signature:main.sum:func:{}{basic:int}"] } }, align 4
@"main$string.2" = internal unnamed_addr constant [14 x i8] c"main.withArray", align 1
@"main$string.3" = internal unnamed_addr constant [3 x i8] c"sum", align 1
@"pointer:named:main.withArray$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"main.$methods.sum:func:{}{basic:int}"], { ptr } { ptr @"(*main.withArray).sum" } }
@"reflect/types.type:struct:{tag:basic:int,buf:array:16:basic:int32}" = linkonce_odr constant { i8, i16, ptr, ptr, i32, i16, [2 x %runtime.structField] } { i8 90, i16 0, ptr @"reflect/types.type:pointer:struct:{tag:basic:int,buf:array:16:basic:int32}", ptr @"reflect/types.type.pkgpath:main", i32 68, i16 2, [2 x %runtime.structField] [%runtime.structField { ptr @"reflect/types.type:basic:int", ptr @"reflect/types.type:struct:{tag:basic:int,buf:array:16:basic:int32}.tag" }, %runtime.structField { ptr @"reflect/types.type:array:16:basic:int32", ptr @"reflect/types.type:struct:{tag:basic:int,buf:array:16:basic:int32}.buf" }] }, align 4
@"reflect/types.type:pointer:struct:{tag:basic:int,buf:array:16:basic:int32}" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:struct:{tag:basic:int,buf:array:16:basic:int32}" }, align 4
@"reflect/types.type:struct:{tag:basic:int,buf:array:16:basic:int32}.tag" = internal unnamed_addr constant [6 x i8] c"\00\00tag\00", align 1
@"reflect/types.type:struct:{tag:basic:int,buf:array:16:basic:int32}.buf" = internal unnamed_addr constant [6 x i8] c"\00\04buf\00", align 1
@"reflect/types.type:array:16:basic:int32" = linkonce_odr constant { i8, i16, ptr, ptr, i32, ptr } { i8 -41, i16 0, ptr @"reflect/types.type:pointer:array:16:basic:int32", ptr @"reflect/types.type:basic:int32", i32 16, ptr @"reflect/types.type:slice:basic:int32" }, align 4
@"reflect/types.type:pointer:array:16:basic:int32" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:array:16:basic:int32" }, align 4
@"reflect/types.type:basic:int32" = linkonce_odr constant { i8, ptr } { i8 -59, ptr @"reflect/types.type:pointer:basic:int32" }, align 4
@"reflect/types.type:pointer:basic:int32" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:basic:int32" }, align 4
@"reflect/types.type:slice:basic:int32" = linkonce_odr constant { i8, i16, ptr, ptr } { i8 22, i16 0, ptr @"reflect/types.type:pointer:slice:basic:int32", ptr @"reflect/types.type:basic:int32" }, align 4
@"reflect/types.type:pointer:slice:basic:int32" = linkonce_odr constant { i8, i16, ptr } { i8 -43, i16 0, ptr @"reflect/types.type:slice:basic:int32" }, align 4
@"named:main.withArray$methodset" = linkonce_odr unnamed_addr constant { i32, [1 x ptr], { ptr } } { i32 1, [1 x ptr] [ptr @"main.$methods.sum:func:{}{basic:int}"], { ptr } { ptr @"sumWithArrayC$invoke" } }

declare void @runtime.trackPointer(ptr nocapture readonly, ptr, ptr) #0

; Function Attrs: nounwind
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: nounwind
define hidden i32 @"(main.big).sum"(ptr readonly dereferenceable_or_null(68) %b, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %b1 = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %b1, ptr nonnull %stackalloc, ptr undef) #9
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 4 dereferenceable(68) %b1, ptr noundef nonnull align 1 dereferenceable(68) %b, i32 68, i1 false)
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  br i1 false, label %deref.throw, label %deref.next2

deref.next2:                                      ; preds = %deref.next
  %0 = getelementptr inbounds nuw i8, ptr %b1, i32 64
  %1 = load i32, ptr %b1, align 4
  %2 = load i32, ptr %0, align 4
  %3 = add i32 %1, %2
  ret i32 %3

deref.throw:                                      ; preds = %deref.next, %entry
  unreachable
}

; Function Attrs: allockind("alloc,zeroed") allocsize(0)
declare noalias nonnull ptr @runtime.alloc(i32, ptr, ptr) #2

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: readwrite)
declare void @llvm.memcpy.p0.p0.i32(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i32, i1 immarg) #3

declare void @runtime.nilPanic(ptr) #0

; Function Attrs: nounwind
define i32 @sumWithArrayC(i32 %w.tag, [16 x i32] %w.buf) #4 {
entry:
  %stackalloc = alloca i8, align 1
  %w = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %w, ptr nonnull %stackalloc, ptr undef) #9
  store i32 %w.tag, ptr %w, align 4
  %w.repack1 = getelementptr inbounds nuw i8, ptr %w, i32 4
  %w.buf.elt = extractvalue [16 x i32] %w.buf, 0
  store i32 %w.buf.elt, ptr %w.repack1, align 4
  %w.repack1.repack3 = getelementptr inbounds nuw i8, ptr %w, i32 8
  %w.buf.elt4 = extractvalue [16 x i32] %w.buf, 1
  store i32 %w.buf.elt4, ptr %w.repack1.repack3, align 4
  %w.repack1.repack5 = getelementptr inbounds nuw i8, ptr %w, i32 12
  %w.buf.elt6 = extractvalue [16 x i32] %w.buf, 2
  store i32 %w.buf.elt6, ptr %w.repack1.repack5, align 4
  %w.repack1.repack7 = getelementptr inbounds nuw i8, ptr %w, i32 16
  %w.buf.elt8 = extractvalue [16 x i32] %w.buf, 3
  store i32 %w.buf.elt8, ptr %w.repack1.repack7, align 4
  %w.repack1.repack9 = getelementptr inbounds nuw i8, ptr %w, i32 20
  %w.buf.elt10 = extractvalue [16 x i32] %w.buf, 4
  store i32 %w.buf.elt10, ptr %w.repack1.repack9, align 4
  %w.repack1.repack11 = getelementptr inbounds nuw i8, ptr %w, i32 24
  %w.buf.elt12 = extractvalue [16 x i32] %w.buf, 5
  store i32 %w.buf.elt12, ptr %w.repack1.repack11, align 4
  %w.repack1.repack13 = getelementptr inbounds nuw i8, ptr %w, i32 28
  %w.buf.elt14 = extractvalue [16 x i32] %w.buf, 6
  store i32 %w.buf.elt14, ptr %w.repack1.repack13, align 4
  %w.repack1.repack15 = getelementptr inbounds nuw i8, ptr %w, i32 32
  %w.buf.elt16 = extractvalue [16 x i32] %w.buf, 7
  store i32 %w.buf.elt16, ptr %w.repack1.repack15, align 4
  %w.repack1.repack17 = getelementptr inbounds nuw i8, ptr %w, i32 36
  %w.buf.elt18 = extractvalue [16 x i32] %w.buf, 8
  store i32 %w.buf.elt18, ptr %w.repack1.repack17, align 4
  %w.repack1.repack19 = getelementptr inbounds nuw i8, ptr %w, i32 40
  %w.buf.elt20 = extractvalue [16 x i32] %w.buf, 9
  store i32 %w.buf.elt20, ptr %w.repack1.repack19, align 4
  %w.repack1.repack21 = getelementptr inbounds nuw i8, ptr %w, i32 44
  %w.buf.elt22 = extractvalue [16 x i32] %w.buf, 10
  store i32 %w.buf.elt22, ptr %w.repack1.repack21, align 4
  %w.repack1.repack23 = getelementptr inbounds nuw i8, ptr %w, i32 48
  %w.buf.elt24 = extractvalue [16 x i32] %w.buf, 11
  store i32 %w.buf.elt24, ptr %w.repack1.repack23, align 4
  %w.repack1.repack25 = getelementptr inbounds nuw i8, ptr %w, i32 52
  %w.buf.elt26 = extractvalue [16 x i32] %w.buf, 12
  store i32 %w.buf.elt26, ptr %w.repack1.repack25, align 4
  %w.repack1.repack27 = getelementptr inbounds nuw i8, ptr %w, i32 56
  %w.buf.elt28 = extractvalue [16 x i32] %w.buf, 13
  store i32 %w.buf.elt28, ptr %w.repack1.repack27, align 4
  %w.repack1.repack29 = getelementptr inbounds nuw i8, ptr %w, i32 60
  %w.buf.elt30 = extractvalue [16 x i32] %w.buf, 14
  store i32 %w.buf.elt30, ptr %w.repack1.repack29, align 4
  %w.repack1.repack31 = getelementptr inbounds nuw i8, ptr %w, i32 64
  %w.buf.elt32 = extractvalue [16 x i32] %w.buf, 15
  store i32 %w.buf.elt32, ptr %w.repack1.repack31, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load i32, ptr %w, align 4
  ret i32 %0

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define hidden i32 @main.takeBig(ptr readonly dereferenceable_or_null(68) %b, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %b1 = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %b1, ptr nonnull %stackalloc, ptr undef) #9
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 4 dereferenceable(68) %b1, ptr noundef nonnull align 1 dereferenceable(68) %b, i32 68, i1 false)
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = getelementptr inbounds nuw i8, ptr %b1, i32 64
  %1 = load i32, ptr %0, align 4
  ret i32 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define hidden i32 @main.takeEdge(%main.edge %e, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %e1 = call align 4 dereferenceable(64) ptr @runtime.alloc(i32 64, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %e1, ptr nonnull %stackalloc, ptr undef) #9
  %e.elt = extractvalue %main.edge %e, 0
  store i32 %e.elt, ptr %e1, align 4
  %e1.repack2 = getelementptr inbounds nuw i8, ptr %e1, i32 4
  %e.elt3 = extractvalue %main.edge %e, 1
  store i32 %e.elt3, ptr %e1.repack2, align 4
  %e1.repack4 = getelementptr inbounds nuw i8, ptr %e1, i32 8
  %e.elt5 = extractvalue %main.edge %e, 2
  store i32 %e.elt5, ptr %e1.repack4, align 4
  %e1.repack6 = getelementptr inbounds nuw i8, ptr %e1, i32 12
  %e.elt7 = extractvalue %main.edge %e, 3
  store i32 %e.elt7, ptr %e1.repack6, align 4
  %e1.repack8 = getelementptr inbounds nuw i8, ptr %e1, i32 16
  %e.elt9 = extractvalue %main.edge %e, 4
  store i32 %e.elt9, ptr %e1.repack8, align 4
  %e1.repack10 = getelementptr inbounds nuw i8, ptr %e1, i32 20
  %e.elt11 = extractvalue %main.edge %e, 5
  store i32 %e.elt11, ptr %e1.repack10, align 4
  %e1.repack12 = getelementptr inbounds nuw i8, ptr %e1, i32 24
  %e.elt13 = extractvalue %main.edge %e, 6
  store i32 %e.elt13, ptr %e1.repack12, align 4
  %e1.repack14 = getelementptr inbounds nuw i8, ptr %e1, i32 28
  %e.elt15 = extractvalue %main.edge %e, 7
  store i32 %e.elt15, ptr %e1.repack14, align 4
  %e1.repack16 = getelementptr inbounds nuw i8, ptr %e1, i32 32
  %e.elt17 = extractvalue %main.edge %e, 8
  store i32 %e.elt17, ptr %e1.repack16, align 4
  %e1.repack18 = getelementptr inbounds nuw i8, ptr %e1, i32 36
  %e.elt19 = extractvalue %main.edge %e, 9
  store i32 %e.elt19, ptr %e1.repack18, align 4
  %e1.repack20 = getelementptr inbounds nuw i8, ptr %e1, i32 40
  %e.elt21 = extractvalue %main.edge %e, 10
  store i32 %e.elt21, ptr %e1.repack20, align 4
  %e1.repack22 = getelementptr inbounds nuw i8, ptr %e1, i32 44
  %e.elt23 = extractvalue %main.edge %e, 11
  store i32 %e.elt23, ptr %e1.repack22, align 4
  %e1.repack24 = getelementptr inbounds nuw i8, ptr %e1, i32 48
  %e.elt25 = extractvalue %main.edge %e, 12
  store i32 %e.elt25, ptr %e1.repack24, align 4
  %e1.repack26 = getelementptr inbounds nuw i8, ptr %e1, i32 52
  %e.elt27 = extractvalue %main.edge %e, 13
  store i32 %e.elt27, ptr %e1.repack26, align 4
  %e1.repack28 = getelementptr inbounds nuw i8, ptr %e1, i32 56
  %e.elt29 = extractvalue %main.edge %e, 14
  store i32 %e.elt29, ptr %e1.repack28, align 4
  %e1.repack30 = getelementptr inbounds nuw i8, ptr %e1, i32 60
  %e.elt31 = extractvalue %main.edge %e, 15
  store i32 %e.elt31, ptr %e1.repack30, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = getelementptr inbounds nuw i8, ptr %e1, i32 60
  %1 = load i32, ptr %0, align 4
  ret i32 %1

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define hidden i32 @main.takeArray(ptr readonly dereferenceable_or_null(68) %a, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %a1 = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %a1, ptr nonnull %stackalloc, ptr undef) #9
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 4 dereferenceable(68) %a1, ptr noundef nonnull align 1 dereferenceable(68) %a, i32 68, i1 false)
  %0 = getelementptr inbounds nuw i8, ptr %a1, i32 64
  %1 = load i32, ptr %0, align 4
  ret i32 %1
}

; Function Attrs: nounwind
define hidden i32 @main.takeWithArray(ptr readonly dereferenceable_or_null(68) %w, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %w1 = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %w1, ptr nonnull %stackalloc, ptr undef) #9
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 4 dereferenceable(68) %w1, ptr noundef nonnull align 1 dereferenceable(68) %w, i32 68, i1 false)
  br i1 false, label %gep.throw, label %gep.next

gep.next:                                         ; preds = %entry
  %0 = getelementptr inbounds nuw i8, ptr %w1, i32 4
  %1 = load i32, ptr %0, align 4
  ret i32 %1

gep.throw:                                        ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define i32 @takeBigC(%main.big %b) #5 {
entry:
  %stackalloc = alloca i8, align 1
  %b1 = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %b1, ptr nonnull %stackalloc, ptr undef) #9
  %b.elt = extractvalue %main.big %b, 0
  store i32 %b.elt, ptr %b1, align 4
  %b1.repack2 = getelementptr inbounds nuw i8, ptr %b1, i32 4
  %b.elt3 = extractvalue %main.big %b, 1
  store i32 %b.elt3, ptr %b1.repack2, align 4
  %b1.repack4 = getelementptr inbounds nuw i8, ptr %b1, i32 8
  %b.elt5 = extractvalue %main.big %b, 2
  store i32 %b.elt5, ptr %b1.repack4, align 4
  %b1.repack6 = getelementptr inbounds nuw i8, ptr %b1, i32 12
  %b.elt7 = extractvalue %main.big %b, 3
  store i32 %b.elt7, ptr %b1.repack6, align 4
  %b1.repack8 = getelementptr inbounds nuw i8, ptr %b1, i32 16
  %b.elt9 = extractvalue %main.big %b, 4
  store i32 %b.elt9, ptr %b1.repack8, align 4
  %b1.repack10 = getelementptr inbounds nuw i8, ptr %b1, i32 20
  %b.elt11 = extractvalue %main.big %b, 5
  store i32 %b.elt11, ptr %b1.repack10, align 4
  %b1.repack12 = getelementptr inbounds nuw i8, ptr %b1, i32 24
  %b.elt13 = extractvalue %main.big %b, 6
  store i32 %b.elt13, ptr %b1.repack12, align 4
  %b1.repack14 = getelementptr inbounds nuw i8, ptr %b1, i32 28
  %b.elt15 = extractvalue %main.big %b, 7
  store i32 %b.elt15, ptr %b1.repack14, align 4
  %b1.repack16 = getelementptr inbounds nuw i8, ptr %b1, i32 32
  %b.elt17 = extractvalue %main.big %b, 8
  store i32 %b.elt17, ptr %b1.repack16, align 4
  %b1.repack18 = getelementptr inbounds nuw i8, ptr %b1, i32 36
  %b.elt19 = extractvalue %main.big %b, 9
  store i32 %b.elt19, ptr %b1.repack18, align 4
  %b1.repack20 = getelementptr inbounds nuw i8, ptr %b1, i32 40
  %b.elt21 = extractvalue %main.big %b, 10
  store i32 %b.elt21, ptr %b1.repack20, align 4
  %b1.repack22 = getelementptr inbounds nuw i8, ptr %b1, i32 44
  %b.elt23 = extractvalue %main.big %b, 11
  store i32 %b.elt23, ptr %b1.repack22, align 4
  %b1.repack24 = getelementptr inbounds nuw i8, ptr %b1, i32 48
  %b.elt25 = extractvalue %main.big %b, 12
  store i32 %b.elt25, ptr %b1.repack24, align 4
  %b1.repack26 = getelementptr inbounds nuw i8, ptr %b1, i32 52
  %b.elt27 = extractvalue %main.big %b, 13
  store i32 %b.elt27, ptr %b1.repack26, align 4
  %b1.repack28 = getelementptr inbounds nuw i8, ptr %b1, i32 56
  %b.elt29 = extractvalue %main.big %b, 14
  store i32 %b.elt29, ptr %b1.repack28, align 4
  %b1.repack30 = getelementptr inbounds nuw i8, ptr %b1, i32 60
  %b.elt31 = extractvalue %main.big %b, 15
  store i32 %b.elt31, ptr %b1.repack30, align 4
  %b1.repack32 = getelementptr inbounds nuw i8, ptr %b1, i32 64
  %b.elt33 = extractvalue %main.big %b, 16
  store i32 %b.elt33, ptr %b1.repack32, align 4
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load i32, ptr %b1, align 4
  ret i32 %0

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: nounwind
define hidden void @main.spawnBig(ptr readonly dereferenceable_or_null(68) %b, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %b1 = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %b1, ptr nonnull %stackalloc, ptr undef) #9
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 4 dereferenceable(68) %b1, ptr noundef nonnull align 1 dereferenceable(68) %b, i32 68, i1 false)
  br i1 false, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %0 = load i32, ptr %b1, align 4
  store i32 %0, ptr @main.sink, align 4
  ret void

deref.throw:                                      ; preds = %entry
  unreachable
}

; Function Attrs: noinline nounwind
define hidden { ptr, ptr } @main.pickTakeBig(ptr %context) unnamed_addr #6 {
entry:
  ret { ptr, ptr } { ptr undef, ptr @main.takeBig }
}

; Function Attrs: nounwind
define hidden i32 @main.callEverything(ptr readonly dereferenceable_or_null(68) %b, %main.edge %e, ptr readonly dereferenceable_or_null(68) %a, ptr readonly dereferenceable_or_null(68) %w, ptr %s.typecode, ptr %s.value, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %0 = call i32 @main.takeBig(ptr %b, ptr undef)
  %1 = call i32 @main.takeEdge(%main.edge %e, ptr undef)
  %2 = call i32 @main.takeArray(ptr %a, ptr undef)
  %3 = call i32 @main.takeWithArray(ptr %w, ptr undef)
  %.unpack = load i32, ptr %b, align 4
  %4 = insertvalue %main.big poison, i32 %.unpack, 0
  %.elt1 = getelementptr inbounds nuw i8, ptr %b, i32 4
  %.unpack2 = load i32, ptr %.elt1, align 4
  %5 = insertvalue %main.big %4, i32 %.unpack2, 1
  %.elt3 = getelementptr inbounds nuw i8, ptr %b, i32 8
  %.unpack4 = load i32, ptr %.elt3, align 4
  %6 = insertvalue %main.big %5, i32 %.unpack4, 2
  %.elt5 = getelementptr inbounds nuw i8, ptr %b, i32 12
  %.unpack6 = load i32, ptr %.elt5, align 4
  %7 = insertvalue %main.big %6, i32 %.unpack6, 3
  %.elt7 = getelementptr inbounds nuw i8, ptr %b, i32 16
  %.unpack8 = load i32, ptr %.elt7, align 4
  %8 = insertvalue %main.big %7, i32 %.unpack8, 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %b, i32 20
  %.unpack10 = load i32, ptr %.elt9, align 4
  %9 = insertvalue %main.big %8, i32 %.unpack10, 5
  %.elt11 = getelementptr inbounds nuw i8, ptr %b, i32 24
  %.unpack12 = load i32, ptr %.elt11, align 4
  %10 = insertvalue %main.big %9, i32 %.unpack12, 6
  %.elt13 = getelementptr inbounds nuw i8, ptr %b, i32 28
  %.unpack14 = load i32, ptr %.elt13, align 4
  %11 = insertvalue %main.big %10, i32 %.unpack14, 7
  %.elt15 = getelementptr inbounds nuw i8, ptr %b, i32 32
  %.unpack16 = load i32, ptr %.elt15, align 4
  %12 = insertvalue %main.big %11, i32 %.unpack16, 8
  %.elt17 = getelementptr inbounds nuw i8, ptr %b, i32 36
  %.unpack18 = load i32, ptr %.elt17, align 4
  %13 = insertvalue %main.big %12, i32 %.unpack18, 9
  %.elt19 = getelementptr inbounds nuw i8, ptr %b, i32 40
  %.unpack20 = load i32, ptr %.elt19, align 4
  %14 = insertvalue %main.big %13, i32 %.unpack20, 10
  %.elt21 = getelementptr inbounds nuw i8, ptr %b, i32 44
  %.unpack22 = load i32, ptr %.elt21, align 4
  %15 = insertvalue %main.big %14, i32 %.unpack22, 11
  %.elt23 = getelementptr inbounds nuw i8, ptr %b, i32 48
  %.unpack24 = load i32, ptr %.elt23, align 4
  %16 = insertvalue %main.big %15, i32 %.unpack24, 12
  %.elt25 = getelementptr inbounds nuw i8, ptr %b, i32 52
  %.unpack26 = load i32, ptr %.elt25, align 4
  %17 = insertvalue %main.big %16, i32 %.unpack26, 13
  %.elt27 = getelementptr inbounds nuw i8, ptr %b, i32 56
  %.unpack28 = load i32, ptr %.elt27, align 4
  %18 = insertvalue %main.big %17, i32 %.unpack28, 14
  %.elt29 = getelementptr inbounds nuw i8, ptr %b, i32 60
  %.unpack30 = load i32, ptr %.elt29, align 4
  %19 = insertvalue %main.big %18, i32 %.unpack30, 15
  %.elt31 = getelementptr inbounds nuw i8, ptr %b, i32 64
  %.unpack32 = load i32, ptr %.elt31, align 4
  %20 = insertvalue %main.big %19, i32 %.unpack32, 16
  %21 = call i32 @takeBigC(%main.big %20)
  %22 = call { ptr, ptr } @main.pickTakeBig(ptr undef)
  %23 = extractvalue { ptr, ptr } %22, 0
  call void @runtime.trackPointer(ptr %23, ptr nonnull %stackalloc, ptr undef) #9
  %24 = extractvalue { ptr, ptr } %22, 1
  call void @runtime.trackPointer(ptr %24, ptr nonnull %stackalloc, ptr undef) #9
  %25 = extractvalue { ptr, ptr } %22, 1
  %26 = icmp eq ptr %25, null
  br i1 %26, label %fpcall.throw, label %fpcall.next

fpcall.next:                                      ; preds = %entry
  %27 = extractvalue { ptr, ptr } %22, 0
  %28 = add i32 %0, %1
  %29 = add i32 %28, %2
  %30 = add i32 %29, %3
  %31 = add i32 %30, %21
  %32 = call i32 %25(ptr nonnull %b, ptr %27) #9
  %33 = add i32 %31, %32
  %34 = call i32 @"interface:{main.sum:func:{}{basic:int}}.sum$invoke"(ptr %s.value, ptr %s.typecode, ptr undef) #9
  %35 = add i32 %33, %34
  %go.param = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %go.param, ptr nonnull %stackalloc, ptr undef) #9
  call void @llvm.memcpy.p0.p0.i32(ptr noundef nonnull align 4 dereferenceable(68) %go.param, ptr noundef nonnull align 1 dereferenceable(68) %b, i32 68, i1 false)
  call void @"internal/task.start"(i32 ptrtoint (ptr @"main.spawnBig$gowrapper" to i32), ptr nonnull %go.param, i32 65536, ptr undef) #9
  ret i32 %35

fpcall.throw:                                     ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #9
  unreachable
}

declare i32 @"interface:{main.sum:func:{}{basic:int}}.sum$invoke"(ptr, ptr, ptr) #7

declare void @runtime.exitGoroutine(ptr) #0

; Function Attrs: nounwind
define linkonce_odr void @"main.spawnBig$gowrapper"(ptr %0) unnamed_addr #8 {
entry:
  call void @main.spawnBig(ptr %0, ptr undef)
  call void @runtime.exitGoroutine(ptr undef) #9
  unreachable
}

declare void @"internal/task.start"(i32, ptr, i32, ptr) #0

; Function Attrs: nounwind
define hidden %runtime._interface @main.makeInterface(ptr readonly dereferenceable_or_null(68) %b, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %.unpack = load i32, ptr %b, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %b, i32 4
  %.unpack2 = load i32, ptr %.elt1, align 4
  %.elt3 = getelementptr inbounds nuw i8, ptr %b, i32 8
  %.unpack4 = load i32, ptr %.elt3, align 4
  %.elt5 = getelementptr inbounds nuw i8, ptr %b, i32 12
  %.unpack6 = load i32, ptr %.elt5, align 4
  %.elt7 = getelementptr inbounds nuw i8, ptr %b, i32 16
  %.unpack8 = load i32, ptr %.elt7, align 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %b, i32 20
  %.unpack10 = load i32, ptr %.elt9, align 4
  %.elt11 = getelementptr inbounds nuw i8, ptr %b, i32 24
  %.unpack12 = load i32, ptr %.elt11, align 4
  %.elt13 = getelementptr inbounds nuw i8, ptr %b, i32 28
  %.unpack14 = load i32, ptr %.elt13, align 4
  %.elt15 = getelementptr inbounds nuw i8, ptr %b, i32 32
  %.unpack16 = load i32, ptr %.elt15, align 4
  %.elt17 = getelementptr inbounds nuw i8, ptr %b, i32 36
  %.unpack18 = load i32, ptr %.elt17, align 4
  %.elt19 = getelementptr inbounds nuw i8, ptr %b, i32 40
  %.unpack20 = load i32, ptr %.elt19, align 4
  %.elt21 = getelementptr inbounds nuw i8, ptr %b, i32 44
  %.unpack22 = load i32, ptr %.elt21, align 4
  %.elt23 = getelementptr inbounds nuw i8, ptr %b, i32 48
  %.unpack24 = load i32, ptr %.elt23, align 4
  %.elt25 = getelementptr inbounds nuw i8, ptr %b, i32 52
  %.unpack26 = load i32, ptr %.elt25, align 4
  %.elt27 = getelementptr inbounds nuw i8, ptr %b, i32 56
  %.unpack28 = load i32, ptr %.elt27, align 4
  %.elt29 = getelementptr inbounds nuw i8, ptr %b, i32 60
  %.unpack30 = load i32, ptr %.elt29, align 4
  %.elt31 = getelementptr inbounds nuw i8, ptr %b, i32 64
  %.unpack32 = load i32, ptr %.elt31, align 4
  %0 = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #9
  store i32 %.unpack, ptr %0, align 4
  %.repack33 = getelementptr inbounds nuw i8, ptr %0, i32 4
  store i32 %.unpack2, ptr %.repack33, align 4
  %.repack35 = getelementptr inbounds nuw i8, ptr %0, i32 8
  store i32 %.unpack4, ptr %.repack35, align 4
  %.repack37 = getelementptr inbounds nuw i8, ptr %0, i32 12
  store i32 %.unpack6, ptr %.repack37, align 4
  %.repack39 = getelementptr inbounds nuw i8, ptr %0, i32 16
  store i32 %.unpack8, ptr %.repack39, align 4
  %.repack41 = getelementptr inbounds nuw i8, ptr %0, i32 20
  store i32 %.unpack10, ptr %.repack41, align 4
  %.repack43 = getelementptr inbounds nuw i8, ptr %0, i32 24
  store i32 %.unpack12, ptr %.repack43, align 4
  %.repack45 = getelementptr inbounds nuw i8, ptr %0, i32 28
  store i32 %.unpack14, ptr %.repack45, align 4
  %.repack47 = getelementptr inbounds nuw i8, ptr %0, i32 32
  store i32 %.unpack16, ptr %.repack47, align 4
  %.repack49 = getelementptr inbounds nuw i8, ptr %0, i32 36
  store i32 %.unpack18, ptr %.repack49, align 4
  %.repack51 = getelementptr inbounds nuw i8, ptr %0, i32 40
  store i32 %.unpack20, ptr %.repack51, align 4
  %.repack53 = getelementptr inbounds nuw i8, ptr %0, i32 44
  store i32 %.unpack22, ptr %.repack53, align 4
  %.repack55 = getelementptr inbounds nuw i8, ptr %0, i32 48
  store i32 %.unpack24, ptr %.repack55, align 4
  %.repack57 = getelementptr inbounds nuw i8, ptr %0, i32 52
  store i32 %.unpack26, ptr %.repack57, align 4
  %.repack59 = getelementptr inbounds nuw i8, ptr %0, i32 56
  store i32 %.unpack28, ptr %.repack59, align 4
  %.repack61 = getelementptr inbounds nuw i8, ptr %0, i32 60
  store i32 %.unpack30, ptr %.repack61, align 4
  %.repack63 = getelementptr inbounds nuw i8, ptr %0, i32 64
  store i32 %.unpack32, ptr %.repack63, align 4
  %1 = insertvalue %runtime._interface { ptr getelementptr ({ ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [9 x i8] }, ptr @"reflect/types.type:named:main.big", i32 0, i32 1), ptr undef }, ptr %0, 1
  call void @runtime.trackPointer(ptr getelementptr ({ ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [9 x i8] }, ptr @"reflect/types.type:named:main.big", i32 0, i32 1), ptr nonnull %stackalloc, ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #9
  ret %runtime._interface %1
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(*main.big).sum"(ptr dereferenceable_or_null(68) %b, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %b, ptr nonnull %stackalloc, ptr undef) #9
  %0 = icmp eq ptr %b, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load i32, ptr %b, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %b, i32 4
  %.unpack2 = load i32, ptr %.elt1, align 4
  %.elt3 = getelementptr inbounds nuw i8, ptr %b, i32 8
  %.unpack4 = load i32, ptr %.elt3, align 4
  %.elt5 = getelementptr inbounds nuw i8, ptr %b, i32 12
  %.unpack6 = load i32, ptr %.elt5, align 4
  %.elt7 = getelementptr inbounds nuw i8, ptr %b, i32 16
  %.unpack8 = load i32, ptr %.elt7, align 4
  %.elt9 = getelementptr inbounds nuw i8, ptr %b, i32 20
  %.unpack10 = load i32, ptr %.elt9, align 4
  %.elt11 = getelementptr inbounds nuw i8, ptr %b, i32 24
  %.unpack12 = load i32, ptr %.elt11, align 4
  %.elt13 = getelementptr inbounds nuw i8, ptr %b, i32 28
  %.unpack14 = load i32, ptr %.elt13, align 4
  %.elt15 = getelementptr inbounds nuw i8, ptr %b, i32 32
  %.unpack16 = load i32, ptr %.elt15, align 4
  %.elt17 = getelementptr inbounds nuw i8, ptr %b, i32 36
  %.unpack18 = load i32, ptr %.elt17, align 4
  %.elt19 = getelementptr inbounds nuw i8, ptr %b, i32 40
  %.unpack20 = load i32, ptr %.elt19, align 4
  %.elt21 = getelementptr inbounds nuw i8, ptr %b, i32 44
  %.unpack22 = load i32, ptr %.elt21, align 4
  %.elt23 = getelementptr inbounds nuw i8, ptr %b, i32 48
  %.unpack24 = load i32, ptr %.elt23, align 4
  %.elt25 = getelementptr inbounds nuw i8, ptr %b, i32 52
  %.unpack26 = load i32, ptr %.elt25, align 4
  %.elt27 = getelementptr inbounds nuw i8, ptr %b, i32 56
  %.unpack28 = load i32, ptr %.elt27, align 4
  %.elt29 = getelementptr inbounds nuw i8, ptr %b, i32 60
  %.unpack30 = load i32, ptr %.elt29, align 4
  %.elt31 = getelementptr inbounds nuw i8, ptr %b, i32 64
  %.unpack32 = load i32, ptr %.elt31, align 4
  %t1 = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %t1, ptr nonnull %stackalloc, ptr undef) #9
  store i32 %.unpack, ptr %t1, align 4
  %t1.repack33 = getelementptr inbounds nuw i8, ptr %t1, i32 4
  store i32 %.unpack2, ptr %t1.repack33, align 4
  %t1.repack35 = getelementptr inbounds nuw i8, ptr %t1, i32 8
  store i32 %.unpack4, ptr %t1.repack35, align 4
  %t1.repack37 = getelementptr inbounds nuw i8, ptr %t1, i32 12
  store i32 %.unpack6, ptr %t1.repack37, align 4
  %t1.repack39 = getelementptr inbounds nuw i8, ptr %t1, i32 16
  store i32 %.unpack8, ptr %t1.repack39, align 4
  %t1.repack41 = getelementptr inbounds nuw i8, ptr %t1, i32 20
  store i32 %.unpack10, ptr %t1.repack41, align 4
  %t1.repack43 = getelementptr inbounds nuw i8, ptr %t1, i32 24
  store i32 %.unpack12, ptr %t1.repack43, align 4
  %t1.repack45 = getelementptr inbounds nuw i8, ptr %t1, i32 28
  store i32 %.unpack14, ptr %t1.repack45, align 4
  %t1.repack47 = getelementptr inbounds nuw i8, ptr %t1, i32 32
  store i32 %.unpack16, ptr %t1.repack47, align 4
  %t1.repack49 = getelementptr inbounds nuw i8, ptr %t1, i32 36
  store i32 %.unpack18, ptr %t1.repack49, align 4
  %t1.repack51 = getelementptr inbounds nuw i8, ptr %t1, i32 40
  store i32 %.unpack20, ptr %t1.repack51, align 4
  %t1.repack53 = getelementptr inbounds nuw i8, ptr %t1, i32 44
  store i32 %.unpack22, ptr %t1.repack53, align 4
  %t1.repack55 = getelementptr inbounds nuw i8, ptr %t1, i32 48
  store i32 %.unpack24, ptr %t1.repack55, align 4
  %t1.repack57 = getelementptr inbounds nuw i8, ptr %t1, i32 52
  store i32 %.unpack26, ptr %t1.repack57, align 4
  %t1.repack59 = getelementptr inbounds nuw i8, ptr %t1, i32 56
  store i32 %.unpack28, ptr %t1.repack59, align 4
  %t1.repack61 = getelementptr inbounds nuw i8, ptr %t1, i32 60
  store i32 %.unpack30, ptr %t1.repack61, align 4
  %t1.repack63 = getelementptr inbounds nuw i8, ptr %t1, i32 64
  store i32 %.unpack32, ptr %t1.repack63, align 4
  %1 = call i32 @"(main.big).sum"(ptr nonnull %t1, ptr undef)
  ret i32 %1

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #9
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i32 @"(main.big).sum$invoke"(ptr %0, ptr %1) unnamed_addr #1 {
entry:
  %ret = call i32 @"(main.big).sum"(ptr %0, ptr %1)
  ret i32 %ret
}

; Function Attrs: nounwind
define hidden %runtime._interface @main.makeInterfaceWithArray(ptr readonly dereferenceable_or_null(68) %w, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  %.unpack = load i32, ptr %w, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %w, i32 4
  %.unpack2.unpack = load i32, ptr %.elt1, align 4
  %.unpack2.elt3 = getelementptr inbounds nuw i8, ptr %w, i32 8
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %.unpack2.elt5 = getelementptr inbounds nuw i8, ptr %w, i32 12
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %.unpack2.elt7 = getelementptr inbounds nuw i8, ptr %w, i32 16
  %.unpack2.unpack8 = load i32, ptr %.unpack2.elt7, align 4
  %.unpack2.elt9 = getelementptr inbounds nuw i8, ptr %w, i32 20
  %.unpack2.unpack10 = load i32, ptr %.unpack2.elt9, align 4
  %.unpack2.elt11 = getelementptr inbounds nuw i8, ptr %w, i32 24
  %.unpack2.unpack12 = load i32, ptr %.unpack2.elt11, align 4
  %.unpack2.elt13 = getelementptr inbounds nuw i8, ptr %w, i32 28
  %.unpack2.unpack14 = load i32, ptr %.unpack2.elt13, align 4
  %.unpack2.elt15 = getelementptr inbounds nuw i8, ptr %w, i32 32
  %.unpack2.unpack16 = load i32, ptr %.unpack2.elt15, align 4
  %.unpack2.elt17 = getelementptr inbounds nuw i8, ptr %w, i32 36
  %.unpack2.unpack18 = load i32, ptr %.unpack2.elt17, align 4
  %.unpack2.elt19 = getelementptr inbounds nuw i8, ptr %w, i32 40
  %.unpack2.unpack20 = load i32, ptr %.unpack2.elt19, align 4
  %.unpack2.elt21 = getelementptr inbounds nuw i8, ptr %w, i32 44
  %.unpack2.unpack22 = load i32, ptr %.unpack2.elt21, align 4
  %.unpack2.elt23 = getelementptr inbounds nuw i8, ptr %w, i32 48
  %.unpack2.unpack24 = load i32, ptr %.unpack2.elt23, align 4
  %.unpack2.elt25 = getelementptr inbounds nuw i8, ptr %w, i32 52
  %.unpack2.unpack26 = load i32, ptr %.unpack2.elt25, align 4
  %.unpack2.elt27 = getelementptr inbounds nuw i8, ptr %w, i32 56
  %.unpack2.unpack28 = load i32, ptr %.unpack2.elt27, align 4
  %.unpack2.elt29 = getelementptr inbounds nuw i8, ptr %w, i32 60
  %.unpack2.unpack30 = load i32, ptr %.unpack2.elt29, align 4
  %.unpack2.elt31 = getelementptr inbounds nuw i8, ptr %w, i32 64
  %.unpack2.unpack32 = load i32, ptr %.unpack2.elt31, align 4
  %0 = call align 4 dereferenceable(68) ptr @runtime.alloc(i32 68, ptr nonnull inttoptr (i32 3 to ptr), ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #9
  store i32 %.unpack, ptr %0, align 4
  %.repack34 = getelementptr inbounds nuw i8, ptr %0, i32 4
  store i32 %.unpack2.unpack, ptr %.repack34, align 4
  %.repack34.repack36 = getelementptr inbounds nuw i8, ptr %0, i32 8
  store i32 %.unpack2.unpack4, ptr %.repack34.repack36, align 4
  %.repack34.repack38 = getelementptr inbounds nuw i8, ptr %0, i32 12
  store i32 %.unpack2.unpack6, ptr %.repack34.repack38, align 4
  %.repack34.repack40 = getelementptr inbounds nuw i8, ptr %0, i32 16
  store i32 %.unpack2.unpack8, ptr %.repack34.repack40, align 4
  %.repack34.repack42 = getelementptr inbounds nuw i8, ptr %0, i32 20
  store i32 %.unpack2.unpack10, ptr %.repack34.repack42, align 4
  %.repack34.repack44 = getelementptr inbounds nuw i8, ptr %0, i32 24
  store i32 %.unpack2.unpack12, ptr %.repack34.repack44, align 4
  %.repack34.repack46 = getelementptr inbounds nuw i8, ptr %0, i32 28
  store i32 %.unpack2.unpack14, ptr %.repack34.repack46, align 4
  %.repack34.repack48 = getelementptr inbounds nuw i8, ptr %0, i32 32
  store i32 %.unpack2.unpack16, ptr %.repack34.repack48, align 4
  %.repack34.repack50 = getelementptr inbounds nuw i8, ptr %0, i32 36
  store i32 %.unpack2.unpack18, ptr %.repack34.repack50, align 4
  %.repack34.repack52 = getelementptr inbounds nuw i8, ptr %0, i32 40
  store i32 %.unpack2.unpack20, ptr %.repack34.repack52, align 4
  %.repack34.repack54 = getelementptr inbounds nuw i8, ptr %0, i32 44
  store i32 %.unpack2.unpack22, ptr %.repack34.repack54, align 4
  %.repack34.repack56 = getelementptr inbounds nuw i8, ptr %0, i32 48
  store i32 %.unpack2.unpack24, ptr %.repack34.repack56, align 4
  %.repack34.repack58 = getelementptr inbounds nuw i8, ptr %0, i32 52
  store i32 %.unpack2.unpack26, ptr %.repack34.repack58, align 4
  %.repack34.repack60 = getelementptr inbounds nuw i8, ptr %0, i32 56
  store i32 %.unpack2.unpack28, ptr %.repack34.repack60, align 4
  %.repack34.repack62 = getelementptr inbounds nuw i8, ptr %0, i32 60
  store i32 %.unpack2.unpack30, ptr %.repack34.repack62, align 4
  %.repack34.repack64 = getelementptr inbounds nuw i8, ptr %0, i32 64
  store i32 %.unpack2.unpack32, ptr %.repack34.repack64, align 4
  %1 = insertvalue %runtime._interface { ptr getelementptr ({ ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [15 x i8] }, ptr @"reflect/types.type:named:main.withArray", i32 0, i32 1), ptr undef }, ptr %0, 1
  call void @runtime.trackPointer(ptr getelementptr ({ ptr, i8, i16, ptr, ptr, ptr, { i32, [1 x ptr] }, [15 x i8] }, ptr @"reflect/types.type:named:main.withArray", i32 0, i32 1), ptr nonnull %stackalloc, ptr undef) #9
  call void @runtime.trackPointer(ptr nonnull %0, ptr nonnull %stackalloc, ptr undef) #9
  ret %runtime._interface %1
}

; Function Attrs: nounwind
define linkonce_odr hidden i32 @"(*main.withArray).sum"(ptr dereferenceable_or_null(68) %w, ptr %context) unnamed_addr #1 {
entry:
  %stackalloc = alloca i8, align 1
  call void @runtime.trackPointer(ptr %w, ptr nonnull %stackalloc, ptr undef) #9
  %0 = icmp eq ptr %w, null
  br i1 %0, label %deref.throw, label %deref.next

deref.next:                                       ; preds = %entry
  %.unpack = load i32, ptr %w, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %w, i32 4
  %.unpack2.unpack = load i32, ptr %.elt1, align 4
  %1 = insertvalue [16 x i32] poison, i32 %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds nuw i8, ptr %w, i32 8
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %2 = insertvalue [16 x i32] %1, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds nuw i8, ptr %w, i32 12
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %3 = insertvalue [16 x i32] %2, i32 %.unpack2.unpack6, 2
  %.unpack2.elt7 = getelementptr inbounds nuw i8, ptr %w, i32 16
  %.unpack2.unpack8 = load i32, ptr %.unpack2.elt7, align 4
  %4 = insertvalue [16 x i32] %3, i32 %.unpack2.unpack8, 3
  %.unpack2.elt9 = getelementptr inbounds nuw i8, ptr %w, i32 20
  %.unpack2.unpack10 = load i32, ptr %.unpack2.elt9, align 4
  %5 = insertvalue [16 x i32] %4, i32 %.unpack2.unpack10, 4
  %.unpack2.elt11 = getelementptr inbounds nuw i8, ptr %w, i32 24
  %.unpack2.unpack12 = load i32, ptr %.unpack2.elt11, align 4
  %6 = insertvalue [16 x i32] %5, i32 %.unpack2.unpack12, 5
  %.unpack2.elt13 = getelementptr inbounds nuw i8, ptr %w, i32 28
  %.unpack2.unpack14 = load i32, ptr %.unpack2.elt13, align 4
  %7 = insertvalue [16 x i32] %6, i32 %.unpack2.unpack14, 6
  %.unpack2.elt15 = getelementptr inbounds nuw i8, ptr %w, i32 32
  %.unpack2.unpack16 = load i32, ptr %.unpack2.elt15, align 4
  %8 = insertvalue [16 x i32] %7, i32 %.unpack2.unpack16, 7
  %.unpack2.elt17 = getelementptr inbounds nuw i8, ptr %w, i32 36
  %.unpack2.unpack18 = load i32, ptr %.unpack2.elt17, align 4
  %9 = insertvalue [16 x i32] %8, i32 %.unpack2.unpack18, 8
  %.unpack2.elt19 = getelementptr inbounds nuw i8, ptr %w, i32 40
  %.unpack2.unpack20 = load i32, ptr %.unpack2.elt19, align 4
  %10 = insertvalue [16 x i32] %9, i32 %.unpack2.unpack20, 9
  %.unpack2.elt21 = getelementptr inbounds nuw i8, ptr %w, i32 44
  %.unpack2.unpack22 = load i32, ptr %.unpack2.elt21, align 4
  %11 = insertvalue [16 x i32] %10, i32 %.unpack2.unpack22, 10
  %.unpack2.elt23 = getelementptr inbounds nuw i8, ptr %w, i32 48
  %.unpack2.unpack24 = load i32, ptr %.unpack2.elt23, align 4
  %12 = insertvalue [16 x i32] %11, i32 %.unpack2.unpack24, 11
  %.unpack2.elt25 = getelementptr inbounds nuw i8, ptr %w, i32 52
  %.unpack2.unpack26 = load i32, ptr %.unpack2.elt25, align 4
  %13 = insertvalue [16 x i32] %12, i32 %.unpack2.unpack26, 12
  %.unpack2.elt27 = getelementptr inbounds nuw i8, ptr %w, i32 56
  %.unpack2.unpack28 = load i32, ptr %.unpack2.elt27, align 4
  %14 = insertvalue [16 x i32] %13, i32 %.unpack2.unpack28, 13
  %.unpack2.elt29 = getelementptr inbounds nuw i8, ptr %w, i32 60
  %.unpack2.unpack30 = load i32, ptr %.unpack2.elt29, align 4
  %15 = insertvalue [16 x i32] %14, i32 %.unpack2.unpack30, 14
  %.unpack2.elt31 = getelementptr inbounds nuw i8, ptr %w, i32 64
  %.unpack2.unpack32 = load i32, ptr %.unpack2.elt31, align 4
  %.unpack233 = insertvalue [16 x i32] %15, i32 %.unpack2.unpack32, 15
  %16 = call i32 @sumWithArrayC(i32 %.unpack, [16 x i32] %.unpack233)
  ret i32 %16

deref.throw:                                      ; preds = %entry
  call void @runtime.nilPanic(ptr undef) #9
  unreachable
}

; Function Attrs: nounwind
define linkonce_odr i32 @"sumWithArrayC$invoke"(ptr %0) unnamed_addr #1 {
entry:
  %.unpack = load i32, ptr %0, align 4
  %.elt1 = getelementptr inbounds nuw i8, ptr %0, i32 4
  %.unpack2.unpack = load i32, ptr %.elt1, align 4
  %1 = insertvalue [16 x i32] poison, i32 %.unpack2.unpack, 0
  %.unpack2.elt3 = getelementptr inbounds nuw i8, ptr %0, i32 8
  %.unpack2.unpack4 = load i32, ptr %.unpack2.elt3, align 4
  %2 = insertvalue [16 x i32] %1, i32 %.unpack2.unpack4, 1
  %.unpack2.elt5 = getelementptr inbounds nuw i8, ptr %0, i32 12
  %.unpack2.unpack6 = load i32, ptr %.unpack2.elt5, align 4
  %3 = insertvalue [16 x i32] %2, i32 %.unpack2.unpack6, 2
  %.unpack2.elt7 = getelementptr inbounds nuw i8, ptr %0, i32 16
  %.unpack2.unpack8 = load i32, ptr %.unpack2.elt7, align 4
  %4 = insertvalue [16 x i32] %3, i32 %.unpack2.unpack8, 3
  %.unpack2.elt9 = getelementptr inbounds nuw i8, ptr %0, i32 20
  %.unpack2.unpack10 = load i32, ptr %.unpack2.elt9, align 4
  %5 = insertvalue [16 x i32] %4, i32 %.unpack2.unpack10, 4
  %.unpack2.elt11 = getelementptr inbounds nuw i8, ptr %0, i32 24
  %.unpack2.unpack12 = load i32, ptr %.unpack2.elt11, align 4
  %6 = insertvalue [16 x i32] %5, i32 %.unpack2.unpack12, 5
  %.unpack2.elt13 = getelementptr inbounds nuw i8, ptr %0, i32 28
  %.unpack2.unpack14 = load i32, ptr %.unpack2.elt13, align 4
  %7 = insertvalue [16 x i32] %6, i32 %.unpack2.unpack14, 6
  %.unpack2.elt15 = getelementptr inbounds nuw i8, ptr %0, i32 32
  %.unpack2.unpack16 = load i32, ptr %.unpack2.elt15, align 4
  %8 = insertvalue [16 x i32] %7, i32 %.unpack2.unpack16, 7
  %.unpack2.elt17 = getelementptr inbounds nuw i8, ptr %0, i32 36
  %.unpack2.unpack18 = load i32, ptr %.unpack2.elt17, align 4
  %9 = insertvalue [16 x i32] %8, i32 %.unpack2.unpack18, 8
  %.unpack2.elt19 = getelementptr inbounds nuw i8, ptr %0, i32 40
  %.unpack2.unpack20 = load i32, ptr %.unpack2.elt19, align 4
  %10 = insertvalue [16 x i32] %9, i32 %.unpack2.unpack20, 9
  %.unpack2.elt21 = getelementptr inbounds nuw i8, ptr %0, i32 44
  %.unpack2.unpack22 = load i32, ptr %.unpack2.elt21, align 4
  %11 = insertvalue [16 x i32] %10, i32 %.unpack2.unpack22, 10
  %.unpack2.elt23 = getelementptr inbounds nuw i8, ptr %0, i32 48
  %.unpack2.unpack24 = load i32, ptr %.unpack2.elt23, align 4
  %12 = insertvalue [16 x i32] %11, i32 %.unpack2.unpack24, 11
  %.unpack2.elt25 = getelementptr inbounds nuw i8, ptr %0, i32 52
  %.unpack2.unpack26 = load i32, ptr %.unpack2.elt25, align 4
  %13 = insertvalue [16 x i32] %12, i32 %.unpack2.unpack26, 12
  %.unpack2.elt27 = getelementptr inbounds nuw i8, ptr %0, i32 56
  %.unpack2.unpack28 = load i32, ptr %.unpack2.elt27, align 4
  %14 = insertvalue [16 x i32] %13, i32 %.unpack2.unpack28, 13
  %.unpack2.elt29 = getelementptr inbounds nuw i8, ptr %0, i32 60
  %.unpack2.unpack30 = load i32, ptr %.unpack2.elt29, align 4
  %15 = insertvalue [16 x i32] %14, i32 %.unpack2.unpack30, 14
  %.unpack2.elt31 = getelementptr inbounds nuw i8, ptr %0, i32 64
  %.unpack2.unpack32 = load i32, ptr %.unpack2.elt31, align 4
  %.unpack233 = insertvalue [16 x i32] %15, i32 %.unpack2.unpack32, 15
  %ret = call i32 @sumWithArrayC(i32 %.unpack, [16 x i32] %.unpack233)
  ret i32 %ret
}

attributes #0 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #1 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #2 = { allockind("alloc,zeroed") allocsize(0) "alloc-family"="runtime.alloc" "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #3 = { nocallback nofree nounwind willreturn memory(argmem: readwrite) }
attributes #4 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-export-name"="sumWithArrayC" }
attributes #5 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "wasm-export-name"="takeBigC" }
attributes #6 = { noinline nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" }
attributes #7 = { "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-invoke"="main.$methods.sum:func:{}{basic:int}" "tinygo-methods"="main.$methods.sum:func:{}{basic:int}" }
attributes #8 = { nounwind "target-features"="+bulk-memory,+bulk-memory-opt,+call-indirect-overlong,+mutable-globals,+nontrapping-fptoint,+sign-ext,-multivalue,-reference-types" "tinygo-gowrapper"="main.spawnBig" }
attributes #9 = { nounwind }
