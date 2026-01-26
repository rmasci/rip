package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// webCmd represents the `web` command for starting the web frontend
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web frontend for DVD ripping",
	Long: `The "web" command starts a daemon that provides a web interface for DVD ripping operations.
It automatically detects available DVD drives and allows you to manage ripping jobs through a browser.`,
	Args: cobra.NoArgs,
	Run:  startWebServer,
}

// webConfig holds the configuration for the web server
type webConfig struct {
	Port        int
	StoragePath string
	RipCommand  string
}

var webCfg webConfig

// init registers the web command with the root command and configures its flags.
func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Port to run the web server on")
	webCmd.Flags().StringVar(&webCfg.StoragePath, "storage", expandHome("~/Videos"), "Storage path for ripped media")
	webCmd.Flags().StringVar(&webCfg.RipCommand, "rip", "rip", "Path to the rip CLI command")
	rootCmd.AddCommand(webCmd)
}

// expandHome expands ~ to home directory
func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// startWebServer initializes and starts the web server
func startWebServer(cmd *cobra.Command, args []string) {
	port, _ := cmd.Flags().GetInt("port")
	webCfg.StoragePath = expandHome(webCfg.StoragePath)

	fmt.Printf("Starting DVD Ripper Web Server\n")
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Storage Path: %s\n", webCfg.StoragePath)
	fmt.Printf("Open http://localhost:%d in your browser\n\n", port)

	// Setup routes
	http.HandleFunc("/api/devices", handleWebDevices)
	http.HandleFunc("/api/categories", handleWebCategories)
	http.HandleFunc("/api/rip", handleWebRip)
	http.HandleFunc("/api/status", handleWebStatus)

	// Serve static files
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Web server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// handleWebDevices returns available DVD devices
func handleWebDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	devices := findWebDVDDevices()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
	})
}

// findWebDVDDevices discovers /dev/sr* (Linux) or /dev/rdisk* (macOS) devices
func findWebDVDDevices() []string {
	var devices []string

	// Try Linux devices first
	matches, err := filepath.Glob("/dev/sr*")
	if err == nil && len(matches) > 0 {
		devices = append(devices, matches...)
	}

	// Try macOS devices
	matches, err = filepath.Glob("/dev/rdisk*")
	if err == nil && len(matches) > 0 {
		devices = append(devices, matches...)
	}

	// If no devices found, return empty list
	if len(devices) == 0 {
		devices = []string{}
	}

	return devices
}

// handleWebCategories manages category operations (GET, POST, PUT, DELETE)
func handleWebCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		handleWebGetCategories(w, r)
	case "POST":
		handleWebCreateCategory(w, r)
	case "PUT":
		handleWebRenameCategory(w, r)
	case "DELETE":
		handleWebDeleteCategory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWebGetCategories returns all categories
func handleWebGetCategories(w http.ResponseWriter, r *http.Request) {
	categories := getWebCategories()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"categories": categories,
	})
}

// handleWebCreateCategory creates a new category
func handleWebCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Category name cannot be empty", http.StatusBadRequest)
		return
	}

	if err := createWebCategory(req.Name); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create category: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "created",
		"name":   req.Name,
	})
}

// handleWebRenameCategory renames an existing category
func handleWebRenameCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldName string `json:"oldName"`
		NewName string `json:"newName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.OldName == "" || req.NewName == "" {
		http.Error(w, "Old and new names are required", http.StatusBadRequest)
		return
	}

	if err := renameWebCategory(req.OldName, req.NewName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to rename category: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "renamed",
	})
}

// handleWebDeleteCategory deletes a category
func handleWebDeleteCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	if err := deleteWebCategory(req.Name); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete category: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})
}

// getWebCategories returns all categories from storage path
func getWebCategories() []string {
	var categories []string

	entries, err := os.ReadDir(webCfg.StoragePath)
	if err != nil {
		log.Printf("Error reading storage directory: %v", err)
		return categories
	}

	for _, entry := range entries {
		if entry.IsDir() {
			categories = append(categories, entry.Name())
		}
	}

	return categories
}

// createWebCategory creates a new category directory
func createWebCategory(name string) error {
	catPath := filepath.Join(webCfg.StoragePath, name)
	return os.MkdirAll(catPath, 0755)
}

// renameWebCategory renames an existing category
func renameWebCategory(oldName, newName string) error {
	oldPath := filepath.Join(webCfg.StoragePath, oldName)
	newPath := filepath.Join(webCfg.StoragePath, newName)
	return os.Rename(oldPath, newPath)
}

// deleteWebCategory deletes a category (if empty)
func deleteWebCategory(name string) error {
	catPath := filepath.Join(webCfg.StoragePath, name)
	return os.Remove(catPath)
}

// handleWebRip processes the rip request from the web frontend
func handleWebRip(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Device   string `json:"device"`
		Category string `json:"category"`
		Movie    string `json:"movie"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Device == "" || req.Category == "" || req.Movie == "" {
		http.Error(w, "Missing required fields: device, category, movie", http.StatusBadRequest)
		return
	}

	// Execute the rip in a goroutine to avoid blocking the response
	go executeWebRipJob(req.Device, req.Category, req.Movie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "rip started",
		"movie":  req.Movie,
	})
}

// handleWebStatus returns the current status of the application
func handleWebStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":           "running",
		"storagePath":      webCfg.StoragePath,
		"ripCommand":       webCfg.RipCommand,
		"availableDevices": findWebDVDDevices(),
	})
}

// executeWebRipJob executes the actual rip operation
// This runs in a background goroutine
func executeWebRipJob(device, category, movie string) {
	log.Printf("Starting rip job: device=%s, category=%s, movie=%s\n", device, category, movie)

	// Build arguments for the rip dvd command
	args := []string{
		"dvd",
		"-d", device,
		"-c", category,
		"-m", movie,
	}

	// Execute the rip command
	cmd := exec.Command(webCfg.RipCommand, args...)

	// Optional: Redirect output to log files or handle as needed
	if err := cmd.Run(); err != nil {
		log.Printf("Error executing rip command: %v\n", err)
	} else {
		log.Printf("Rip job completed successfully for: %s\n", movie)
	}
}
