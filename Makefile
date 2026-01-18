# Define output directory
OUTPUT_DIR = binaries
RELEASE_DIR = release

# Define binary name
BINARY_NAME = rip

# Define platforms
PLATFORMS = darwin/amd64 linux/amd64 windows/amd64

# Get the current version from git tags or use default
CURRENT_VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")

# Get the next version using a shell script
NEW_VERSION = $(shell bash scripts/next-version.sh)

# Build flags with version injection
LDFLAGS = -ldflags "-X github.com/rmasci/dvdrip/cmd.Version=$(NEW_VERSION)"

# Default target
all: clean mac windows linux

# Build for macOS
mac:
	@mkdir -p $(OUTPUT_DIR)/mac
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/mac/$(BINARY_NAME) .

# Build for Windows
windows:
	@mkdir -p $(OUTPUT_DIR)/windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/windows/$(BINARY_NAME).exe .

# Build for Linux
linux:
	@mkdir -p $(OUTPUT_DIR)/linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/linux/$(BINARY_NAME) .

# Build release binaries and create git tag
release: clean all
	@mkdir -p $(RELEASE_DIR)
	@cp $(OUTPUT_DIR)/linux/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)-linux
	@cp $(OUTPUT_DIR)/mac/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)-mac
	@cp $(OUTPUT_DIR)/windows/$(BINARY_NAME).exe $(RELEASE_DIR)/$(BINARY_NAME)-windows.exe
	@echo "Release binaries created in $(RELEASE_DIR)/"
	@echo "Creating git tag $(NEW_VERSION)..."
	@git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"
	@git push origin $(NEW_VERSION)
	@echo "Git tag $(NEW_VERSION) created and pushed"
	@echo "Built with version: $(NEW_VERSION)"

# Clean up binaries
clean:
	@rm -rf $(OUTPUT_DIR)
	@echo "Cleaned up binaries."

# Show current and next version
version-info:
	@echo "Current version: $(CURRENT_VERSION)"
	@echo "Next version: $(NEW_VERSION)"
