BINARY := video-lang-check
CC ?= clang

# Use the flox environment's whisper.cpp libraries and headers
FLOX_PREFIX := $(CURDIR)/.flox/run/x86_64-linux.video-lang-check.dev
INCLUDE_PATH := $(FLOX_PREFIX)/include
LIB_PATH := $(FLOX_PREFIX)/lib
LOCAL_LIB := $(CURDIR)/lib

# Static build: whisper.cpp built from source
WHISPER_COMMIT := c81b2dabbc45
WHISPER_SRC := $(CURDIR)/third_party/whisper.cpp
WHISPER_BUILD := $(WHISPER_SRC)/build
STATIC_LIB := $(WHISPER_BUILD)/src
STATIC_GGML_LIB := $(WHISPER_BUILD)/ggml/src
STUB_LIB := $(WHISPER_BUILD)/stub

.PHONY: build clean deps fmt help link-fixup test static whisper-source whisper-static

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

# Clone whisper.cpp source at the pinned commit
whisper-source:
	@if [ ! -d "$(WHISPER_SRC)" ]; then \
		echo "Cloning whisper.cpp..."; \
		mkdir -p third_party; \
		git clone https://github.com/ggml-org/whisper.cpp.git $(WHISPER_SRC); \
		cd $(WHISPER_SRC) && git checkout $(WHISPER_COMMIT); \
	fi

# Build whisper.cpp as static libraries
whisper-static: whisper-source
	@echo "Building whisper.cpp static libraries..."
	cmake -B $(WHISPER_BUILD) -S $(WHISPER_SRC) \
		-DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=OFF \
		-DWHISPER_BUILD_EXAMPLES=OFF \
		-DWHISPER_BUILD_TESTS=OFF \
		-DGGML_OPENMP=OFF
	cmake --build $(WHISPER_BUILD) --config Release -j$$(nproc)
	@# The Go bindings unconditionally pass -fopenmp which links -lomp.
	@# Since we build without OpenMP, create a stub libomp.a to satisfy the linker.
	@mkdir -p $(STUB_LIB)
	@echo "void __omp_stub(void){}" | $(CC) -c -x c - -o $(STUB_LIB)/omp_stub.o
	@ar rcs $(STUB_LIB)/libomp.a $(STUB_LIB)/omp_stub.o

# Build a portable binary with whisper.cpp statically linked.
# Only glibc remains dynamic so the binary runs on any x86_64 Linux.
static: fmt whisper-static deps
	CGO_ENABLED=1 CC=$(CC) \
	C_INCLUDE_PATH=$(WHISPER_SRC)/include:$(WHISPER_SRC)/ggml/include \
	LIBRARY_PATH=$(STATIC_LIB):$(STATIC_GGML_LIB):$(STATIC_GGML_LIB)/ggml-cpu:$(STUB_LIB) \
	go build -o $(BINARY) ./cmd/video-lang-check
	patchelf --set-interpreter /lib64/ld-linux-x86-64.so.2 --remove-rpath $(BINARY)
	@echo ""
	@echo "Static build complete. Verify with: ldd $(BINARY)"

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

clean-all: clean
	rm -rf third_party $(WHISPER_BUILD)

deps:
	go mod tidy

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the $(BINARY) binary (default)"
	@echo "  static       Build a portable binary with static whisper.cpp"
	@echo "  test         Run tests"
	@echo "  fmt          Format Go source files"
	@echo "  clean        Remove build artifacts"
	@echo "  clean-all    Remove build artifacts and whisper.cpp source"
	@echo "  deps         Tidy Go modules"
	@echo "  link-fixup   Create libggml-cpu symlink for linker"
	@echo "  help         Show this help message"
