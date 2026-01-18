package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/bitfield/script"
	"github.com/spf13/cobra"
)

// dvdCmd represents the `dvd` command for ripping DVDs.
// It automates the process of ripping DVDs using MakeMKV,
// organizing them by category, fetching metadata, and renaming files.
// The command supports discovering movie names from disc metadata or accepting them as a flag.
var dvdCmd = &cobra.Command{
	Use:   "dvd",
	Short: "Rip DVDs using MakeMKV and organize by category",
	Long: `The "dvd" command automates the ripping of DVDs using MakeMKV,
categorizing them for use with media libraries like Plex. It requires
you to provide a physical device path, a category, and optionally a movie name.`,
	Args: cobra.NoArgs, // No non-flag arguments are required
	Run:  dvdrip,
}

// dvdrip executes the DVD ripping workflow.
// It performs the following steps:
// 1. Discovers or accepts the movie name from the DVD or user input
// 2. Uses FileBot to look up the correct movie name and year (with fallback to user input)
// 3. Validates the MergerFS mountpoint
// 4. Fetches metadata from TheMovieDB via FileBot
// 5. Creates the output directory structure with CamelCase naming
// 6. Executes MakeMKV to rip the longest title
// 7. Renames the movie file using FileBot
// 8. Ejects the disc
// 9. Displays completion summary
func dvdrip(cmd *cobra.Command, args []string) {
	// Parse command-line flags
	device, _ := cmd.Flags().GetString("device")
	category, _ := cmd.Flags().GetString("category")
	movie, _ := cmd.Flags().GetString("movie")

	// Determine movie name from one of three sources (in priority order):
	// 1. Explicit -m flag provided by user
	// 2. Command-line argument (if provided)
	// 3. Discovered from DVD metadata using MakeMKV
	var query string
	if movie != "" {
		// User provided movie name via -m flag
		query = movie
		fmt.Printf("Using provided movie name: %s\n", query)
	} else if len(args) > 0 {
		// Movie name provided as command-line argument
		query = args[0]
	} else {
		// Attempt to discover movie name from DVD
		fmt.Println("Discovering movie name from DVD...")
		query = discoverMovieName(device)
		if query == "" {
			log.Fatal("Error: Could not discover movie name from DVD. Please provide it manually using the -m flag.")
		}
		fmt.Printf("Discovered movie name: %s\n", query)
	}

	// Validate that category flag was provided
	if category == "" {
		log.Fatalf("Error: Target category must be provided.\n\nUsage:\n")
	}

	// Step 1: Verify MergerFS pool is mounted before proceeding
	if !isMountpoint("/plex/storage") {
		log.Fatal("Error: /plex/storage is not a mountpoint. Check MergerFS!")
	}

	// Step 2: Try to look up the correct movie name and year using FileBot
	// Format: Movie Name (Year)
	fmt.Printf("Looking up movie info in TMDB for: %s...\n", query)
	finalName := fetchMetadata(query, "{n} ({y})")
	if finalName == "" {
		// Fallback to user-provided name if FileBot lookup fails
		fmt.Printf("Warning: Could not find movie in TMDB, using provided name: %s\n", query)
		finalName = query
	} else {
		fmt.Printf("Found: %s\n", finalName)
	}

	// Step 3: Create output directory structure with CamelCase naming
	// Directory format: /plex/storage/Category/MovieNameYear/ (no spaces, camelCase)
	baseDir := "/plex/storage"
	dirName := toCamelCase(finalName)
	outDir := filepath.Join(baseDir, category, dirName)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}
	fmt.Printf("Putting movie in %s/%s\n", outDir, finalName)

	// Step 4: Extract drive information and format for MakeMKV
	driveIndex := extractDriveIndex(device)
	drive := fmt.Sprintf("disc:%s", driveIndex)

	fmt.Printf("Target: %s/%s.mkv\n", outDir, finalName)

	// Step 5: Execute MakeMKV rip operation (rips the longest title)
	if err := runDVDMakeMKV(drive, outDir); err != nil {
		fmt.Println("Longest failed, attempting Stage 2...")
		// Add fallback logic here (Stage 2, 3, etc.)
	}

	// Step 6: Rename movie file with metadata from TheMovieDB
	fmt.Println("Renaming movie file with proper name from FileBot...")
	if err := renameMovieWithFileBot(finalName, outDir); err != nil {
		fmt.Printf("Warning: FileBot rename failed: %v\n", err)
	}

	// Step 7: Eject the disc from the drive
	devicePath := extractDevicePath(drive)
	if err := ejectDisc(devicePath); err != nil {
		fmt.Printf("Warning: Could not eject disc: %v\n", err)
	}

	// Step 8: Display completion summary
	fmt.Println("-------------------------------------------------------")
	fmt.Println("RIP COMPLETE!")
	fmt.Printf("Files are in: %s\n", outDir)
}

// init registers the dvd command with the root command and configures its flags.
func init() {
	// Define command-line flags
	dvdCmd.Flags().StringP("device", "d", "/dev/sr0", "Physical device path (e.g. /dev/sr0)")
	dvdCmd.Flags().StringP("category", "c", "", "Target category folder (e.g. Comedy, Action)")
	dvdCmd.Flags().StringP("movie", "m", "", "Movie name to bypass discovery and use directly")

	// Register the dvd command as a subcommand of the root command
	rootCmd.AddCommand(dvdCmd)
}

// isMountpoint checks if the specified path is a valid mountpoint using the mountpoint command.
// Returns true if the path is a mountpoint, false otherwise.
func isMountpoint(path string) bool {
	// Use the mountpoint command with -q (quiet) flag
	// Returns nil (exit code 0) if path is a mountpoint, error otherwise
	cmd := exec.Command("mountpoint", "-q", path)
	return cmd.Run() == nil
}

// fetchMetadata queries FileBot to retrieve metadata from TheMovieDB database.
// It uses the provided query string and format string to look up and format movie information.
//
// Parameters:
//
//	query - the search query (movie name)
//	format - the FileBot format string for output (e.g., "{n} ({y})")
//
// Returns the formatted metadata string or empty string if lookup fails.
func fetchMetadata(query, format string) string {
	// Execute FileBot list command to query TheMovieDB
	p := script.Exec(fmt.Sprintf("filebot -list --db TheMovieDB --q '%s' --format '%s'", query, format))
	out, err := p.String()
	if err != nil {
		log.Printf("Error fetching metadata: %v\n", err)
		return ""
	}
	// Parse the first line of output as the metadata result
	lines := strings.Split(out, "\n")
	return strings.TrimSpace(lines[0])
}

// extractDriveIndex extracts the numeric drive index from a device path.
// For example, converts "/dev/sr0" to "0", "/dev/sr1" to "1", etc.
// This index is used to identify the disc in MakeMKV commands (e.g., "disc:0").
func extractDriveIndex(devicePath string) string {
	// Use regex to find all numeric digits in the device path
	re := regexp.MustCompile(`[0-9]+`)
	return re.FindString(devicePath)
}

// runDVDMakeMKV executes the MakeMKV command to rip the longest title from a DVD.
// It first queries the disc to identify all available titles, finds the longest one,
// then executes the rip operation. After ripping completes, it ejects the disc.
//
// Parameters:
//
//	drive - the disc specification (e.g., "disc:0")
//	outDir - the output directory where the MKV file will be saved
//
// Returns an error if the makemkvcon command fails.
func runDVDMakeMKV(drive, outDir string) error {
	// Extract device path from the disc specification for ejection after ripping
	devicePath := extractDevicePath(drive)
	defer ejectDisc(devicePath)

	// Step 1: Query the disc to get information about all available titles
	// Uses the -r flag for robot mode (machine-readable output)
	p := script.Exec(fmt.Sprintf("makemkvcon -r info %s", drive))
	infoOutput, err := p.String()
	if err != nil {
		return fmt.Errorf("error running makemkvcon info: %v", err)
	}

	// Step 2: Parse the output to identify all titles and find the longest one
	// TINFO line format: TINFO:title_id,27,duration_in_seconds,"duration_in_ms"
	re := regexp.MustCompile(`TINFO:(\d+),27,\d+,"(\d+)"`)
	var longestTitleID string
	var maxDuration int
	for _, line := range strings.Split(infoOutput, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			titleID := matches[1]
			duration, _ := strconv.Atoi(matches[2])
			// Keep track of the title with the maximum duration
			if duration > maxDuration {
				maxDuration = duration
				longestTitleID = titleID
			}
		}
	}

	// Step 3: Run the mkv rip command with the longest title ID, or fall back to title 0
	titleID := longestTitleID
	if titleID == "" {
		titleID = "0"
	}

	// Execute makemkvcon mkv command to rip the longest title
	// --minlength=3600 ensures we only rip titles longer than 1 hour (for movies)
	return script.Exec(fmt.Sprintf("makemkvcon mkv %s %s %s --minlength=3600", drive, titleID, outDir)).Error()
}

// discoverMovieName attempts to extract the movie title from the DVD disc metadata using MakeMKV.
// It queries the disc information and parses the output to extract the disc title.
//
// Parameters:
//
//	devicePath - the device path of the DVD drive (e.g., "/dev/sr0")
//
// Returns the discovered movie name or empty string if discovery fails.
func discoverMovieName(devicePath string) string {
	// Extract the drive index from the device path to format for MakeMKV
	driveIndex := extractDriveIndex(devicePath)
	drive := fmt.Sprintf("disc:%s", driveIndex)

	// Query disc information using makemkvcon with robot mode (-r) output
	p := script.Exec(fmt.Sprintf("makemkvcon -r info %s", drive))
	out, err := p.String()
	if err != nil {
		log.Printf("Error running makemkvcon: %v", err)
		return ""
	}

	// Parse the output to extract the movie name
	// CINFO line format: CINFO:2,0,"movie_title"
	// where: 2 = disc, 0 = disc title field
	re := regexp.MustCompile(`(?m)^CINFO:2,0,"(.+)"`)
	matches := re.FindStringSubmatch(out)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// renameMovieWithFileBot uses FileBot to rename a movie file with proper name from TheMovieDB database.
// It renames files recursively in the output directory using a Plex-compatible naming format.
//
// Parameters:
//
//	movieName - the movie name (currently unused, kept for potential future use)
//	outDir - the directory containing the movie file to rename
//
// The format string produces names like: "Movie Title (Year)"
//
// Returns an error if the FileBot rename command fails.
func renameMovieWithFileBot(movieName, outDir string) error {
	// FileBot rename command format for movies
	// Format string: {n} ({y})
	// where: n=movie name, y=year
	renameFormat := "{n} ({y})"

	// Execute FileBot rename command with --action move to actually rename files
	// Uses TheMovieDB database for metadata lookup
	return script.Exec(fmt.Sprintf("filebot -rename %s -r --db TheMovieDB --format '%s' --action move", outDir, renameFormat)).Error()
}

// toCamelCase converts a string to CamelCase with no spaces.
// It removes all spaces and special characters (except alphanumeric), and converts to PascalCase.
// For example:
//
//	"The Matrix (1999)" -> "TheMatrix1999"
//	"Inception" -> "Inception"
//	"Star Wars: A New Hope (1977)" -> "StarWarsANewHope1977"
//
// Parameters:
//
//	s - the string to convert
//
// Returns the CamelCase version of the string with no spaces.
func toCamelCase(s string) string {
	// Remove special characters and split on spaces
	var result strings.Builder
	words := strings.FieldsFunc(s, func(r rune) bool {
		// Split on spaces and special characters (keep only alphanumeric)
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	// Capitalize first letter of each word
	for _, word := range words {
		if len(word) > 0 {
			// Capitalize first rune, keep rest as-is
			result.WriteRune(unicode.ToUpper(rune(word[0])))
			result.WriteString(word[1:])
		}
	}

	return result.String()
}
