//go:build none

// This file is included in the build on systems that support the Boehm GC,
// despite the //go:build line above.

#include <stdint.h>
#include <string.h>

#include <gc/gc_mark.h>
#include <gc/gc_typed.h>

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

static void GC_CALLBACK callback(void) {
    tinygo_runtime_bdwgc_callback();
}

static void GC_CALLBACK warn_proc(const char *msg, GC_word arg) {
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

// This matches the start of finalizerEntry in gc_boehm_finalizer.go.
struct finalizer_entry {
    struct finalizer_entry *next;
    void *obj;
    uintptr_t offset;
};

// These point to the Go variables finalizerPending and finalizersQueued.
static struct finalizer_entry **finalizer_pending;
static unsigned char *finalizers_queued;

static void GC_CALLBACK finalizers_ready(void) {
    *finalizers_queued = 1;
}

// Add the entry to the Go list. Go runs the finalizer later.
static void GC_CALLBACK finalize(void *obj, void *data) {
    struct finalizer_entry *entry = data;

    entry->obj = (char *)obj + entry->offset;
    entry->next = *finalizer_pending;
    *finalizer_pending = entry;
}

// Queue finalizers during a collection and run them later from Go.
void tinygo_runtime_bdwgc_enable_finalizers(uintptr_t pending,
                                            uintptr_t queued) {
    finalizer_pending = (struct finalizer_entry **)pending;
    finalizers_queued = (unsigned char *)queued;
    GC_set_finalize_on_demand(1);
    // Keep what a queued object refers to alive until its finalizer ran.
    // See GC_finalize in lib/bdwgc/finalize.c.
    GC_set_java_finalization(1);
    GC_set_finalizer_notifier(finalizers_ready);
}

// Set or clear the finalizer of obj and return the data of the old one.
uintptr_t tinygo_runtime_bdwgc_register_finalizer(uintptr_t obj, uintptr_t data) {
    GC_finalization_proc old_proc = 0;
    void *old_data = NULL;

    GC_register_finalizer_no_order((void *)obj, data != 0 ? finalize : 0,
                                   (void *)data, &old_proc, &old_data);
    return (uintptr_t)old_data;
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
