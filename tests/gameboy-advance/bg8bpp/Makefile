# User-configurable flags
NAME           := bg8bpp
TINYGO_FLAGS   :=
TILEPREP_FLAGS := --bits=8
TILES_PREFIX   := tiles

# Auto-generated inputs
GO_INPUTS    := $(wildcard *.go)
TILE_INPUTS  := $(wildcard *.png)
TILE_OUTPUTS := $(TILES_PREFIX)_gen.s $(TILES_PREFIX)_gen.go

# Auto-generated outputs
GBA  := $(NAME).gba
OUTS := $(GBA)

# Auto-detect dependencies
BUILD_TARGETS := $(GBA)

# If there are .png assets in the directory, run tileprep
ifneq (,$(TILE_INPUTS))
BUILD_TARGETS += tiles
TINYGO_FLAGS  += --extra="$(TILES_PREFIX)_gen.s"
OUTS          += $(TILE_OUTPUTS)
endif

.PHONY : build clean tiles

build : tiles $(BUILD_TARGETS)

clean :
	rm -f $(OUTS)

tiles :
	-tileprep $(TILEPREP_FLAGS) --go_out=$(TILES_PREFIX)_gen.go --asm_out=$(TILES_PREFIX)_gen.s $(TILE_INPUTS)
	-gofmt -w $(TILES_PREFIX)_gen.go

$(GBA) : $(filter-out $(OUTS),$(GO_INPUTS))
	go get . >/dev/null 2>&1 || true
	tinygo build --size=full --target=gameboy-advance $(TINYGO_FLAGS) -o $@ .
