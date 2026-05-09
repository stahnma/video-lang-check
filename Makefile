BINARY := speech-check
CC ?= clang

# Use the flox environment's whisper.cpp libraries and headers
FLOX_PREFIX := $(CURDIR)/.flox/run/x86_64-linux.speech-check.dev
INCLUDE_PATH := $(FLOX_PREFIX)/include
LIB_PATH := $(FLOX_PREFIX)/lib
LOCAL_LIB := $(CURDIR)/lib

.PHONY: build clean deps link-fixup

build: link-fixup deps
	CGO_ENABLED=1 CC=$(CC) \
	C_INCLUDE_PATH=$(INCLUDE_PATH) \
	LIBRARY_PATH=$(LOCAL_LIB):$(LIB_PATH) \
	CGO_LDFLAGS="-Wl,-rpath,$(LIB_PATH)" \
	go build -o $(BINARY) .

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

clean:
	rm -f $(BINARY)
	rm -rf $(LOCAL_LIB)

deps:
	go mod tidy
