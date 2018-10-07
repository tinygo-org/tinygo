package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// These are the GENERIC_SOURCES according to CMakeList.txt.
var genericBuiltins = []string{
	"absvdi2",
	"absvsi2",
	"absvti2",
	"adddf3",
	"addsf3",
	"addtf3",
	"addvdi3",
	"addvsi3",
	"addvti3",
	"apple_versioning",
	"ashldi3",
	"ashlti3",
	"ashrdi3",
	"ashrti3",
	"bswapdi2",
	"bswapsi2",
	"clzdi2",
	"clzsi2",
	"clzti2",
	"cmpdi2",
	"cmpti2",
	"comparedf2",
	"comparesf2",
	"ctzdi2",
	"ctzsi2",
	"ctzti2",
	"divdc3",
	"divdf3",
	"divdi3",
	"divmoddi4",
	"divmodsi4",
	"divsc3",
	"divsf3",
	"divsi3",
	"divtc3",
	"divti3",
	"divtf3",
	"extendsfdf2",
	"extendhfsf2",
	"ffsdi2",
	"ffssi2",
	"ffsti2",
	"fixdfdi",
	"fixdfsi",
	"fixdfti",
	"fixsfdi",
	"fixsfsi",
	"fixsfti",
	"fixunsdfdi",
	"fixunsdfsi",
	"fixunsdfti",
	"fixunssfdi",
	"fixunssfsi",
	"fixunssfti",
	"floatdidf",
	"floatdisf",
	"floatsidf",
	"floatsisf",
	"floattidf",
	"floattisf",
	"floatundidf",
	"floatundisf",
	"floatunsidf",
	"floatunsisf",
	"floatuntidf",
	"floatuntisf",
	//"int_util",
	"lshrdi3",
	"lshrti3",
	"moddi3",
	"modsi3",
	"modti3",
	"muldc3",
	"muldf3",
	"muldi3",
	"mulodi4",
	"mulosi4",
	"muloti4",
	"mulsc3",
	"mulsf3",
	"multi3",
	"multf3",
	"mulvdi3",
	"mulvsi3",
	"mulvti3",
	"negdf2",
	"negdi2",
	"negsf2",
	"negti2",
	"negvdi2",
	"negvsi2",
	"negvti2",
	"os_version_check",
	"paritydi2",
	"paritysi2",
	"parityti2",
	"popcountdi2",
	"popcountsi2",
	"popcountti2",
	"powidf2",
	"powisf2",
	"powitf2",
	"subdf3",
	"subsf3",
	"subvdi3",
	"subvsi3",
	"subvti3",
	"subtf3",
	"trampoline_setup",
	"truncdfhf2",
	"truncdfsf2",
	"truncsfhf2",
	"ucmpdi2",
	"ucmpti2",
	"udivdi3",
	"udivmoddi4",
	"udivmodsi4",
	"udivmodti4",
	"udivsi3",
	"udivti3",
	"umoddi3",
	"umodsi3",
	"umodti3",
}

// Get the builtins archive, possibly generating it as needed.
func loadBuiltins(target string) (path string, err error) {
	outfile := "librt-" + target + ".a"
	builtinsDir := filepath.Join(sourceDir(), "lib", "compiler-rt", "lib", "builtins")

	srcs := make([]string, len(genericBuiltins))
	for i, name := range genericBuiltins {
		srcs[i] = filepath.Join(builtinsDir, name+".c")
	}

	if path, err := cacheLoad(outfile, commands["clang"], srcs); path != "" || err != nil {
		return path, err
	}

	dir, err := ioutil.TempDir("", "tinygo-builtins")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	// Compile all builtins.
	// TODO: use builtins optimized for a given target if available.
	objs := make([]string, 0, len(genericBuiltins))
	for _, name := range genericBuiltins {
		objpath := filepath.Join(dir, name+".o")
		objs = append(objs, objpath)
		srcpath := filepath.Join(builtinsDir, name+".c")
		cmd := exec.Command(commands["clang"], "-c", "-Oz", "-g", "-Werror", "-Wall", "-std=c11", "--target="+target, "-o", objpath, srcpath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = dir
		err = cmd.Run()
		if err != nil {
			return "", err
		}
	}

	// Put all builtins in an archive to link as a static library.
	arpath := filepath.Join(dir, "librt.a")
	cmd := exec.Command(commands["ar"], append([]string{"cr", arpath}, objs...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return cacheStore(arpath, outfile, commands["clang"], srcs)
}
