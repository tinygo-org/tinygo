GBA    := sprites.gba
SRCS   := *.go

TINYGO_FLAGS :=

.PHONY : build clean

build : $(GBA)

clean :
	@rm -f $(GBA)

$(GBA) : *.go
	@go get . >/dev/null 2>&1 || true
	@tinygo build --target=gameboy-advance $(TINYGO_FLAGS) -o $@ .

