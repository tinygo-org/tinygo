/*
  pgmspace.h - Definitions for compatibility with AVR pgmspace macros

  Copyright (c) 2015 Arduino LLC

  Based on work of Paul Stoffregen on Teensy 3 (http://pjrc.com)

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in
  all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
  THE SOFTWARE
*/

#ifndef __PGMSPACE_H_
#define __PGMSPACE_H_ 1

#include <inttypes.h>

#define PROGMEM
#define PGM_P  const char *
#define PSTR(str) (str)

#define _SFR_BYTE(n) (n)

typedef void prog_void;
typedef char prog_char;
typedef unsigned char prog_uchar;
typedef int8_t prog_int8_t;
typedef uint8_t prog_uint8_t;
typedef int16_t prog_int16_t;
typedef uint16_t prog_uint16_t;
typedef int32_t prog_int32_t;
typedef uint32_t prog_uint32_t;
typedef int64_t prog_int64_t;
typedef uint64_t prog_uint64_t;

typedef const void* int_farptr_t;
typedef const void* uint_farptr_t;

#define memchr_P(s, c, n) memchr((s), (c), (n))
#define memcmp_P(s1, s2, n) memcmp((s1), (s2), (n))
#define memccpy_P(dest, src, c, n) memccpy((dest), (src), (c), (n))
#define memcpy_P(dest, src, n) memcpy((dest), (src), (n))
#define memmem_P(haystack, haystacklen, needle, needlelen) memmem((haystack), (haystacklen), (needle), (needlelen))
#define memrchr_P(s, c, n) memrchr((s), (c), (n))
#define strcat_P(dest, src) strcat((dest), (src))
#define strchr_P(s, c) strchr((s), (c))
#define strchrnul_P(s, c) strchrnul((s), (c))
#define strcmp_P(a, b) strcmp((a), (b))
#define strcpy_P(dest, src) strcpy((dest), (src))
#define strcasecmp_P(s1, s2) strcasecmp((s1), (s2))
#define strcasestr_P(haystack, needle) strcasestr((haystack), (needle))
#define strcspn_P(s, accept) strcspn((s), (accept))
#define strlcat_P(s1, s2, n) strlcat((s1), (s2), (n))
#define strlcpy_P(s1, s2, n) strlcpy((s1), (s2), (n))
#define strlen_P(a) strlen((a))
#define strnlen_P(s, n) strnlen((s), (n))
#define strncmp_P(s1, s2, n) strncmp((s1), (s2), (n))
#define strncasecmp_P(s1, s2, n) strncasecmp((s1), (s2), (n))
#define strncat_P(s1, s2, n) strncat((s1), (s2), (n))
#define strncpy_P(s1, s2, n) strncpy((s1), (s2), (n))
#define strpbrk_P(s, accept) strpbrk((s), (accept))
#define strrchr_P(s, c) strrchr((s), (c))
#define strsep_P(sp, delim) strsep((sp), (delim))
#define strspn_P(s, accept) strspn((s), (accept))
#define strstr_P(a, b) strstr((a), (b))
#define strtok_P(s, delim) strtok((s), (delim))
#define strtok_rP(s, delim, last) strtok((s), (delim), (last))

#define strlen_PF(a) strlen((a))
#define strnlen_PF(src, len) strnlen((src), (len))
#define memcpy_PF(dest, src, len) memcpy((dest), (src), (len))
#define strcpy_PF(dest, src) strcpy((dest), (src))
#define strncpy_PF(dest, src, len) strncpy((dest), (src), (len))
#define strcat_PF(dest, src) strcat((dest), (src))
#define strlcat_PF(dest, src, len) strlcat((dest), (src), (len))
#define strncat_PF(dest, src, len) strncat((dest), (src), (len))
#define strcmp_PF(s1, s2) strcmp((s1), (s2))
#define strncmp_PF(s1, s2, n) strncmp((s1), (s2), (n))
#define strcasecmp_PF(s1, s2) strcasecmp((s1), (s2))
#define strncasecmp_PF(s1, s2, n) strncasecmp((s1), (s2), (n))
#define strstr_PF(s1, s2) strstr((s1), (s2))
#define strlcpy_PF(dest, src, n) strlcpy((dest), (src), (n))
#define memcmp_PF(s1, s2, n) memcmp((s1), (s2), (n))

#define sprintf_P(s, f, ...) sprintf((s), (f), __VA_ARGS__)
#define snprintf_P(s, f, ...) snprintf((s), (f), __VA_ARGS__)

#define pgm_read_byte(addr) (*(const unsigned char *)(addr))
#define pgm_read_word(addr) (*(const unsigned short *)(addr))
#define pgm_read_dword(addr) (*(const unsigned long *)(addr))
#define pgm_read_float(addr) (*(const float *)(addr))
#define pgm_read_ptr(addr) (*(const void *)(addr))

#define pgm_read_byte_near(addr) pgm_read_byte(addr)
#define pgm_read_word_near(addr) pgm_read_word(addr)
#define pgm_read_dword_near(addr) pgm_read_dword(addr)
#define pgm_read_float_near(addr) pgm_read_float(addr)
#define pgm_read_ptr_near(addr) pgm_read_ptr(addr)

#define pgm_read_byte_far(addr) pgm_read_byte(addr)
#define pgm_read_word_far(addr) pgm_read_word(addr)
#define pgm_read_dword_far(addr) pgm_read_dword(addr)
#define pgm_read_float_far(addr) pgm_read_float(addr)
#define pgm_read_ptr_far(addr) pgm_read_ptr(addr)

#define pgm_get_far_address(addr) (&(addr))

#endif
