# Define output directories
OUTPUT_DIR = binaries
RELEASE_DIR = release

# Define binary names
CLI_BINARY = rip

# Get the current version from git tags or use default
CURRENT_VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")

# Get the next version using a shell script
NEW_VERSION = $(shell bash scripts/next-version.sh)

# Build flags with version injection
LDFLAGS = -ldflags "-X github.com/rmasci/dvdrip/cmd.Version=$(NEW_VERSION)"

# Default target
all: clean mac windows linux

# Build for macOS (both amd64 and arm64)
mac:
	@mkdir -p $(OUTPUT_DIR)/mac
	@echo "Building $(CLI_BINARY) for macOS arm64..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/mac/$(CLI_BINARY)-arm64 .
	@echo "Building $(CLI_BINARY) for macOS amd64..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/mac/$(CLI_BINARY)-amd64 .

# Build for Windows (both amd64 and arm64)
windows:
	@mkdir -p $(OUTPUT_DIR)/windows
	@echo "Building $(CLI_BINARY) for Windows amd64..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/windows/$(CLI_BINARY)-amd64.exe .
	@echo "Building $(CLI_BINARY) for Windows arm64..."
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/windows/$(CLI_BINARY)-arm64.exe .


# Build for Linux (both amd64 and arm64)
linux:
	@mkdir -p $(OUTPUT_DIR)/linux
	@echo "Building $(CLI_BINARY) for Linux amd64..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/linux/$(CLI_BINARY)-amd64 .
	@echo "Building $(CLI_BINARY) for Linux arm64..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/linux/$(CLI_BINARY)-arm64 .


# Build release binaries and create git tag
release: clean all
	@mkdir -p $(RELEASE_DIR)
	@echo "Copying release binaries..."
	@cp $(OUTPUT_DIR)/mac/$(CLI_BINARY)-amd64 $(RELEASE_DIR)/$(CLI_BINARY)-mac-amd64
	@cp $(OUTPUT_DIR)/mac/$(CLI_BINARY)-arm64 $(RELEASE_DIR)/$(CLI_BINARY)-mac-arm64
	@cp $(OUTPUT_DIR)/linux/$(CLI_BINARY)-amd64 $(RELEASE_DIR)/$(CLI_BINARY)-linux-amd64
	@cp $(OUTPUT_DIR)/linux/$(CLI_BINARY)-arm64 $(RELEASE_DIR)/$(CLI_BINARY)-linux-arm64
	@cp $(OUTPUT_DIR)/windows/$(CLI_BINARY)-amd64.exe $(RELEASE_DIR)/$(CLI_BINARY)-windows-amd64.exe
	@cp $(OUTPUT_DIR)/windows/$(CLI_BINARY)-arm64.exe $(RELEASE_DIR)/$(CLI_BINARY)-windows-arm64.exe

	@echo "Release binaries created in $(RELEASE_DIR)/"
	@ls -lh $(RELEASE_DIR)/ | grep -E "rip"
	@echo ""
	@echo "Creating git tag $(NEW_VERSION)..."
	@git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"
	@git push origin $(NEW_VERSION)
	@echo "Git tag $(NEW_VERSION) created and pushed"
	@echo "Built with version: $(NEW_VERSION)"

# Clean up binaries
clean:
	@rm -rf $(OUTPUT_DIR) $(RELEASE_DIR)
	@echo "Cleaned up binaries and release directories."

# Show current and next version
version-info:
	@echo "Current version: $(CURRENT_VERSION)"
	@echo "Next version: $(NEW_VERSION)"

# Update MakeMKV to the latest version
update-makemkv:
	@echo "Checking for the latest MakeMKV version..."
	@mkdir -p ~/src/makemkv && cd ~/src/makemkv && \
	LATEST_VERSION=$$(curl -s https://www.makemkv.com/download/ | grep -oP 'makemkv-oss-\K[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
	if [ -z "$$LATEST_VERSION" ]; then \
		echo "Could not determine latest version. Please visit https://www.makemkv.com/download/"; \
		exit 1; \
	fi; \
	echo "Latest MakeMKV version: $$LATEST_VERSION"; \
	echo "Downloading MakeMKV $$LATEST_VERSION..."; \
	wget -q https://www.makemkv.com/download/makemkv-oss-$$LATEST_VERSION.tar.gz && \
	wget -q https://www.makemkv.com/download/makemkv-bin-$$LATEST_VERSION.tar.gz && \
	echo "Extracting and building MakeMKV OSS..."; \
	tar xzf makemkv-oss-$$LATEST_VERSION.tar.gz && \
	cd makemkv-oss-$$LATEST_VERSION && \
	./configure > /dev/null 2>&1 && \
	make > /dev/null 2>&1 && \
	sudo make install > /dev/null 2>&1 && \
	cd .. && \
	echo "Extracting and building MakeMKV bin..."; \
	tar xzf makemkv-bin-$$LATEST_VERSION.tar.gz && \
	cd makemkv-bin-$$LATEST_VERSION && \
	./configure > /dev/null 2>&1 && \
	make > /dev/null 2>&1 && \
	sudo make install > /dev/null 2>&1 && \
	cd .. && \
	echo "Verifying installation..."; \
	makemkvcon -r info disc:0 2>&1 | head -1 && \
	echo "MakeMKV updated successfully to version $$LATEST_VERSION!"

