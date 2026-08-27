# Developer tooling: lint, spellcheck, and the help target.

.PHONY: tools
tools:
	go generate -tags tools ./

LINTDIRS=src/os/ src/reflect/
.PHONY: lint
lint: tools ## Lint source tree
	revive -version
	# TODO: lint more directories!
	# revive.toml isn't flexible enough to filter out just one kind of error from a checker, so do it with grep here.
	# Can't use grep with friendly formatter.  Plain output isn't too bad, though.
	# Use 'grep .' to get rid of stray blank line
	revive -config revive.toml compiler/... $$( find $(LINTDIRS) -type f -name '*.go' ) \
		| grep -v "should have comment or be unexported" \
		| grep '.' \
		| awk '{print}; END {exit NR>0}'

SPELLDIRSCMD=find . -depth 1 -type d  | egrep -wv '.git|lib|llvm|src'; find src -depth 1 | egrep -wv 'device|internal|net|vendor'; find src/internal -depth 1 -type d | egrep -wv src/internal/wasi
.PHONY: spell
spell: tools ## Spellcheck source tree
	misspell -error --dict misspell.csv -i 'ackward,devided,extint,rela' $$( $(SPELLDIRSCMD) ) *.go *.md

.PHONY: spellfix
spellfix: tools ## Same as spell, but fixes what it finds
	misspell -w --dict misspell.csv -i 'ackward,devided,extint,rela' $$( $(SPELLDIRSCMD) ) *.go *.md

# https://www.client9.com/self-documenting-makefiles/
.PHONY: help
help:
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {\
	gsub(/\$$\(LLVM_BUILDDIR\)/, "$(LLVM_BUILDDIR)"); \
        printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF \
        }' $(MAKEFILE_LIST)
#.DEFAULT_GOAL=help
