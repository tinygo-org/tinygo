#if defined(SOME_CONSTANT) || !defined(SHARED_CONSTANT)
#error incorrect C++ flags
#endif

template <bool enabled> int constant() {
    if constexpr (enabled) {
        return SHARED_CONSTANT + CXX_CONSTANT;
    } else {
        return 0;
    }
}

extern "C" int cxxConstant(void) {
    return constant<true>();
}
