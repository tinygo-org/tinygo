package builder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

// These are the GENERIC_SOURCES according to CMakeList.txt except for
// divmodsi4.c and udivmodsi4.c.
var genericBuiltins = []string{
	"absvdi2.c",
	"absvsi2.c",
	"absvti2.c",
	"adddf3.c",
	"addsf3.c",
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
	//"divmodsi4.c",
	"divmodti4.c",
	"divsc3.c",
	"divsf3.c",
	"divsi3.c",
	"divti3.c",
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
	"fp_mode.c",
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
	"subdf3.c",
	"subsf3.c",
	"subvdi3.c",
	"subvsi3.c",
	"subvti3.c",
	"trampoline_setup.c",
	"truncdfhf2.c",
	"truncdfsf2.c",
	"truncsfhf2.c",
	"ucmpdi2.c",
	"ucmpti2.c",
	"udivdi3.c",
	"udivmoddi4.c",
	//"udivmodsi4.c",
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

	// These two are not technically EABI builtins but are used by them and only
	// seem to be used on ARM. LLVM seems to use __divsi3 and __modsi3 on most
	// other architectures.
	// Most importantly, they have a different calling convention on AVR so
	// should not be used on AVR.
	"divmodsi4.c",
	"udivmodsi4.c",
}

var avrBuiltins = []string{
	"avr/divmodhi4.S",
	"avr/divmodqi4.S",
	"avr/mulhi3.S",
	"avr/mulqi3.S",
	"avr/udivmodhi4.S",
	"avr/udivmodqi4.S",
}

// CompilerRT is a library with symbols required by programs compiled with LLVM.
// These symbols are for operations that cannot be emitted with a single
// instruction or a short sequence of instructions for that target.
//
// For more information, see: https://compiler-rt.llvm.org/
var CompilerRT = Library{
	name: "compiler-rt",
	cflags: func(target, headerPath string) []string {
		return []string{"-Werror", "-Wall", "-std=c11", "-nostdlibinc"}
	},
	sourceDir: func() string {
		llvmDir := filepath.Join(goenv.Get("TINYGOROOT"), "llvm-project/compiler-rt/lib/builtins")
		if _, err := os.Stat(llvmDir); err == nil {
			// Release build.
			return llvmDir
		}
		// Development build.
		return filepath.Join(goenv.Get("TINYGOROOT"), "lib/compiler-rt-builtins")
	},
	librarySources: func(target string) ([]string, error) {
		builtins := append([]string{}, genericBuiltins...) // copy genericBuiltins
		if strings.HasPrefix(target, "arm") || strings.HasPrefix(target, "thumb") {
			builtins = append(builtins, aeabiBuiltins...)
		}
		if strings.HasPrefix(target, "avr") {
			builtins = append(builtins, avrBuiltins...)
		}
		return builtins, nil
	},
}
