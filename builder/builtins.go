package builder

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blakesmith/ar"
	"github.com/tinygo-org/tinygo/goenv"
)

// These are the GENERIC_SOURCES according to CMakeList.txt.
var genericBuiltins = []string{
	"absvdi2.c",
	"absvsi2.c",
	"absvti2.c",
	"adddf3.c",
	"addsf3.c",
	"addtf3.c",
	"addvdi3.c",
	"addvsi3.c",
	"addvti3.c",
	"apple_versioning.c",
	"ashldi3.c",
	"ashlti3.c",
	"ashrdi3.c",
	"ashrti3.c",
	"bswapdi2.c",
	"bswapsi2.c",
	"clzdi2.c",
	"clzsi2.c",
	"clzti2.c",
	"cmpdi2.c",
	"cmpti2.c",
	"comparedf2.c",
	"comparesf2.c",
	"ctzdi2.c",
	"ctzsi2.c",
	"ctzti2.c",
	"divdc3.c",
	"divdf3.c",
	"divdi3.c",
	"divmoddi4.c",
	"divmodsi4.c",
	"divsc3.c",
	"divsf3.c",
	"divsi3.c",
	"divtc3.c",
	"divti3.c",
	"divtf3.c",
	"extendsfdf2.c",
	"extendhfsf2.c",
	"ffsdi2.c",
	"ffssi2.c",
	"ffsti2.c",
	"fixdfdi.c",
	"fixdfsi.c",
	"fixdfti.c",
	"fixsfdi.c",
	"fixsfsi.c",
	"fixsfti.c",
	"fixunsdfdi.c",
	"fixunsdfsi.c",
	"fixunsdfti.c",
	"fixunssfdi.c",
	"fixunssfsi.c",
	"fixunssfti.c",
	"floatdidf.c",
	"floatdisf.c",
	"floatsidf.c",
	"floatsisf.c",
	"floattidf.c",
	"floattisf.c",
	"floatundidf.c",
	"floatundisf.c",
	"floatunsidf.c",
	"floatunsisf.c",
	"floatuntidf.c",
	"floatuntisf.c",
	//"int_util.c",
	"lshrdi3.c",
	"lshrti3.c",
	"moddi3.c",
	"modsi3.c",
	"modti3.c",
	"muldc3.c",
	"muldf3.c",
	"muldi3.c",
	"mulodi4.c",
	"mulosi4.c",
	"muloti4.c",
	"mulsc3.c",
	"mulsf3.c",
	"multi3.c",
	"multf3.c",
	"mulvdi3.c",
	"mulvsi3.c",
	"mulvti3.c",
	"negdf2.c",
	"negdi2.c",
	"negsf2.c",
	"negti2.c",
	"negvdi2.c",
	"negvsi2.c",
	"negvti2.c",
	"os_version_check.c",
	"paritydi2.c",
	"paritysi2.c",
	"parityti2.c",
	"popcountdi2.c",
	"popcountsi2.c",
	"popcountti2.c",
	"powidf2.c",
	"powisf2.c",
	"powitf2.c",
	"subdf3.c",
	"subsf3.c",
	"subvdi3.c",
	"subvsi3.c",
	"subvti3.c",
	"subtf3.c",
	"trampoline_setup.c",
	"truncdfhf2.c",
	"truncdfsf2.c",
	"truncsfhf2.c",
	"ucmpdi2.c",
	"ucmpti2.c",
	"udivdi3.c",
	"udivmoddi4.c",
	"udivmodsi4.c",
	"udivmodti4.c",
	"udivsi3.c",
	"udivti3.c",
	"umoddi3.c",
	"umodsi3.c",
	"umodti3.c",
}

var aeabiBuiltins = []string{
	"arm/aeabi_cdcmp.S",
	"arm/aeabi_cdcmpeq_check_nan.c",
	"arm/aeabi_cfcmp.S",
	"arm/aeabi_cfcmpeq_check_nan.c",
	"arm/aeabi_dcmp.S",
	"arm/aeabi_div0.c",
	"arm/aeabi_drsub.c",
	"arm/aeabi_fcmp.S",
	"arm/aeabi_frsub.c",
	"arm/aeabi_idivmod.S",
	"arm/aeabi_ldivmod.S",
	"arm/aeabi_memcmp.S",
	"arm/aeabi_memcpy.S",
	"arm/aeabi_memmove.S",
	"arm/aeabi_memset.S",
	"arm/aeabi_uidivmod.S",
	"arm/aeabi_uldivmod.S",
}

func builtinFiles(target string) []string {
	builtins := append([]string{}, genericBuiltins...) // copy genericBuiltins
	if strings.HasPrefix(target, "arm") {
		builtins = append(builtins, aeabiBuiltins...)
	}
	return builtins
}

// builtinsDir returns the directory where the sources for compiler-rt are kept.
func builtinsDir() string {
	return filepath.Join(goenv.Get("TINYGOROOT"), "lib", "compiler-rt", "lib", "builtins")
}

// Get the builtins archive, possibly generating it as needed.
func loadBuiltins(target string) (path string, err error) {
	// Try to load a precompiled compiler-rt library.
	precompiledPath := filepath.Join(goenv.Get("TINYGOROOT"), "pkg", target, "compiler-rt.a")
	if _, err := os.Stat(precompiledPath); err == nil {
		// Found a precompiled compiler-rt for this OS/architecture. Return the
		// path directly.
		return precompiledPath, nil
	}

	outfile := "librt-" + target + ".a"
	builtinsDir := builtinsDir()

	builtins := builtinFiles(target)
	srcs := make([]string, len(builtins))
	for i, name := range builtins {
		srcs[i] = filepath.Join(builtinsDir, name)
	}

	if path, err := cacheLoad(outfile, commands["clang"][0], srcs); path != "" || err != nil {
		return path, err
	}

	var cachepath string
	err = CompileBuiltins(target, func(path string) error {
		path, err := cacheStore(path, outfile, commands["clang"][0], srcs)
		cachepath = path
		return err
	})
	return cachepath, err
}

// CompileBuiltins compiles builtins from compiler-rt into a static library.
// When it succeeds, it will call the callback with the resulting path. The path
// will be removed after callback returns. If callback returns an error, this is
// passed through to the return value of this function.
func CompileBuiltins(target string, callback func(path string) error) error {
	builtinsDir := builtinsDir()

	builtins := builtinFiles(target)
	srcs := make([]string, len(builtins))
	for i, name := range builtins {
		srcs[i] = filepath.Join(builtinsDir, name)
	}

	dirPrefix := "tinygo-builtins"
	remapDir := filepath.Join(os.TempDir(), dirPrefix)
	dir, err := ioutil.TempDir(os.TempDir(), dirPrefix)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	// Compile all builtins.
	// TODO: use builtins optimized for a given target if available.
	objs := make([]string, 0, len(builtins))
	for _, name := range builtins {
		objname := name
		if strings.LastIndexByte(objname, '/') >= 0 {
			objname = objname[strings.LastIndexByte(objname, '/'):]
		}
		objpath := filepath.Join(dir, objname+".o")
		objs = append(objs, objpath)
		srcpath := filepath.Join(builtinsDir, name)
		// Note: -fdebug-prefix-map is necessary to make the output archive
		// reproducible. Otherwise the temporary directory is stored in the
		// archive itself, which varies each run.
		err := execCommand(commands["clang"], "-c", "-Oz", "-g", "-Werror", "-Wall", "-std=c11", "-fshort-enums", "-nostdlibinc", "-ffunction-sections", "-fdata-sections", "--target="+target, "-fdebug-prefix-map="+dir+"="+remapDir, "-o", objpath, srcpath)
		if err != nil {
			return &commandError{"failed to build", srcpath, err}
		}
	}

	// Put all builtins in an archive to link as a static library.
	// Note: this does not create a symbol index, but ld.lld doesn't seem to
	// care.
	arpath := filepath.Join(dir, "librt.a")
	arfile, err := os.Create(arpath)
	if err != nil {
		return err
	}
	defer arfile.Close()
	arwriter := ar.NewWriter(arfile)
	err = arwriter.WriteGlobalHeader()
	if err != nil {
		return &os.PathError{"write ar header", arpath, err}
	}
	for _, objpath := range objs {
		name := filepath.Base(objpath)
		objfile, err := os.Open(objpath)
		if err != nil {
			return err
		}
		defer objfile.Close()
		st, err := objfile.Stat()
		if err != nil {
			return err
		}
		arwriter.WriteHeader(&ar.Header{
			Name:    name,
			ModTime: time.Unix(0, 0),
			Uid:     0,
			Gid:     0,
			Mode:    0644,
			Size:    st.Size(),
		})
		n, err := io.Copy(arwriter, objfile)
		if err != nil {
			return err
		}
		if n != st.Size() {
			return errors.New("file modified during ar creation: " + arpath)
		}
	}

	// Give the caller the resulting file. The callback must copy the file,
	// because after it returns the temporary directory will be removed.
	arfile.Close()
	return callback(arpath)
}
