#!/bin/bash
set -euo pipefail

# Caddy Setup Helper for VPS
# This script helps configure Caddy to serve clpr.tv
# Can be used for initial setup or updating configuration

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log() { echo -e "${BLUE}[caddy-setup]${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}!${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }

# Check if Caddy is running
CADDY_CONTAINER=""
if docker ps --format '{{.Names}}' | grep -q caddy; then
    CADDY_CONTAINER=$(docker ps --format '{{.Names}}' | grep caddy | head -1)
    success "Found running Caddy container: $CADDY_CONTAINER"
else
    warn "No Caddy container found running"
fi

# Check if web network exists
if docker network inspect web >/dev/null 2>&1; then
    success "Network 'web' exists"
else
    warn "Network 'web' does not exist"
    log "Creating network 'web'..."
    docker network create web
    success "Network 'web' created"
fi

# Function to start new Caddy container
start_caddy_container() {
    log "Starting new Caddy container..."
    
    # Ensure directories exist
    mkdir -p ~/projects/caddy/data ~/projects/caddy/config ~/projects/caddy/logs
    
    # Copy Caddyfile
    if [ ! -f ~/projects/caddy/Caddyfile ]; then
        log "Copying Caddyfile.vps to ~/projects/caddy/Caddyfile..."
        cp "$PROJECT_ROOT/Caddyfile.vps" ~/projects/caddy/Caddyfile
        success "Caddyfile copied"
    else
        warn "Caddyfile already exists at ~/projects/caddy/Caddyfile"
        warn "Update it manually or remove it to use the default"
    fi
    
    # Start Caddy
    docker run -d \
        --name caddy \
        --network web \
        -p 80:80 \
        -p 443:443 \
        -p 2019:2019 \
        -v ~/projects/caddy/Caddyfile:/etc/caddy/Caddyfile:ro \
        -v ~/projects/caddy/data:/data \
        -v ~/projects/caddy/config:/config \
        -v ~/projects/caddy/logs:/var/log/caddy \
        --restart unless-stopped \
        caddy:2-alpine
    
    success "Caddy container started"
    CADDY_CONTAINER="caddy"
}

# Function to reload Caddy configuration
reload_caddy() {
    if [ -z "$CADDY_CONTAINER" ]; then
        error "No Caddy container running"
        return 1
    fi
    
    log "Reloading Caddy configuration..."
    if docker exec "$CADDY_CONTAINER" caddy reload --config /etc/caddy/Caddyfile; then
        success "Caddy configuration reloaded"
    else
        error "Failed to reload Caddy configuration"
        log "Trying to restart Caddy instead..."
        docker restart "$CADDY_CONTAINER"
        success "Caddy restarted"
    fi
}

# Function to update Caddyfile
update_caddyfile() {
    log "Updating Caddyfile..."
    
    # Find where Caddy's Caddyfile is
    if [ -f ~/projects/caddy/Caddyfile ]; then
        CADDY_CONFIG_PATH=~/projects/caddy/Caddyfile
    elif [ -n "$CADDY_CONTAINER" ]; then
        # Try to find it in the container
        CADDY_CONFIG_PATH=$(docker inspect "$CADDY_CONTAINER" | grep -A1 "Caddyfile" | grep "Source" | cut -d'"' -f4 | head -1)
    fi
    
    if [ -n "$CADDY_CONFIG_PATH" ] && [ -f "$CADDY_CONFIG_PATH" ]; then
        log "Found Caddyfile at: $CADDY_CONFIG_PATH"
        log "Creating backup..."
        cp "$CADDY_CONFIG_PATH" "${CADDY_CONFIG_PATH}.backup.$(date +%Y%m%d-%H%M%S)"
        
        log "Updating Caddyfile..."
        cp "$PROJECT_ROOT/Caddyfile.vps" "$CADDY_CONFIG_PATH"
        success "Caddyfile updated"
        
        reload_caddy
    else
        error "Could not find Caddyfile location"
    fi
}

# Function to verify configuration
verify_config() {
    log "Verifying Caddy configuration..."
    
    if [ -z "$CADDY_CONTAINER" ]; then
        warn "Caddy not running, cannot verify"
        return 1
    fi
    
    # Check if Caddy can see the backend
    log "Checking if Caddy can reach clpr-backend..."
    if docker exec "$CADDY_CONTAINER" wget -qO- --timeout=5 http://clpr-backend:8080/api/v1/health 2>/dev/null; then
        success "Caddy can reach clpr-backend"
    else
        warn "Caddy cannot reach clpr-backend"
        log "Ensure clpr-backend is running and on the 'web' network"
    fi
    
    # Check if Caddy can see the frontend
    log "Checking if Caddy can reach clpr-frontend..."
    if docker exec "$CADDY_CONTAINER" wget -qO- --timeout=5 http://clpr-frontend:80/ 2>/dev/null >/dev/null; then
        success "Caddy can reach clpr-frontend"
    else
        warn "Caddy cannot reach clpr-frontend"
        log "Ensure clpr-frontend is running and on the 'web' network"
    fi
    
    # Check TLS certificate status
    log "Checking TLS certificate status..."
    docker exec "$CADDY_CONTAINER" caddy list-certificates 2>/dev/null || warn "Could not list certificates"
}

# Main menu
show_menu() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║   Caddy Setup Helper for clpr.tv              ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Choose an action:"
    echo "  1) Start new Caddy container"
    echo "  2) Update Caddyfile configuration"
    echo "  3) Reload Caddy configuration"
    echo "  4) Verify Caddy configuration"
    echo "  5) View Caddy logs"
    echo "  6) Stop Caddy container"
    echo "  7) Show Caddy status"
    echo "  0) Exit"
    echo ""
    read -p "Enter choice [0-7]: " choice
    
    case $choice in
        1) start_caddy_container ;;
        2) update_caddyfile ;;
        3) reload_caddy ;;
        4) verify_config ;;
        5)
            if [ -n "$CADDY_CONTAINER" ]; then
                docker logs -f "$CADDY_CONTAINER"
            else
                error "No Caddy container running"
            fi
            ;;
        6)
            if [ -n "$CADDY_CONTAINER" ]; then
                docker stop "$CADDY_CONTAINER"
                success "Caddy stopped"
            else
                error "No Caddy container running"
            fi
            ;;
        7)
            log "Caddy status:"
            docker ps --filter "name=caddy" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
            echo ""
            if [ -n "$CADDY_CONTAINER" ]; then
                log "Networks:"
                docker inspect "$CADDY_CONTAINER" | grep -A 10 "Networks"
            fi
            ;;
        0) exit 0 ;;
        *) error "Invalid choice" ;;
    esac
}

# If arguments provided, run non-interactively
if [ $# -gt 0 ]; then
    case $1 in
        start) start_caddy_container ;;
        update) update_caddyfile ;;
        reload) reload_caddy ;;
        verify) verify_config ;;
        logs) 
            if [ -n "$CADDY_CONTAINER" ]; then
                docker logs -f "$CADDY_CONTAINER"
            else
                error "No Caddy container running"
                exit 1
            fi
            ;;
        *) 
            echo "Usage: $0 [start|update|reload|verify|logs]"
            exit 1
            ;;
    esac
else
    # Interactive mode
    while true; do
        show_menu
        echo ""
        read -p "Press Enter to continue..."
    done
fi
