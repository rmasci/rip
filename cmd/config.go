package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Config holds the application configuration
type Config struct {
	StoragePath string // Path where ripped media will be stored
}

// LoadConfig loads the configuration from ~/.rip.conf or creates it if it doesn't exist
func LoadConfig() *Config {
	configPath := getConfigPath()

	config := &Config{
		StoragePath: "/plex/storage", // Default value
	}

	// Try to read existing config file
	if _, err := os.Stat(configPath); err == nil {
		content, err := os.ReadFile(configPath)
		if err != nil {
			log.Printf("Warning: Could not read config file: %v\n", err)
			return config
		}

		// Parse config file
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse key=value pairs
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if key == "storage_path" {
					// Expand ~ to home directory
					if strings.HasPrefix(value, "~/") {
						home, _ := os.UserHomeDir()
						value = filepath.Join(home, value[2:])
					}
					config.StoragePath = value
				}
			}
		}
		return config
	}

	// Config file doesn't exist, create it with defaults
	createDefaultConfig(configPath, config)
	return config
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	return filepath.Join(home, ".rip.conf")
}

// createDefaultConfig creates a default config file
func createDefaultConfig(configPath string, config *Config) {
	content := `# rip configuration file
# This file is automatically created if it doesn't exist

# Storage path where ripped media will be organized
# Default: /plex/storage (your MergerFS mount point)
# You can change this to any directory where you want media stored
# Example: /mnt/media or ~/Videos/Rips
storage_path=/plex/storage
`

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		log.Printf("Warning: Could not create default config file at %s: %v\n", configPath, err)
		return
	}

	fmt.Printf("Created default config file at: %s\n", configPath)
	fmt.Printf("You can edit this file to customize your storage location\n")
}

// VerifyStoragePath checks if the storage path exists and is writable
func VerifyStoragePath(storagePath string) error {
	// Check if directory exists
	info, err := os.Stat(storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("storage path does not exist: %s", storagePath)
		}
		return fmt.Errorf("error accessing storage path: %v", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("storage path is not a directory: %s", storagePath)
	}

	// Check if it's writable by trying to create a test file
	testFile := filepath.Join(storagePath, ".rip_test")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		return fmt.Errorf("storage path is not writable: %v", err)
	}
	os.Remove(testFile)

	return nil
}

// GetMergerFSDisks reads /etc/fstab and returns the list of disks in the MergerFS pool
// that mounts at the specified mount point (e.g., /plex/storage)
// Returns empty slice if MergerFS is not found for that mount point
func GetMergerFSDisks() []string {
	return getMergerFSDisksForPath(AppConfig.StoragePath)
}

// getMergerFSDisksForPath reads /etc/fstab and returns the list of disks in the MergerFS pool
// that mounts at the specified path. Returns empty slice if not found.
func getMergerFSDisksForPath(mountPath string) []string {
	fstabPath := "/etc/fstab"
	content, err := os.ReadFile(fstabPath)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse fstab line: device mountpoint fstype options dump pass
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		device := parts[0]
		mountpoint := parts[1]
		fstype := parts[2]

		// Check if this line matches our mount point and is mergerfs
		if mountpoint == mountPath && strings.Contains(fstype, "mergerfs") {
			// Format: /mnt/disk1:/mnt/disk2:/mnt/disk3 /plex/storage fuse.mergerfs ...
			// Split the disk list by colons
			disks := strings.Split(device, ":")
			return disks
		}
	}

	return nil
}

// DiskSpace represents information about a disk's available space
type DiskSpace struct {
	Path      string
	Available uint64
	Total     uint64
}

// GetDiskWithMostSpace returns the mount point with the most available free space
// from a list of disk mount points. Returns an error if no disk has at least 5GB free.
func GetDiskWithMostSpace(diskPaths []string) (string, error) {
	if len(diskPaths) == 0 {
		return "", fmt.Errorf("no disk paths provided")
	}

	const minFreeSpaceGB = 5
	const minFreeSpaceBytes = minFreeSpaceGB * 1024 * 1024 * 1024

	var maxSpace uint64
	var selectedDisk string

	for _, diskPath := range diskPaths {
		// Get filesystem stats
		var stat syscall.Statfs_t
		err := syscall.Statfs(diskPath, &stat)
		if err != nil {
			// Skip disks that can't be accessed
			fmt.Printf("Warning: Could not stat %s: %v\n", diskPath, err)
			continue
		}

		// Calculate available space
		available := stat.Bavail * uint64(stat.Bsize)

		if available > maxSpace {
			maxSpace = available
			selectedDisk = diskPath
		}
	}

	if selectedDisk == "" {
		return "", fmt.Errorf("could not determine disk with most space")
	}

	// Check if the disk with most space has at least 5GB free
	if maxSpace < minFreeSpaceBytes {
		availableGB := float64(maxSpace) / (1024 * 1024 * 1024)
		return "", fmt.Errorf("insufficient free space: only %.2f GB available, need at least %d GB", availableGB, minFreeSpaceGB)
	}

	// Convert to GB for display
	availableGB := float64(maxSpace) / (1024 * 1024 * 1024)
	fmt.Printf("Disk with most space: %s (%.2f GB available)\n", selectedDisk, availableGB)

	return selectedDisk, nil
}
