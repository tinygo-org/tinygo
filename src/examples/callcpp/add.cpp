template <typename T>
T add(T a, T b) {
    return a + b;
}

extern "C" int add(int a, int b) {
    return add<int>(a, b);
}

