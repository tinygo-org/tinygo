package builder

import (
	"path/filepath"

	"github.com/tinygo-org/tinygo/goenv"
)

// Picolibc is a C library for bare metal embedded devices. It was originally
// based on newlib.
var Picolibc = Library{
	name: "picolibc",
	//lib/picolibc/newlib/libc/include/
	cflags: func() []string {
		picolibcDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/picolibc/newlib/libc")
		return []string{"-Werror", "-Wall", "-std=gnu11", "-D_COMPILING_NEWLIB", "-I"+picolibcDir+"/include","--sysroot=" + picolibcDir, "-I" + picolibcDir + "/tinystdio", "-I" + goenv.Get("TINYGOROOT") + "/lib/picolibc-include"}
	},
	sourceDir: "lib/picolibc/newlib/libc",
	sources: func(target string) []string {
		return picolibcSources
	},
}

var picolibcSources = []string{
	"string/bcmp.c",
	"string/bcopy.c",
	"string/bzero.c",
	"string/explicit_bzero.c",
	"string/ffsl.c",
	"string/ffsll.c",
	"string/fls.c",
	"string/flsl.c",
	"string/flsll.c",
	"string/gnu_basename.c",
	"string/index.c",
	"string/memccpy.c",
	"string/memchr.c",
	"string/memcmp.c",
	"string/memcpy.c",
	"string/memmem.c",
	"string/memmove.c",
	"string/mempcpy.c",
	"string/memrchr.c",
	"string/memset.c",
	"string/rawmemchr.c",
	"string/rindex.c",
	"string/stpcpy.c",
	"string/stpncpy.c",
	"string/strcasecmp.c",
	"string/strcasecmp_l.c",
	"string/strcasestr.c",
	"string/strcat.c",
	"string/strchr.c",
	"string/strchrnul.c",
	"string/strcmp.c",
	"string/strcoll.c",
	"string/strcoll_l.c",
	"string/strcpy.c",
	"string/strcspn.c",
	"string/strdup.c",
	"string/strerror.c",
	"string/strerror_r.c",
	"string/strlcat.c",
	"string/strlcpy.c",
	"string/strlen.c",
	"string/strlwr.c",
	"string/strncasecmp.c",
	"string/strncasecmp_l.c",
	"string/strncat.c",
	"string/strncmp.c",
	"string/strncpy.c",
	"string/strndup.c",
	"string/strnlen.c",
	"string/strnstr.c",
	"string/strpbrk.c",
	"string/strrchr.c",
	"string/strsep.c",
	"string/strsignal.c",
	"string/strspn.c",
	"string/strstr.c",
	"string/strtok.c",
	"string/strtok_r.c",
	"string/strupr.c",
	"string/strverscmp.c",
	"string/strxfrm.c",
	"string/strxfrm_l.c",
	"string/swab.c",
	"string/timingsafe_bcmp.c",
	"string/timingsafe_memcmp.c",
	"string/u_strerr.c",
	"string/wcpcpy.c",
	"string/wcpncpy.c",
	"string/wcscasecmp.c",
	"string/wcscasecmp_l.c",
	"string/wcscat.c",
	"string/wcschr.c",
	"string/wcscmp.c",
	"string/wcscoll.c",
	"string/wcscoll_l.c",
	"string/wcscpy.c",
	"string/wcscspn.c",
	"string/wcsdup.c",
	"string/wcslcat.c",
	"string/wcslcpy.c",
	"string/wcslen.c",
	"string/wcsncasecmp.c",
	"string/wcsncasecmp_l.c",
	"string/wcsncat.c",
	"string/wcsncmp.c",
	"string/wcsncpy.c",
	"string/wcsnlen.c",
	"string/wcspbrk.c",
	"string/wcsrchr.c",
	"string/wcsspn.c",
	"string/wcsstr.c",
	"string/wcstok.c",
	"string/wcswidth.c",
	"string/wcsxfrm.c",
	"string/wcsxfrm_l.c",
	"string/wcwidth.c",
	"string/wmemchr.c",
	"string/wmemcmp.c",
	"string/wmemcpy.c",
	"string/wmemmove.c",
	"string/wmempcpy.c",
	"string/wmemset.c",
	"string/xpg_strerror_r.c",
}
