# Systemd Configuration for Clipper Scraper

This directory contains systemd service and timer units for running the Clipper clip scraper automatically.

## Installation

### 1. Copy Service Files

```bash
# Copy service and timer files to systemd directory
sudo cp clpr-scraper.service /etc/systemd/system/
sudo cp clpr-scraper.timer /etc/systemd/system/

# Set correct permissions
sudo chmod 644 /etc/systemd/system/clpr-scraper.service
sudo chmod 644 /etc/systemd/system/clpr-scraper.timer
```

### 2. Update Paths

Edit the service file to match your installation:

```bash
sudo nano /etc/systemd/system/clpr-scraper.service
```

Update these values:
- `User=clpr` - Change to your application user
- `Group=clpr` - Change to your application group
- `WorkingDirectory=/opt/clpr/backend` - Change to your installation path
- `ExecStart=/opt/clpr/backend/bin/scrape_clips` - Change to your binary path
- `EnvironmentFile=-/opt/clpr/backend/.env` - Change to your .env path

### 3. Reload Systemd

```bash
sudo systemctl daemon-reload
```

### 4. Enable and Start Timer

```bash
# Enable timer to start on boot
sudo systemctl enable clpr-scraper.timer

# Start timer now
sudo systemctl start clpr-scraper.timer
```

## Usage

### Check Timer Status

```bash
# Check if timer is active
sudo systemctl status clpr-scraper.timer

# List all timers
sudo systemctl list-timers

# View detailed timer info
sudo systemctl show clpr-scraper.timer
```

### View Service Status

```bash
# Check last run status
sudo systemctl status clpr-scraper.service
```

### View Logs

```bash
# View all logs
sudo journalctl -u clpr-scraper.service

# Follow logs in real-time
sudo journalctl -u clpr-scraper.service -f

# View logs from last run
sudo journalctl -u clpr-scraper.service -n 100

# View logs from specific date
sudo journalctl -u clpr-scraper.service --since "2024-01-01"
```

### Manual Execution

```bash
# Run service manually (bypasses timer)
sudo systemctl start clpr-scraper.service

# Watch logs while running
sudo journalctl -u clpr-scraper.service -f
```

### Stop/Disable Timer

```bash
# Stop timer (prevents future runs)
sudo systemctl stop clpr-scraper.timer

# Disable timer (won't start on boot)
sudo systemctl disable clpr-scraper.timer

# Stop running service
sudo systemctl stop clpr-scraper.service
```

### Restart After Changes

```bash
# After editing service or timer files
sudo systemctl daemon-reload
sudo systemctl restart clpr-scraper.timer
```

## Configuration

### Changing Schedule

Edit the timer file:

```bash
sudo nano /etc/systemd/system/clpr-scraper.timer
```

Common schedules:

```ini
# Daily at 2 AM
OnCalendar=*-*-* 02:00:00

# Every 6 hours
OnCalendar=*-*-* 00,06,12,18:00:00

# Weekly on Sunday at 2 AM
OnCalendar=Sun *-*-* 02:00:00

# Twice daily (2 AM and 2 PM)
OnCalendar=*-*-* 02,14:00:00
```

After changes:

```bash
sudo systemctl daemon-reload
sudo systemctl restart clpr-scraper.timer
```

### Adding Scraper Options

Edit the service file:

```bash
sudo nano /etc/systemd/system/clpr-scraper.service
```

Modify `ExecStart`:

```ini
# With custom options
ExecStart=/opt/clpr/backend/bin/scrape_clips --batch-size 100 --min-views 200

# With dry-run
ExecStart=/opt/clpr/backend/bin/scrape_clips --dry-run
```

## Monitoring

### Check Next Run Time

```bash
sudo systemctl list-timers clpr-scraper.timer
```

### Enable Email Notifications

Install and configure `mailutils`:

```bash
sudo apt-get install mailutils
```

Edit service file to add email on failure:

```ini
[Service]
# ... existing config ...
OnFailure=status-email@%n.service
```

### Export Logs

```bash
# Export logs to file
sudo journalctl -u clpr-scraper.service > scraper-logs.txt

# Export logs as JSON
sudo journalctl -u clpr-scraper.service -o json > scraper-logs.json
```

## Troubleshooting

### Timer Not Running

```bash
# Check timer status
sudo systemctl status clpr-scraper.timer

# Check for errors
sudo journalctl -u clpr-scraper.timer -xe
```

### Service Fails to Start

```bash
# Check service status
sudo systemctl status clpr-scraper.service

# View detailed logs
sudo journalctl -u clpr-scraper.service -xe

# Test manually
cd /opt/clpr/backend
./bin/scrape_clips --dry-run
```

### Permission Issues

```bash
# Check file permissions
ls -la /opt/clpr/backend/bin/scrape_clips

# Make executable
sudo chmod +x /opt/clpr/backend/bin/scrape_clips

# Check user/group
id clpr
```

### Environment Variables Not Loading

Ensure `.env` file exists and is readable:

```bash
ls -la /opt/clpr/backend/.env
sudo chmod 640 /opt/clpr/backend/.env
sudo chown clpr:clpr /opt/clpr/backend/.env
```

## Security Considerations

The service includes several security hardening options:

- `NoNewPrivileges=true` - Prevents privilege escalation
- `PrivateTmp=true` - Private /tmp directory
- `ProtectSystem=strict` - Read-only system directories
- `ProtectHome=true` - Inaccessible home directories
- Resource limits (Memory, CPU)

Adjust based on your security requirements.

## Uninstallation

```bash
# Stop and disable timer
sudo systemctl stop clpr-scraper.timer
sudo systemctl disable clpr-scraper.timer

# Remove service files
sudo rm /etc/systemd/system/clpr-scraper.service
sudo rm /etc/systemd/system/clpr-scraper.timer

# Reload systemd
sudo systemctl daemon-reload
```
