#!/bin/bash

# --- PHASE 0: Input Validation ---
QUERY="$1"
SEASON="$2" # Expects a number like 1

if [ -z "$QUERY" ] || [ -z "$SEASON" ]; then
    echo "Usage: ./tvrip.sh 'Show Name' SeasonNumber"
    echo "Example: ./tvrip.sh 'The Middle' 1"
    exit 1
fi

# Padding season number (e.g., 1 becomes 01)
printf -v S_PAD "%02d" "$SEASON"

# SAFETY CHECK: Ensure the mergerfs pool is actually mounted
if ! mountpoint -q /plex/storage; then
    echo "Error: /plex/storage is not mounted! Check mergerfs before ripping."
    exit 1
fi

echo "Searching TMDB for: $QUERY..."

# Get Path: Genre/Show Name (Year) {tmdb-ID}
# Using TheTVDB via FileBot for better TV accuracy
SMART_FOLDER=$(filebot -list --db TheTVDB --q "$QUERY" --format "{genre.toCamelCase()}/{n} ({y}) {tmdb-\$id}" | head -n 1)

if [ -z "$SMART_FOLDER" ]; then
    echo "Error: Could not find show matching '$QUERY' on TMDB/TVDB."
    exit 1
fi

# Define Paths
BASE_DIR="/plex/storage"
OUT_DIR="$BASE_DIR/$SMART_FOLDER/Season $S_PAD"

mkdir -p "$OUT_DIR"
echo "Target Folder: $OUT_DIR"

# --- CONFIGURATION ---
DRIVE="disc:0"
# Sitcom episodes are ~22 mins (1320s). 
# We set min to 15 mins (900s) and max to 50 mins (3000s).
MIN_LENGTH=900
MAX_LENGTH=3000

# --- THE RIP ---
echo "Ripping all titles longer than 15 minutes..."
# 'all' tells MakeMKV to grab every episode-length title on the disc
makemkvcon mkv "$DRIVE" all "$OUT_DIR" --minlength="$MIN_LENGTH --maxlen=3600"

# --- PHASE 2: "Play All" Cleanup ---
# This part scans the folder and removes the giant 'merged' track if it exists
echo "Cleaning up extra-long 'Play All' tracks..."
cd "$OUT_DIR" || exit

# Check for ffmpeg/ffprobe
if ! command -v ffprobe &> /dev/null; then
    echo "Warning: ffprobe not found. Skipping automatic Play-All cleanup."
else
    for f in *.mkv; do
        # Get duration in seconds
        duration=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$f")
        iduration=${duration%.*}
        
        # If the file is longer than 50 minutes, it's likely a Play-All or a movie
        if [ "$iduration" -gt "$MAX_LENGTH" ]; then
            echo "Removing giant/Play-All file: $f ($(Scale_Time $iduration))"
            rm "$f"
        fi
    done
fi

echo "-------------------------------------------------------"
echo "RIP COMPLETE!"
echo "Files are in: $OUT_DIR"
echo "Step 1: Verify episodes match S${S_PAD}E01, S${S_PAD}E02, etc."
echo "Step 2: Rename them so Jellyfin identifies them correctly."
echo "Step 3: Scan library in Jellyfin Dashboard."
