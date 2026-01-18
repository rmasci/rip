# Define output directory
OUTPUT_DIR = binaries
RELEASE_DIR = release

# Define binary name
BINARY_NAME = rip

# Define platforms
PLATFORMS = darwin/amd64 linux/amd64 windows/amd64

# Default target
all: clean mac windows linux

# Build for macOS
mac:
	@mkdir -p $(OUTPUT_DIR)/mac
	GOOS=darwin GOARCH=arm64 go build -o $(OUTPUT_DIR)/mac/$(BINARY_NAME) .

# Build for Windows
windows:
	@mkdir -p $(OUTPUT_DIR)/windows
	GOOS=windows GOARCH=amd64 go build -o $(OUTPUT_DIR)/windows/$(BINARY_NAME).exe .

# Build for Linux
linux:
	@mkdir -p $(OUTPUT_DIR)/linux
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT_DIR)/linux/$(BINARY_NAME) .

# Build release binaries
release: clean all
	@mkdir -p $(RELEASE_DIR)
	@cp $(OUTPUT_DIR)/linux/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)-linux
	@cp $(OUTPUT_DIR)/mac/$(BINARY_NAME) $(RELEASE_DIR)/$(BINARY_NAME)-mac
	@cp $(OUTPUT_DIR)/windows/$(BINARY_NAME).exe $(RELEASE_DIR)/$(BINARY_NAME)-windows.exe
	@echo "Release binaries created in $(RELEASE_DIR)/"

# Clean up binaries
clean:
	@rm -rf $(OUTPUT_DIR)
	@echo "Cleaned up binaries."