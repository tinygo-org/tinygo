//go:build none

// This file is included in the build on systems that support the Boehm GC,
// despite the //go:build line above.

#include <stdint.h>
#include <string.h>

typedef uintptr_t GC_word;
typedef GC_word GC_descr;
typedef void (*GC_push_other_roots_proc)(void);
typedef void (*GC_warn_proc)(const char *msg, uintptr_t arg);

#define GC_WORDSZ (8 * sizeof(GC_word))
#define GC_set_bit(bitmap, index)                                         \
    ((bitmap)[(index) / GC_WORDSZ] |=                                     \
     (GC_word)1 << ((index) % GC_WORDSZ))

void GC_set_push_other_roots(GC_push_other_roots_proc);
void GC_set_warn_proc(GC_warn_proc);
void *GC_malloc_atomic_uncollectable(uintptr_t);
void GC_free(void *);
GC_descr GC_make_descriptor(const GC_word *, uintptr_t);

void tinygo_runtime_bdwgc_callback(void);

struct descriptor_cache_entry {
    uintptr_t layout;
    GC_descr descriptor;
    struct descriptor_cache_entry *next;
};

static struct descriptor_cache_entry **descriptor_cache;
static size_t descriptor_cache_capacity;
static size_t descriptor_cache_count;

static size_t descriptor_cache_index(uintptr_t layout, size_t capacity) {
#if UINTPTR_MAX > UINT32_MAX
    layout ^= layout >> 33;
#endif
    layout ^= layout >> 16;
    layout *= 0x45d9f3b;
    layout ^= layout >> 16;
    return layout & (capacity - 1);
}

static int grow_descriptor_cache(void) {
    size_t new_capacity =
        descriptor_cache_capacity == 0 ? 64 : descriptor_cache_capacity * 2;
    struct descriptor_cache_entry **new_cache;
    size_t i;

    new_cache =
        GC_malloc_atomic_uncollectable(new_capacity * sizeof(*new_cache));
    if (new_cache == NULL) {
        return 0;
    }
    memset(new_cache, 0, new_capacity * sizeof(*new_cache));

    for (i = 0; i < descriptor_cache_capacity; i++) {
        struct descriptor_cache_entry *entry = descriptor_cache[i];
        while (entry != NULL) {
            struct descriptor_cache_entry *next = entry->next;
            size_t index =
                descriptor_cache_index(entry->layout, new_capacity);

            entry->next = new_cache[index];
            new_cache[index] = entry;
            entry = next;
        }
    }

    if (descriptor_cache != NULL) {
        GC_free(descriptor_cache);
    }
    descriptor_cache = new_cache;
    descriptor_cache_capacity = new_capacity;
    return 1;
}

static void callback(void) {
    tinygo_runtime_bdwgc_callback();
}

static void warn_proc(const char *msg, uintptr_t arg) {
}

void tinygo_runtime_bdwgc_init(void) {
    GC_set_push_other_roots(callback);
#if defined(__wasm__)
    // There are a lot of warnings on WebAssembly in the form:
    //
    //     GC Warning: Repeated allocation of very large block (appr. size 68 KiB):
    //         May lead to memory leak and poor performance
    //
    // The usual advice is to use something like GC_malloc_ignore_off_page but
    // unfortunately for most allocations that's not allowed: Go allocations can
    // legitimately hold pointers further than one page in the allocation. So
    // instead we just disable the warning.
    GC_set_warn_proc(warn_proc);
#endif
}

GC_descr tinygo_runtime_bdwgc_make_descriptor(uintptr_t layout) {
    struct descriptor_cache_entry *entry;
    GC_word inline_bitmap;
    GC_word *bitmap;
    size_t index;
    size_t bit_count;
    size_t word_count;
    GC_descr descriptor;

    if (descriptor_cache_count >= descriptor_cache_capacity * 2 &&
        !grow_descriptor_cache()) {
        return 0;
    }

    index = descriptor_cache_index(layout, descriptor_cache_capacity);
    for (entry = descriptor_cache[index]; entry != NULL; entry = entry->next) {
        if (entry->layout == layout) {
            return entry->descriptor;
        }
    }

    entry = GC_malloc_atomic_uncollectable(sizeof(*entry));
    if (entry == NULL) {
        return 0;
    }

    if (layout & 1) {
        const size_t size_bits = 4 + sizeof(uintptr_t) / 4;
        const uintptr_t size_mask = ((uintptr_t)1 << size_bits) - 1;

        bit_count = (layout >> 1) & size_mask;
        inline_bitmap = layout >> (size_bits + 1);
        bitmap = &inline_bitmap;
    } else {
        const uint8_t *bytes = (const uint8_t *)(layout + sizeof(uintptr_t));
        size_t i;

        bit_count = *(const uintptr_t *)layout;
        word_count = (bit_count + GC_WORDSZ - 1) / GC_WORDSZ;
        bitmap = GC_malloc_atomic_uncollectable(word_count * sizeof(GC_word));
        if (bitmap == NULL) {
            GC_free(entry);
            return 0;
        }
        memset(bitmap, 0, word_count * sizeof(GC_word));
        for (i = 0; i < bit_count; i++) {
            if ((bytes[i / 8] >> (i % 8)) & 1) {
                GC_set_bit(bitmap, i);
            }
        }
    }

    descriptor = GC_make_descriptor(bitmap, bit_count);
    if (!(layout & 1)) {
        GC_free(bitmap);
    }
    if (descriptor == 0) {
        GC_free(entry);
        return 0;
    }

    entry->layout = layout;
    entry->descriptor = descriptor;
    entry->next = descriptor_cache[index];
    descriptor_cache[index] = entry;
    descriptor_cache_count++;
    return descriptor;
}
