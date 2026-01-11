.PHONY: build build-vosk setup setup-whisper setup-vosk clean test

# Whisper model options: tiny, small, base, medium
WHISPER_MODEL ?= small
VOSK_VERSION ?= 0.3.42

MODELS_DIR := models
LIB_DIR := lib

# Basic build (whisper only, no vosk)
build:
	go build -o yapatype .

# Build with vosk support (macOS)
# Handles CGO flags and fixes runtime library path
build-vosk: $(LIB_DIR)/libvosk.dylib
	CGO_CFLAGS="-I$$(pwd)/$(LIB_DIR)" \
	CGO_LDFLAGS="-L$$(pwd)/$(LIB_DIR) -lvosk" \
	go build -tags vosk -o yapatype .
	@# Fix rpath so binary finds libvosk.dylib at runtime
	install_name_tool -add_rpath $$(pwd)/$(LIB_DIR) yapatype 2>/dev/null || true
	install_name_tool -change libvosk.dylib @rpath/libvosk.dylib yapatype

# Download whisper model
# Usage: make setup-whisper WHISPER_MODEL=small
setup-whisper: $(MODELS_DIR)
	@echo "Downloading whisper model: ggml-$(WHISPER_MODEL).en.bin"
	curl -L --progress-bar -o $(MODELS_DIR)/ggml-$(WHISPER_MODEL).en.bin \
		https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-$(WHISPER_MODEL).en.bin

# Download vosk library and model (macOS only)
setup-vosk: $(MODELS_DIR) $(LIB_DIR)
	@echo "Downloading vosk native library v$(VOSK_VERSION)..."
	curl -L -o /tmp/vosk-osx.zip \
		https://github.com/alphacep/vosk-api/releases/download/v$(VOSK_VERSION)/vosk-osx-$(VOSK_VERSION).zip
	unzip -o /tmp/vosk-osx.zip -d /tmp
	cp /tmp/vosk-osx-$(VOSK_VERSION)/* $(LIB_DIR)/
	rm -rf /tmp/vosk-osx.zip /tmp/vosk-osx-$(VOSK_VERSION)
	@echo "Downloading vosk speech model..."
	curl -L -o /tmp/vosk-model.zip \
		https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip
	unzip -o /tmp/vosk-model.zip -d $(MODELS_DIR)
	mv $(MODELS_DIR)/vosk-model-small-en-us-0.15 $(MODELS_DIR)/vosk-model-small-en-us 2>/dev/null || true
	rm -f /tmp/vosk-model.zip
	@echo "Vosk setup complete"

# Full setup (whisper + vosk)
setup: setup-whisper setup-vosk

# Run tests
test:
	go test ./...

$(MODELS_DIR):
	mkdir -p $(MODELS_DIR)

$(LIB_DIR):
	mkdir -p $(LIB_DIR)

$(LIB_DIR)/libvosk.dylib:
	@echo "Error: libvosk.dylib not found. Run 'make setup-vosk' first."
	@exit 1

clean:
	rm -f yapatype
