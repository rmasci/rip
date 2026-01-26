# DVD Ripper Web Interface - Deployment & Usage Guide

A modern, user-friendly web interface for managing DVD ripping operations. Built with Jellyfin-inspired design and served alongside the `rip` CLI tool.

## Quick Start

### Prerequisites
- Go 1.25.0 or later (for building from source)
- `rip` binary in your PATH or specified with `--rip` flag
- Available DVD drive
- Storage directory for ripped media

### Build & Run

```bash
# Build the project
make all

# Run the web interface
./rip web --port 8080 --storage ~/Videos
```

Then open your browser to: **http://localhost:8080**

## Installation

### From Source

```bash
git clone https://github.com/rmasci/dvdrip.git
cd dvdrip
make release
```

Binaries will be in the `release/` directory:
- `rip-mac-amd64` / `rip-mac-arm64` (macOS)
- `rip-linux-amd64` / `rip-linux-arm64` (Linux)
- `rip-windows-amd64.exe` / `rip-windows-arm64.exe` (Windows)

### System-Wide Installation

```bash
# Copy binary to system path
sudo cp release/rip-<os>-<arch> /usr/local/bin/rip
sudo chmod +x /usr/local/bin/rip

# Verify installation
rip --version
```

## Configuration

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `--port`, `-p` | 8080 | Port to run the web server on |
| `--storage` | ~/Videos | Storage path for ripped media and categories |
| `--rip` | rip | Path to the rip CLI command |

### Usage Examples

```bash
# Default configuration
rip web

# Custom port and storage
rip web --port 9000 --storage /plex/storage

# Specify custom rip command path
rip web --port 8080 --storage ~/Videos --rip /opt/dvdrip/rip

# All options combined
rip web -p 8080 --storage /media/dvdrip --rip /usr/local/bin/rip
```

## Web Interface Features

### Start Rip Tab
The main interface for initiating DVD rip jobs:

1. **Select DVD Device** - Choose from auto-detected DVD drives
2. **Enter Movie Name** - Name of the movie to rip (auto-looked up in TMDB)
3. **Select Category** - Organize rips into categories (Drama, Comedy, etc.)
4. **Click "Start Rip"** - Background job begins immediately

Features:
- Automatic movie metadata lookup
- Device auto-detection (Linux `/dev/sr*`, macOS `/dev/rdisk*`)
- Category creation on-the-fly
- Non-blocking background processing

### Settings Tab
Configure application behavior and transcoding options:

**Application Settings** (read-only):
- Storage Path: Where ripped files are organized
- Rip Command: Path to the rip executable

**makemkvcon Settings**:
- Minimum Title Length (seconds) - Skip titles shorter than this
- Preferred Format - Choose between MKV and M2TS

**Transcoding Settings**:
- Enable/Disable transcoding
- Video Codec - H.264, H.265 (HEVC), or VP9
- Bitrate - Video bitrate in kbps (1000-50000)

All settings are saved in browser localStorage.

## Directory Structure

The web interface organizes ripped content by category:

```
/plex/storage/
├── Drama/
│   ├── The Break-Up (2006)/
│   │   └── The Break-Up (2006).mkv
│   ├── Inception (2010)/
│   │   └── Inception (2010).mkv
│   └── ...
├── Comedy/
│   ├── The Hangover (2009)/
│   │   └── The Hangover (2009).mkv
│   └── ...
├── Action/
│   └── ...
└── Horror/
    └── ...
```

**Path Format**: `<storage>/<category>/<MovieName YYYY>/<MovieName YYYY>.mkv`

## API Reference

The web daemon exposes RESTful API endpoints for integration:

### Device Management

**GET `/api/devices`**
List available DVD drives
```json
{
  "devices": ["/dev/sr0", "/dev/rdisk6"]
}
```

**GET `/api/status`**
Get daemon status and configuration
```json
{
  "status": "running",
  "storagePath": "/plex/storage",
  "ripCommand": "rip",
  "availableDevices": ["/dev/sr0"]
}
```

### Category Management

**GET `/api/categories`**
List all categories
```json
{
  "categories": ["Drama", "Comedy", "Action"]
}
```

**POST `/api/categories`**
Create a new category
```json
{
  "name": "Horror"
}
```

**PUT `/api/categories`**
Rename a category
```json
{
  "oldName": "Old Name",
  "newName": "New Name"
}
```

**DELETE `/api/categories`**
Delete a category (must be empty)
```json
{
  "name": "Category Name"
}
```

### Ripping

**POST `/api/rip`**
Start a rip job
```json
{
  "device": "/dev/sr0",
  "category": "Drama",
  "movie": "The Shawshank Redemption"
}
```

Response: `202 Accepted`
```json
{
  "status": "rip started",
  "movie": "The Shawshank Redemption"
}
```

## Running as a System Service

### macOS (launchd)

Create `~/Library/LaunchAgents/com.rmasci.dvdrip-web.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.rmasci.dvdrip-web</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/rip</string>
        <string>web</string>
        <string>--port</string>
        <string>8080</string>
        <string>--storage</string>
        <string>/plex/storage</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>/var/log/dvdrip-web.err.log</string>
    <key>StandardOutPath</key>
    <string>/var/log/dvdrip-web.out.log</string>
</dict>
</plist>
```

**Load the service:**
```bash
launchctl load ~/Library/LaunchAgents/com.rmasci.dvdrip-web.plist
```

**Unload the service:**
```bash
launchctl unload ~/Library/LaunchAgents/com.rmasci.dvdrip-web.plist
```

**View logs:**
```bash
tail -f /var/log/dvdrip-web.out.log
tail -f /var/log/dvdrip-web.err.log
```

### Linux (systemd)

Create `/etc/systemd/system/dvdrip-web.service`:

```ini
[Unit]
Description=DVD Ripper Web Interface
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=nobody
Group=nogroup
ExecStart=/usr/local/bin/rip web --port 8080 --storage /plex/storage
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=dvdrip-web

[Install]
WantedBy=multi-user.target
```

**Enable and start:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable dvdrip-web
sudo systemctl start dvdrip-web
```

**Manage the service:**
```bash
# Check status
sudo systemctl status dvdrip-web

# View logs
sudo journalctl -u dvdrip-web -f

# Restart
sudo systemctl restart dvdrip-web

# Stop
sudo systemctl stop dvdrip-web
```

## Reverse Proxy Setup

### nginx

Configure nginx to proxy the web interface:

```nginx
server {
    listen 80;
    server_name media.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_redirect off;
    }
}
```

### Apache

```apache
<VirtualHost *:80>
    ServerName media.example.com
    
    ProxyPreserveHost On
    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/
    
    RequestHeader set X-Real-IP %{REMOTE_ADDR}s
    RequestHeader set X-Forwarded-For %{HTTP_X_FORWARDED_FOR}e
    RequestHeader set X-Forwarded-Proto %{REQUEST_SCHEME}s
</VirtualHost>
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using the port
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use a different port
rip web --port 9000
```

### No DVD Devices Found

**Check devices:**
```bash
# macOS
ls -la /dev/rdisk*

# Linux
ls -la /dev/sr*
```

**Fix permissions (Linux):**
```bash
# Add user to optical group
sudo usermod -a -G optical $(whoami)

# Or change device permissions
sudo chmod 666 /dev/sr0
```

**Check makemkvcon:**
```bash
makemkvcon info disc:0
```

### Storage Path Not Found

```bash
# Create storage directory
mkdir -p /plex/storage

# Set permissions
chmod 755 /plex/storage

# Run with correct path
rip web --storage /plex/storage
```

### Rip Job Fails

**Verify rip binary is accessible:**
```bash
which rip
rip --version
```

**Check permissions:**
```bash
# DVD drive
ls -la /dev/sr0  # Linux
ls -la /dev/rdisk*  # macOS

# Storage directory
ls -la /plex/storage
```

**Check logs:**
```bash
# macOS
tail -f /var/log/dvdrip-web.out.log

# Linux
sudo journalctl -u dvdrip-web -f
```

### Settings Not Persisting

- Ensure browser localStorage is enabled
- Check browser privacy settings
- Try a different browser
- Clear browser cache: `Ctrl+Shift+Del` (or `Cmd+Shift+Del` on macOS)

## Browser Compatibility

| Browser | Supported | Notes |
|---------|-----------|-------|
| Chrome 90+ | ✅ | Recommended |
| Firefox 88+ | ✅ | Full support |
| Safari 14+ | ✅ | Full support |
| Edge 90+ | ✅ | Full support |

## Security Considerations

### Local Network Only (Recommended)
By default, the web interface is only accessible from localhost. For local network access:

```bash
# This allows access from any machine on the network
rip web --port 8080 --storage /plex/storage
# Then access via: http://<ip-address>:8080
```

### Behind a Firewall
When exposing to the internet, use a reverse proxy with authentication:

```nginx
server {
    listen 443 ssl http2;
    server_name media.example.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    auth_basic "DVD Ripper";
    auth_basic_user_file /etc/nginx/.htpasswd;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Performance

### Recommended Specifications
- **CPU**: Modern multi-core processor (Intel/AMD or Apple Silicon)
- **RAM**: 4GB minimum (8GB+ for concurrent operations)
- **Storage**: SSD recommended for faster transcoding
- **Network**: Gigabit for optimal streaming

### Optimization Tips
- Run the daemon on the same machine as storage
- Use wired Ethernet for reliability
- Enable hardware acceleration if available
- Monitor system resources during transcoding

## Getting Help

### Debug Mode
Add verbose logging:

```bash
# Run with verbose output
rip web --port 8080 --storage ~/Videos 2>&1 | tee dvdrip-web.log
```

### Common Issues & Solutions

1. **Device not detected** - Check DVD drive connection and permissions
2. **Slow transcoding** - Check CPU usage and available RAM
3. **Storage full** - Ensure adequate space before starting rips
4. **Network issues** - Test connectivity with `ping` and `curl`

## Environment Variables

The web interface respects these environment variables:

```bash
# Set storage path via environment
export DVDRIP_STORAGE=/plex/storage
export DVDRIP_PORT=8080
export DVDRIP_RIP_CMD=/usr/local/bin/rip

rip web
```

## Theme & Design

The interface uses Jellyfin-inspired colors for a modern, media-center aesthetic:

- **Primary Color**: #00a4dc (Jellyfin Blue)
- **Accent Color**: #a335ee (Jellyfin Purple)
- **Background**: #1c1c1c (Dark)
- **Text**: #e0e0e0 (Light)

Icons and assets are included in `web/assets/` directory.

## License

DVD Ripper is licensed under the terms specified in the LICENSE file. See the main repository for details.

## Support & Feedback

For issues, feature requests, or questions:
- Open an issue on [GitHub](https://github.com/rmasci/dvdrip)
- Check [TROUBLESHOOTING.md](../TROUBLESHOOTING.md) for common problems
- Review [WEB_INTERFACE.md](../WEB_INTERFACE.md) for additional details

---

**Happy Ripping! 🎬**
