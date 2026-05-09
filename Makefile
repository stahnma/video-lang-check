BINARY := video-lang-check
CC ?= clang

# Use the flox environment's whisper.cpp libraries and headers
FLOX_PREFIX := $(CURDIR)/.flox/run/x86_64-linux.video-lang-check.dev
INCLUDE_PATH := $(FLOX_PREFIX)/include
LIB_PATH := $(FLOX_PREFIX)/lib
LOCAL_LIB := $(CURDIR)/lib

.PHONY: build clean deps fmt help link-fixup test

build: fmt link-fixup deps
	CGO_ENABLED=1 CC=$(CC) \
	C_INCLUDE_PATH=$(INCLUDE_PATH) \
	LIBRARY_PATH=$(LOCAL_LIB):$(LIB_PATH) \
	CGO_LDFLAGS="-Wl,-rpath,$(LIB_PATH)" \
	go build -o $(BINARY) ./cmd/video-lang-check

test: link-fixup deps
	CGO_ENABLED=1 CC=$(CC) \
	C_INCLUDE_PATH=$(INCLUDE_PATH) \
	LIBRARY_PATH=$(LOCAL_LIB):$(LIB_PATH) \
	go test ./...

# The Go bindings link -lggml-cpu but flox provides arch-specific variants.
# Create a local symlink to satisfy the linker.
link-fixup:
	@mkdir -p $(LOCAL_LIB)
	@if [ ! -e "$(LOCAL_LIB)/libggml-cpu.so" ]; then \
		variant=$$(ls $(LIB_PATH)/libggml-cpu-*.so 2>/dev/null | head -1); \
		if [ -n "$$variant" ]; then \
			ln -sf "$$variant" "$(LOCAL_LIB)/libggml-cpu.so"; \
			echo "Created symlink: lib/libggml-cpu.so -> $$variant"; \
		else \
			echo "Error: no libggml-cpu-*.so found in $(LIB_PATH)" >&2; \
			exit 1; \
		fi \
	fi

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY)
	rm -rf $(LOCAL_LIB)

deps:
	go mod tidy

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the $(BINARY) binary (default)"
	@echo "  test         Run tests"
	@echo "  fmt          Format Go source files"
	@echo "  clean        Remove build artifacts"
	@echo "  deps         Tidy Go modules"
	@echo "  link-fixup   Create libggml-cpu symlink for linker"
	@echo "  help         Show this help message"
