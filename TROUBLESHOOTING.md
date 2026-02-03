# Troubleshooting Guide

This guide helps you solve common issues when using the `rip` command-line tool.

---

## Keeping MakeMKV Updated

**Problem:** MakeMKV version expires and you get "This application version is too old" errors

**Solution:** Use the built-in update command to automatically fetch and build the latest version

### Manual Update

Run this command anytime to update MakeMKV to the latest version:

```bash
rip update
```

This will:
1. Check for the latest MakeMKV version on the official website
2. Download both OSS and bin components
3. Build and install them automatically
4. Verify the installation

### Automatic Weekly Updates (via Cron)

Set up a weekly cron job to automatically keep MakeMKV current. Edit your crontab:

```bash
crontab -e
```

Add this line to run updates every Sunday at 2:00 AM:

```bash
# Update MakeMKV every Sunday at 2:00 AM
0 2 * * 0 /usr/local/bin/rip update >> /var/log/rip-update.log 2>&1
```

Or for a different day/time:
- `0 2 * * 1` - Monday at 2:00 AM
- `0 2 * * 2` - Tuesday at 2:00 AM
- `0 2 * * 3` - Wednesday at 2:00 AM
- etc.

**Check the cron log:**
```bash
tail -f /var/log/rip-update.log
```

---

## File Not Named Correctly After Rip

**Problem:** The file is named `title_t00.mkv` instead of `The Break-Up (2006).mkv`

**Cause:** FileBot couldn't find the movie in the TMDB database or couldn't rename the file.

**Solution:**

1. **Check the output** - Look at the "Target:" line in the rip output:
   ```
   Target: /Volumes/Users/rmasci/Videos/Romantic/TheBreakUp2006/The Break-Up (2006).mkv
   ```

2. **Navigate to the directory:**
   ```bash
   cd /Volumes/Users/rmasci/Videos/Romantic/TheBreakUp2006
   ```

3. **Manually rename the file** to match the target name:
   ```bash
   mv title_t00.mkv "The Break-Up (2006).mkv"
   ```

4. **Verify the rename:**
   ```bash
   ls -la
   ```

5. **Rescan in Plex/Jellyfin** - Go to your media server and trigger a library scan. It will now recognize the properly named file.

**Note:** The filename format is: `"Movie Name (Year).mkv"`

---

## Command Not Found: filebot

**Problem:** Error message says `filebot: command not found`

**Cause:** FileBot is not installed or not in your PATH

**Solution - macOS:**
```bash
# Install via Homebrew
brew install filebot

# Or install via snap (if you have snap installed)
snap install filebot --classic

# Or download from: https://www.filebot.net/
```

**Solution - Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install filebot

# Or install via snap
snap install filebot --classic

# Or download from: https://www.filebot.net/
```

**Verify installation:**
```bash
filebot -version
```

---

## Command Not Found: makemkvcon

**Problem:** Error message says `makemkvcon: command not found`

**Cause:** MakeMKV is not installed or not in your PATH

**Solution - macOS:**
```bash
# Install via Homebrew
brew install makemkv

# Or download from: https://www.makemkv.com/
```

**Solution - Linux:**

MakeMKV is not available in standard Linux repositories. You must download and build it manually:

```bash
# Install dependencies first
sudo apt-get install build-essential pkg-config libssl-dev libavcodec-extra

# Create a working directory
mkdir -p ~/makemkv && cd ~/makemkv

# Download the latest version (check https://www.makemkv.com/download/ for latest version)
# As of 2026, replace VERSION with the actual version number
wget https://www.makemkv.com/download/makemkv-oss-VERSION.tar.gz
wget https://www.makemkv.com/download/makemkv-bin-VERSION.tar.gz

# Extract and build MakeMKV OSS
tar xzf makemkv-oss-VERSION.tar.gz
cd makemkv-oss-VERSION
./configure
make
sudo make install
cd ..

# Extract and build MakeMKV bin
tar xzf makemkv-bin-VERSION.tar.gz
cd makemkv-bin-VERSION
./configure
make
sudo make install
cd ..

# Verify installation
makemkvcon -v full
```

**Simplified Alternative (if you have snap):**
```bash
# Install via snap (if available on your system)
sudo snap install makemkv
```

Or visit https://www.makemkv.com/download/ for precompiled binaries or other installation methods.

**Verify installation:**
```bash
makemkvcon -v full info disc:0
```

---

## Storage Path Does Not Exist

**Problem:** Error message: `storage path does not exist: /path/to/storage`

**Cause:** The storage path in `~/.rip.conf` doesn't exist

**Solution:**

1. **Edit your config file:**
   ```bash
   nano ~/.rip.conf
   ```

2. **Change the `storage_path` to an existing directory:**
   ```
   # Change from this:
   storage_path=/plex/storage
   
   # To this (for example):
   storage_path=~/Videos
   ```

3. **Make sure the directory exists:**
   ```bash
   mkdir -p ~/Videos
   ```

4. **Save and exit** (Ctrl+X, then Y, then Enter)

5. **Try the rip command again**

---

## Storage Path Is Not Writable

**Problem:** Error message: `storage path is not writable: permission denied`

**Cause:** The directory exists but you don't have write permissions

**Solution:**

1. **Check the permissions:**
   ```bash
   ls -ld /path/to/storage
   ```

2. **Fix permissions** (if you own the directory):
   ```bash
   chmod 755 /path/to/storage
   ```

3. **Or change the ownership** (if needed):
   ```bash
   sudo chown $USER:$USER /path/to/storage
   ```

4. **Try the rip command again**

---

## DVD Drive Not Found

**Problem:** Error message about device not found or MakeMKV fails to access disc

**Cause:** Wrong device path or DVD drive not recognized

**Solution:**

1. **Find your DVD drive:**
   ```bash
   # macOS
   ls -la /dev/rdisk*
   
   # Linux
   ls -la /dev/sr*
   ```

2. **Insert the DVD** and try again

3. **Use the correct device path:**
   ```bash
   # macOS example
   rip dvd -c Action -m "Movie Title" -d /dev/rdisk6
   
   # Linux example
   rip dvd -c Action -m "Movie Title" -d /dev/sr0
   ```

4. **Try a different drive number** if you have multiple drives:
   ```bash
   rip dvd -c Action -m "Movie Title" -d /dev/rdisk7
   ```

---

## MakeMKV Extraction Failed

**Problem:** Error message: `MakeMKV extraction failed. Please check your DVD and try again.`

**Cause:** Could be a dirty disc, copy protection issue, or MakeMKV configuration problem

**Solution:**

1. **Clean the DVD** - Use a soft cloth to gently clean the disc

2. **Check if the DVD plays** in a regular DVD player first

3. **Try again** with a fresh insertion

4. **Enable MakeMKV debug output** (advanced):
   ```bash
   # Run makemkvcon directly to see detailed errors
   makemkvcon -r info disc:0
   ```

5. **Check MakeMKV license** - Some DVDs require a valid license:
   ```bash
   filebot -version
   makemkvcon -v full
   ```

---

## MakeMKV Version Too Old or Expired

**Problem:** Error message: `This application version is too old. Please download the latest version at http://www.makemkv.com/ or enter a registration key to continue using the current version.`

Or: `Error during MakeMKV rip: error running makemkvcon info: exit status 253`

**Cause:** MakeMKV version has expired or is outdated. MakeMKV periodically expires old beta versions.

**Solution:**

1. **Update MakeMKV to the latest version:**
   ```bash
   # macOS
   brew upgrade makemkv
   
   # Linux - Build from source (see "Command Not Found: makemkvcon" section above)
   # Visit https://www.makemkv.com/download/ to get the latest version
   # Then follow the build instructions in the Linux section
   
   # Or use snap if available
   sudo snap refresh makemkv
   ```

2. **Verify the updated version:**
   ```bash
   makemkvcon -v full
   ```

3. **Or enter a registration key** (if you have one):
   ```bash
   # Edit MakeMKV configuration file
   nano ~/.config/MakeMKV/settings.conf
   
   # Add your registration key line:
   # app_key = "xxxx-xxxx-xxxx-xxxx"
   ```

4. **Try the rip command again** after updating

---

## FileBot Can't Find the Movie or Show

**Problem:** FileBot returns "No matches found" or the wrong movie/show is selected

**Cause:** Movie/show name doesn't match exactly in TMDB/TheTVDB database, or DVD title is unclear

**Solution:**

### For Movies:

1. **Search TMDB directly** to find the exact title:
   - Visit https://www.themoviedb.org/
   - Search for your movie
   - Note the exact title and year

2. **Use the exact title when ripping:**
   ```bash
   # Bad (might not match)
   rip dvd -c Action -m "Dark Knight"
   
   # Good (exact title from TMDB)
   rip dvd -c Action -m "The Dark Knight"
   
   # Even better (with year)
   rip dvd -c Action -m "The Dark Knight (2008)"
   ```

3. **Common issues to avoid:**
   - Missing "The" at the beginning: "Matrix" vs "The Matrix"
   - Punctuation differences: "Don't Look Up" vs "Dont Look Up"
   - Foreign titles: Use English title from TMDB
   - Subtitles: "Spider-Man: Into the Spider-Verse" - include the full title

4. **If still not found:**
   - Try just the year: `rip dvd -c Action -m "2008"`
   - Try without year: `rip dvd -c Action -m "The Dark Knight"`
   - Try a different spelling or alternate title
   - Check if it's in TMDB at all (some old DVDs may not be)

### For TV Shows:

1. **Search TheTVDB directly** to find the exact title:
   - Visit https://www.thetvdb.com/
   - Search for your show
   - Note the exact title

2. **Use the exact title when ripping:**
   ```bash
   # Bad (might not match)
   rip tv "Office" 1-1
   
   # Good (exact title from TheTVDB)
   rip tv "The Office" 1-1
   
   # With year (sometimes helps)
   rip tv "The Office (2005)" 1-1
   ```

3. **Common issues to avoid:**
   - Missing "The" at the beginning
   - Regional variations: "Skins (US)" vs "Skins (UK)"
   - Checking the right version in TheTVDB

### Manual Workaround if FileBot Fails:

If FileBot can't find the movie/show after multiple attempts:

1. **Rip the DVD anyway** (FileBot naming failure won't stop the rip):
   ```bash
   rip dvd -c Action -m "Some Movie" -d /dev/rdisk6
   ```

2. **File will be named** something like `title_t00.mkv`

3. **Manually research the correct name** on TMDB/TheTVDB

4. **Rename the file manually:**
   ```bash
   cd ~/Videos/Action/SomeMovie
   mv title_t00.mkv "The Correct Movie Title (2020).mkv"
   ```

5. **Rescan library** in Plex/Jellyfin - it will recognize the properly named file

---

## DVD Title Cannot Be Auto-Discovered

**Problem:** Rip command can't automatically detect the movie/show name from the DVD

**Cause:** DVD metadata doesn't contain readable title information

**Solution:**

### Always Provide the Movie/Show Name Explicitly:

Instead of relying on auto-discovery, use the `-m` flag for movies or provide the name for TV:

```bash
# DVD command - always use -m flag
rip dvd -c Action -m "The Movie Title" -d /dev/rdisk6

# TV command - always provide show name
rip tv "The Show Name" 1-1 -d /dev/rdisk6
```

### How to Find the Correct Name:

1. **Look at the DVD case/label** - This is usually accurate

2. **Search online:**
   - IMDb: https://www.imdb.com/
   - TMDB: https://www.themoviedb.org/
   - TheTVDB: https://www.thetvdb.com/

3. **Check the DVD menu** - Sometimes displayed when you play it

4. **Use a DVD player app** - May show title information

### Best Practices:

1. **Always use `-m` flag for movies** to avoid relying on auto-discovery:
   ```bash
   rip dvd -c "Category" -m "Exact Movie Title (Year)" -d /dev/rdisk6
   ```

2. **Always provide show name for TV** - Don't skip it:
   ```bash
   rip tv "Exact Show Name (Year)" 1-1 -d /dev/rdisk6
   ```

3. **Research before ripping** - Take 30 seconds to verify the exact title on TMDB/TheTVDB

4. **Use the exact spelling and punctuation** from TMDB/TheTVDB:
   - "Spider-Man" not "Spiderman"
   - "Doctor Who" not "Doctor Who (2005)" - check which version you have

---

## Testing FileBot Directly

If you're unsure about the movie/show name, test FileBot directly:

```bash
# Test if FileBot can find a movie
filebot -list --db TheMovieDB --q "The Matrix" --format '{n} ({y})'
# Output: The Matrix (1999)

# Test if FileBot can find a show
filebot -list --db TheTVDB --q "The Office" --format '{n} ({y})'
# Output: The Office (2005)
```

If FileBot returns matches, use that exact name in your rip command.

---

## Movie Title Variations

If your DVD has a subtitle or regional variation, try these:

```bash
# Main title only
rip dvd -c Action -m "Spider-Man" -d /dev/rdisk6

# With subtitle
rip dvd -c Action -m "Spider-Man: Into the Spider-Verse" -d /dev/rdisk6

# With year (if ambiguous)
rip dvd -c Action -m "Spider-Man (2002)" -d /dev/rdisk6

# Try alternative titles listed on TMDB
rip dvd -c Action -m "The Dark Knight" -d /dev/rdisk6
# If that fails, try the alternate:
rip dvd -c Action -m "Batman: The Dark Knight" -d /dev/rdisk6
```

Check TMDB for all alternate titles if the main title doesn't work.

---



## Configuration File Questions

**Problem:** Need to change storage location or other settings

**Solution:**

1. **Edit the config file:**
   ```bash
   nano ~/.rip.conf
   ```

2. **Available settings:**
   ```
   # Path where media will be stored
   storage_path=/plex/storage
   
   # You can use home directory shortcut
   storage_path=~/Videos/DVDRips
   
   # Or absolute paths
   storage_path=/mnt/media/rips
   ```

3. **Save and restart** the rip command (changes take effect immediately)

4. **View your current config:**
   ```bash
   cat ~/.rip.conf
   ```

---

## Files Not Appearing in Plex/Jellyfin After Rip

**Problem:** Rip completed successfully but files don't show up in media server

**Cause:** Library hasn't been scanned yet or files aren't in the monitored directory

**Solution:**

1. **Verify files exist** in the storage location:
   ```bash
   ls -lah ~/Videos/  # or your storage directory
   ```

2. **Trigger a library scan in Plex:**
   - Open Plex web interface
   - Click Settings (gear icon)
   - Go to Libraries
   - Click the three dots next to your library
   - Select "Refresh Library"
   - Or press the Refresh button at the top

3. **Trigger a library scan in Jellyfin:**
   - Open Jellyfin dashboard
   - Click Administration
   - Go to Libraries
   - Click the refresh icon next to your library
   - Or use Dashboard → Libraries → Scan Library

4. **Wait for scan to complete** - It can take a few minutes for large libraries

5. **Restart your media server** if files still don't appear:
   ```bash
   # Jellyfin (if running as service)
   sudo systemctl restart jellyfin
   
   # Plex (restart through app)
   ```

---

## TV Episodes Not Named Correctly

**Problem:** TV episodes named `title_t00.mkv` instead of proper episode names

**Cause:** Same as movies - FileBot couldn't find the show or rename files

**Solution:**

1. **Check the output** for the expected filename format

2. **Manually rename episodes** following this pattern:
   ```bash
   # Standard format: Show Name - S01E01 - Episode Title.mkv
   mv title_t00.mkv "The Office - S01E01 - Pilot.mkv"
   mv title_t01.mkv "The Office - S01E02 - Diversity Day.mkv"
   ```

3. **Use FileBot manually** to rename all episodes at once:
   ```bash
   filebot -rename ~/Videos/TVShow/ -r --db TheTVDB --format '{n} - S{s}E{e} - {t}' --action move
   ```

4. **Plex/Jellyfin will auto-recognize** properly named episodes on next library scan

---

## Rip Command Hangs or Takes Too Long

**Problem:** The rip command seems stuck or takes much longer than expected

**Cause:** Could be slow disk I/O, large files, or network issues

**Solution:**

1. **Check if it's still running:**
   ```bash
   # In another terminal, list processes
   ps aux | grep makemkv
   ps aux | grep filebot
   ```

2. **Give it more time** - Ripping can take 20-60+ minutes depending on:
   - DVD quality
   - File size
   - Disk speed
   - System resources

3. **Monitor disk usage** while ripping:
   ```bash
   df -h  # See available disk space
   iostat 1  # Monitor disk I/O
   ```

4. **Check system resources:**
   ```bash
   # macOS
   top -l 1 | grep "CPU\|Mem"
   
   # Linux
   free -h
   ```

5. **Stop the process** if needed:
   ```bash
   # Press Ctrl+C in the terminal
   # Or kill the process
   pkill makemkv
   pkill filebot
   ```

---

## Not Enough Disk Space

**Problem:** Error about disk space or rip fails partway through

**Cause:** Storage device doesn't have enough free space

**Solution:**

1. **Check available space:**
   ```bash
   df -h ~/Videos  # or your storage directory
   ```

2. **Estimate file size needed:**
   - DVD: Usually 3-8 GB per movie
   - TV Season: 500 MB - 2 GB per episode
   - Blu-ray: 15-50 GB

3. **Free up space:**
   ```bash
   # Remove old files
   rm -rf ~/Videos/OldMovie/
   
   # Clear temporary files
   rm -rf /tmp/*
   
   # Check what's using space
   du -sh ~/*  # Shows directory sizes
   ```

4. **Add more storage:**
   - Add a new external drive
   - Connect to a network storage device
   - Update `~/.rip.conf` to point to new location

5. **Try the rip again** once you have enough space

---

## Permission Denied When Creating Directories

**Problem:** Error like `permission denied` when trying to create folders

**Cause:** Don't have write access to the storage directory

**Solution:**

1. **Check current permissions:**
   ```bash
   ls -ld ~/Videos
   ```

2. **Fix permissions** (if you own the directory):
   ```bash
   chmod u+rwx ~/Videos
   ```

3. **Fix ownership** (if needed):
   ```bash
   # Change to your user
   sudo chown -R $USER:$USER ~/Videos
   ```

4. **Create test file** to verify:
   ```bash
   touch ~/Videos/test.txt
   rm ~/Videos/test.txt
   ```

5. **Try rip again**

---

## DVD Disc Has Copy Protection

**Problem:** MakeMKV can't extract from the DVD

**Cause:** DVD has copy protection that MakeMKV can't handle

**Solution:**

1. **Check if it's a known issue:**
   - Update MakeMKV to the latest version
   - Check MakeMKV forum for this specific disc

2. **Try MakeMKV manually:**
   ```bash
   makemkvcon -r info disc:0
   ```

3. **Use different software** (if needed):
   - VLC with libdvdcss
   - HandBrake with copy protection support
   - Commercial DVD ripping software

4. **Check disc integrity:**
   - Try playing it in a DVD player first
   - If it won't play, it might be damaged
   - Clean the disc with a soft cloth

---

## Category Folder Not Created

**Problem:** Storage path created but category subfolder is missing

**Cause:** Directory creation failed or rip stopped early

**Solution:**

1. **Create category manually:**
   ```bash
   mkdir -p ~/Videos/Action
   mkdir -p ~/Videos/Comedy
   mkdir -p ~/Videos/Drama
   ```

2. **Verify permissions:**
   ```bash
   ls -la ~/Videos/
   chmod 755 ~/Videos/Action
   ```

3. **Try rip again**

---

## Want to Organize Media Differently

**Problem:** Want a different directory structure than Category/MovieName

**Cause:** Current setup uses category-based organization

**Solution:**

Unfortunately, the current code uses a fixed directory structure:
- Movies: `~/Videos/Category/MovieName/`
- TV: `~/Videos/Genre/ShowName/Season XX/`

**Workarounds:**

1. **Create symbolic links** to organize by year:
   ```bash
   ln -s ~/Videos/Action/MovieName ~/Videos/2006/MovieName
   ```

2. **Move files after ripping:**
   ```bash
   mv ~/Videos/Action/MovieName ~/Videos/MovieName-Action
   ```

3. **Use Plex/Jellyfin collections** to group by genre/rating

4. **Request custom organization** in future versions

---

## macOS vs Linux Differences

**Problem:** Command works on one OS but not the other

**Cause:** Different default paths and tools between operating systems

**Common differences:**

| Feature | macOS | Linux |
|---------|-------|-------|
| DVD Device | `/dev/rdisk*` | `/dev/sr*` |
| Config Location | `~/.rip.conf` | `~/.rip.conf` |
| Package Manager | Homebrew | apt/yum/pacman |
| Home Directory | `/Users/username` | `/home/username` |
| Temp Files | `/tmp` | `/tmp` |

**Solution:**

1. **Use platform-specific commands:**
   ```bash
   # macOS
   brew install filebot makemkv
   
   # Linux
   sudo apt install filebot makemkv-bin
   ```

2. **Use correct device paths:**
   ```bash
   # macOS
   rip dvd -d /dev/rdisk6
   
   # Linux
   rip dvd -d /dev/sr0
   ```

---

## Still Not Working?

**Problem:** Issue not covered in troubleshooting guide

**Cause:** Unique or unusual configuration

**Solution:**

1. **Check the complete output** - Read all the messages from the rip command carefully

2. **Verify all dependencies are installed:**
   ```bash
   which filebot
   which makemkvcon
   which ffprobe
   ```

3. **Check file permissions:**
   ```bash
   ls -la ~/.rip.conf
   ls -la ~/Videos/  # or your storage directory
   ```

4. **Try running commands manually:**
   ```bash
   # Test FileBot
   filebot -version
   
   # Test MakeMKV
   makemkvcon -v full info disc:0
   
   # Test FFprobe
   ffprobe -version
   ```

5. **Check available disk space:**
   ```bash
   df -h
   ```

---

## Getting Help

When reporting issues, include:

1. The complete command you ran
2. The full error message from the output
3. Your operating system and version
4. The output of `rip --version`
5. The result of `which filebot && which makemkvcon`

---

## Quick Reference

**Common Commands:**

```bash
# Check version
rip --version

# Get help
rip --help
rip dvd --help
rip tv --help

# Edit config
nano ~/.rip.conf

# View config
cat ~/.rip.conf

# Check dependencies
filebot -version
makemkvcon -v full
ffprobe -version

# Manually rename a file
mv old_name.mkv "New Name (2006).mkv"

# Trigger library scan in Plex
# Go to Settings > Libraries > Scan Library
# Or use Plex API

# Find your DVD drive
ls -la /dev/rdisk*     # macOS
ls -la /dev/sr*        # Linux

# Check disk space
df -h
```

---

## Common Device Paths

**macOS:**
```
/dev/rdisk0  - First DVD drive
/dev/rdisk1  - Second DVD drive
/dev/rdisk6  - etc.
```

**Linux:**
```
/dev/sr0   - First DVD drive
/dev/sr1   - Second DVD drive
/dev/dvd   - Symbolic link to default DVD drive
```

---

## Tips for Success

1. **Always insert the DVD first** before running the rip command
2. **Use exact movie/show names** from TMDB/TheTVDB for better metadata matching
3. **Keep your FileBot license current** for best renaming results
4. **Monitor the output** for any warnings or errors
5. **Check your storage space** before ripping large files
6. **Clean dirty DVDs** to avoid extraction errors
7. **Verify renamed files** in your media server after scanning
8. **Wait for library scan** - It can take several minutes
9. **Restart media server** if files still don't appear after scan
10. **Keep a backup** of your DVDs for archival purposes

---

## Still Have Questions?

Refer to the documentation:
- FileBot: https://www.filebot.net/
- MakeMKV: https://www.makemkv.com/
- FFmpeg: https://ffmpeg.org/
- Plex: https://www.plex.tv/
- Jellyfin: https://jellyfin.org/
