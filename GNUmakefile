# aliases
all: tinygo

# Build rules are split by topic. config.mk must come first because the other
# files use its variables in immediate assignments and conditionals.
include make/config.mk
include make/llvm.mk
include make/gen-device.mk
include make/build.mk
include make/test.mk
include make/smoketest.mk
include make/release.mk
include make/tools.mk
