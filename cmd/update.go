package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rmasci/script"
	"github.com/spf13/cobra"
)

// updateCmd represents the `update` command for updating MakeMKV
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update MakeMKV to the latest version",
	Long: `Updates MakeMKV to the latest available version.
This command can be run regularly via cron to keep MakeMKV current.

Usage for cron (weekly):
  0 2 * * 0 /usr/local/bin/rip update >> /var/log/rip-update.log 2>&1`,
	Args: cobra.NoArgs,
	Run:  updateMakeMKV,
}

// updateMakeMKV downloads and installs the latest version of MakeMKV
func updateMakeMKV(_ *cobra.Command, _ []string) {
	fmt.Println("Checking for the latest MakeMKV version...")

	// Create working directory in /tmp for temporary build files
	workDir := "/tmp/makemkv"
	if err := os.MkdirAll(workDir, 0755); err != nil {
		log.Fatalf("Error creating work directory: %v", err)
	}

	// Get the latest version
	latestVersion := getLatestMakeMKVVersion()
	if latestVersion == "" {
		log.Fatal("Could not determine latest MakeMKV version. Please visit https://www.makemkv.com/download/")
	}

	fmt.Printf("Latest MakeMKV version: %s\n", latestVersion)

	// Check if already installed
	if isMakeMKVVersionInstalled(latestVersion) {
		fmt.Printf("MakeMKV %s is already installed.\n", latestVersion)
		return
	}

	fmt.Printf("Downloading MakeMKV %s...\n", latestVersion)

	// Download files
	ossFile := filepath.Join(workDir, fmt.Sprintf("makemkv-oss-%s.tar.gz", latestVersion))
	binFile := filepath.Join(workDir, fmt.Sprintf("makemkv-bin-%s.tar.gz", latestVersion))

	if err := downloadFile(fmt.Sprintf("https://www.makemkv.com/download/makemkv-oss-%s.tar.gz", latestVersion), ossFile); err != nil {
		log.Fatalf("Error downloading MakeMKV OSS: %v", err)
	}

	if err := downloadFile(fmt.Sprintf("https://www.makemkv.com/download/makemkv-bin-%s.tar.gz", latestVersion), binFile); err != nil {
		log.Fatalf("Error downloading MakeMKV bin: %v", err)
	}

	fmt.Println("Extracting and building MakeMKV OSS...")
	if err := buildMakeMKV(workDir, ossFile, "makemkv-oss"); err != nil {
		log.Fatalf("Error building MakeMKV OSS: %v", err)
	}

	fmt.Println("Extracting and building MakeMKV bin...")
	if err := buildMakeMKV(workDir, binFile, "makemkv-bin"); err != nil {
		log.Fatalf("Error building MakeMKV bin: %v", err)
	}

	fmt.Println("Verifying installation...")
	if err := verifyMakeMKVInstallation(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	fmt.Printf("MakeMKV updated successfully to version %s!\n", latestVersion)
}

// getLatestMakeMKVVersion fetches the latest version number from the MakeMKV website
func getLatestMakeMKVVersion() string {
	p := script.Exec("curl -s https://www.makemkv.com/download/")
	out, err := p.String()
	if err != nil {
		return ""
	}

	// Extract version from the HTML using regex
	re := regexp.MustCompile(`makemkv-oss-(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// isMakeMKVVersionInstalled checks if the specified version is already installed
func isMakeMKVVersionInstalled(version string) bool {
	p := script.Exec("makemkvcon info disc:0 2>&1")
	out, err := p.String()
	if err != nil {
		return false
	}

	// Check if version appears in output
	return strings.Contains(out, fmt.Sprintf("v%s", version))
}

// downloadFile downloads a file from a URL to the specified path
func downloadFile(url, filepath string) error {
	// Skip if file already exists
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("File already exists: %s\n", filepath)
		return nil
	}

	cmd := exec.Command("wget", "-q", url, "-O", filepath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wget failed: %v", err)
	}

	return nil
}

// buildMakeMKV extracts and builds MakeMKV from a tar.gz file
func buildMakeMKV(workDir, tarFile, pkgName string) error {
	// Extract tar file
	cmd := exec.Command("tar", "xzf", tarFile, "-C", workDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tar extraction failed: %v", err)
	}

	// Find the extracted directory
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	var buildDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), pkgName) && !strings.HasSuffix(entry.Name(), ".tar.gz") {
			buildDir = filepath.Join(workDir, entry.Name())
			break
		}
	}

	if buildDir == "" {
		return fmt.Errorf("could not find extracted directory for %s", pkgName)
	}

	// Configure
	configCmd := exec.Command("bash", "-c", "cd "+buildDir+" && ./configure")
	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("configure failed: %v", err)
	}

	// Build
	makeCmd := exec.Command("bash", "-c", "cd "+buildDir+" && make")
	if err := makeCmd.Run(); err != nil {
		return fmt.Errorf("make failed: %v", err)
	}

	// Install (with sudo)
	installCmd := exec.Command("bash", "-c", "cd "+buildDir+" && sudo make install")
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("make install failed: %v", err)
	}

	return nil
}

// verifyMakeMKVInstallation checks if MakeMKV is working
func verifyMakeMKVInstallation() error {
	p := script.Exec("makemkvcon info disc:0 2>&1")
	out, err := p.String()
	if err != nil {
		return fmt.Errorf("makemkvcon verification failed: %v", err)
	}

	// Check for the version line in output
	lines := strings.Split(out, "\n")
	if len(lines) > 0 && strings.Contains(lines[0], "MakeMKV") {
		fmt.Println(lines[0])
		return nil
	}

	return fmt.Errorf("unexpected makemkvcon output")
}

// init registers the update command with the root command
func init() {
	rootCmd.AddCommand(updateCmd)
}
