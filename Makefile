# Define output directories
OUTPUT_DIR = binaries
RELEASE_DIR = release

# Define binary names
BINARY_NAME = rip

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
	@echo "Building for macOS arm64..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/mac/$(BINARY_NAME)-arm64 .
	@echo "Building for macOS amd64..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/mac/$(BINARY_NAME)-amd64 .

# Build for Windows (both amd64 and arm64)
windows:
	@mkdir -p $(OUTPUT_DIR)/windows
	@echo "Building for Windows amd64..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/windows/$(BINARY_NAME)-amd64.exe .
	@echo "Building for Windows arm64..."
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/windows/$(BINARY_NAME)-arm64.exe .

# Build for Linux (both amd64 and arm64)
linux:
	@mkdir -p $(OUTPUT_DIR)/linux
	@echo "Building for Linux amd64..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/linux/$(BINARY_NAME)-amd64 .
	@echo "Building for Linux arm64..."
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/linux/$(BINARY_NAME)-arm64 .

# Build release binaries and create git tag
release: clean all
	@mkdir -p $(RELEASE_DIR)
	@echo "Copying release binaries..."
	@cp $(OUTPUT_DIR)/mac/$(BINARY_NAME)-amd64 $(RELEASE_DIR)/$(BINARY_NAME)-mac-amd64
	@cp $(OUTPUT_DIR)/mac/$(BINARY_NAME)-arm64 $(RELEASE_DIR)/$(BINARY_NAME)-mac-arm64
	@cp $(OUTPUT_DIR)/linux/$(BINARY_NAME)-amd64 $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64
	@cp $(OUTPUT_DIR)/linux/$(BINARY_NAME)-arm64 $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64
	@cp $(OUTPUT_DIR)/windows/$(BINARY_NAME)-amd64.exe $(RELEASE_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@cp $(OUTPUT_DIR)/windows/$(BINARY_NAME)-arm64.exe $(RELEASE_DIR)/$(BINARY_NAME)-windows-arm64.exe
	@echo "Release binaries created in $(RELEASE_DIR)/"
	@ls -lh $(RELEASE_DIR)/
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
