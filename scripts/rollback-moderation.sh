#!/bin/bash
set -e

# Rollback Script for Moderation System
# Safely rolls back moderation migrations with data backup and verification

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENVIRONMENT="${ENVIRONMENT:-staging}"
DRY_RUN="${DRY_RUN:-false}"
SKIP_BACKUP="${SKIP_BACKUP:-false}"
TARGET_VERSION=""

# Moderation migrations in reverse order
MODERATION_MIGRATIONS=(
    "97"  # update_moderation_audit_logs
    "69"  # add_forum_moderation
    "50"  # add_moderation_appeals
    "49"  # add_moderation_queue_system
    "11"  # add_moderation_audit_logs
)

# Usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Rollback script for moderation system migrations.

OPTIONS:
    -e, --env ENV          Environment (staging|production) [default: staging]
    --target VERSION       Target migration version to rollback to
    --dry-run              Show what would be done without executing
    --skip-backup          Skip data backup before rollback (NOT RECOMMENDED)
    -h, --help             Show this help message

EXAMPLES:
    $0 --env staging --dry-run
    $0 --env production --target 11  # Rollback to before moderation queue
    $0 --target 10  # Rollback all moderation features

EOF
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --target)
            TARGET_VERSION="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-backup)
            SKIP_BACKUP=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Logging functions
log_header() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  $1"
    echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
}

log_step() {
    echo -e "\n${BLUE}[Step]${NC} $1"
}

log_info() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Load environment
load_environment() {
    log_header "Loading Environment"
    
    if [ -f "$PROJECT_ROOT/backend/.env" ]; then
        log_info "Loading environment from backend/.env"
        set -a
        source "$PROJECT_ROOT/backend/.env"
        set +a
    elif [ -f "$PROJECT_ROOT/.env" ]; then
        log_info "Loading environment from .env"
        set -a
        source "$PROJECT_ROOT/.env"
        set +a
    else
        log_error "No .env file found"
        exit 1
    fi
    
    log_info "Environment: $ENVIRONMENT"
    
    # Validate required variables
    if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
        log_error "Database configuration incomplete"
        exit 1
    fi
    
    # Ensure secure SSL mode configuration
    if [ -z "$DB_SSLMODE" ]; then
        if [ "$ENVIRONMENT" = "production" ] || [ "$ENVIRONMENT" = "prod" ]; then
            log_error "DB_SSLMODE must be set in production to enforce secure database connections"
            exit 1
        else
            DB_SSLMODE="require"
            log_warn "DB_SSLMODE not set, defaulting to 'require' for secure database connection"
        fi
    fi
    
    # Build database URL
    DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
    
    if [ "$DRY_RUN" = true ]; then
        log_warn "DRY RUN MODE - No changes will be made"
    fi
}

# Check migration tool
check_migration_tool() {
    log_header "Migration Tool Check"
    
    if ! command_exists migrate; then
        log_error "golang-migrate not installed"
        log_error "Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        exit 1
    fi
    
    MIGRATE_VERSION=$(migrate -version 2>/dev/null | head -1 || echo "unknown")
    log_info "golang-migrate installed: $MIGRATE_VERSION"
}

# Get current migration version
get_current_version() {
    if [ "$DRY_RUN" = true ]; then
        echo "000097"
        return
    fi
    
    CURRENT_VERSION=$(migrate -path "$PROJECT_ROOT/backend/migrations" -database "$DB_URL" version 2>/dev/null | awk '{print $1}' || echo "0")
    echo "$CURRENT_VERSION"
}

# Backup moderation data
backup_moderation_data() {
    if [ "$SKIP_BACKUP" = true ]; then
        log_warn "Skipping data backup (--skip-backup flag set)"
        return
    fi
    
    log_header "Data Backup"
    
    if [ "$DRY_RUN" = true ]; then
        log_warn "Would backup moderation data"
        return
    fi
    
    BACKUP_DIR="${BACKUP_DIR:-/tmp/clpr-backups}"
    mkdir -p "$BACKUP_DIR"
    BACKUP_FILE="$BACKUP_DIR/moderation-rollback-$(date +%Y%m%d-%H%M%S).sql"
    
    export PGPASSWORD="$DB_PASSWORD"
    
    log_step "Backing up moderation tables..."
    
    # Backup specific moderation tables
    TABLES=("moderation_audit_logs" "moderation_queue" "moderation_decisions" "moderation_appeals")
    
    for table in "${TABLES[@]}"; do
        # Check if table exists
        if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" 2>/dev/null | grep -q 't'; then
            log_info "Backing up table: $table"
            pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t "$table" -F p >> "$BACKUP_FILE" 2>/dev/null || log_warn "Failed to backup $table"
        else
            log_warn "Table not found: $table (skipping)"
        fi
    done
    
    if [ -f "$BACKUP_FILE" ]; then
        log_info "Backup created: $BACKUP_FILE"
        log_info "Backup size: $(du -h "$BACKUP_FILE" | cut -f1)"
        
        # Compress backup
        gzip "$BACKUP_FILE"
        log_info "Backup compressed: ${BACKUP_FILE}.gz"
    else
        log_warn "No backup created (tables may not exist)"
    fi
    
    unset PGPASSWORD
}

# Confirm rollback
confirm_rollback() {
    log_header "Rollback Confirmation"
    
    CURRENT_VERSION=$(get_current_version)
    log_info "Current migration version: $CURRENT_VERSION"
    
    if [ -z "$TARGET_VERSION" ]; then
        # Default: rollback to before first moderation migration (10)
        TARGET_VERSION="10"
        log_info "Target version not specified, defaulting to: $TARGET_VERSION"
    fi
    
    TARGET_PADDED=$(printf "%06d" "$TARGET_VERSION")
    log_info "Target migration version: $TARGET_PADDED"
    
    # Calculate what will be rolled back
    log_info "This will rollback the following features:"
    
    for migration_num in "${MODERATION_MIGRATIONS[@]}"; do
        migration_padded=$(printf "%06d" "$migration_num")
        if [ "$migration_padded" -gt "$TARGET_PADDED" ] && [ "$CURRENT_VERSION" -ge "$migration_padded" ]; then
            case "$migration_num" in
                "11")
                    echo "  - Moderation audit logs"
                    ;;
                "49")
                    echo "  - Moderation queue system"
                    ;;
                "50")
                    echo "  - Moderation appeals"
                    ;;
                "69")
                    echo "  - Forum moderation"
                    ;;
                "97")
                    echo "  - Updated moderation audit logs"
                    ;;
            esac
        fi
    done
    
    if [ "$ENVIRONMENT" = "production" ] && [ "$DRY_RUN" != true ]; then
        echo ""
        echo -e "${RED}WARNING: You are about to rollback migrations in PRODUCTION${NC}"
        echo -e "${RED}This will DELETE moderation tables and ALL moderation data!${NC}"
        echo ""
        read -p "Type 'ROLLBACK' to continue: " confirm
        if [ "$confirm" != "ROLLBACK" ]; then
            echo "Aborted"
            exit 1
        fi
    fi
}

# Run rollback
run_rollback() {
    log_header "Running Rollback"
    
    MIGRATIONS_PATH="$PROJECT_ROOT/backend/migrations"
    
    if [ ! -d "$MIGRATIONS_PATH" ]; then
        log_error "Migrations directory not found: $MIGRATIONS_PATH"
        exit 1
    fi
    
    TARGET_PADDED=$(printf "%06d" "$TARGET_VERSION")
    
    if [ "$DRY_RUN" = true ]; then
        log_warn "Would rollback to version $TARGET_PADDED"
        return
    fi
    
    log_step "Rolling back to version $TARGET_PADDED..."
    
    if migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" goto "$TARGET_PADDED"; then
        log_info "Rollback completed successfully"
    else
        log_error "Rollback failed!"
        log_error "Database may be in dirty state. Check with: migrate -database '$DB_URL' -path '$MIGRATIONS_PATH' version"
        exit 1
    fi
    
    # Verify final version
    FINAL_VERSION=$(get_current_version)
    log_info "Final migration version: $FINAL_VERSION"
    
    # Compare numeric values (remove leading zeros for comparison)
    FINAL_NUM=$(echo "$FINAL_VERSION" | sed 's/^0*//')
    TARGET_NUM=$(echo "$TARGET_VERSION" | sed 's/^0*//')
    
    if [ "$FINAL_NUM" -ne "$TARGET_NUM" ]; then
        log_warn "Final version ($FINAL_VERSION) doesn't match target ($TARGET_PADDED)"
    fi
}

# Verify rollback
verify_rollback() {
    log_header "Rollback Verification"
    
    if [ "$DRY_RUN" = true ]; then
        log_warn "Would verify rollback"
        return
    fi
    
    export PGPASSWORD="$DB_PASSWORD"
    
    # Check that moderation tables are removed (if rolled back completely)
    TARGET_NUM=$(echo "$TARGET_VERSION" | sed 's/^0*//')
    
    if [ "$TARGET_NUM" -lt 11 ]; then
        log_step "Verifying moderation tables are removed..."
        
        TABLES=("moderation_audit_logs" "moderation_queue" "moderation_decisions" "moderation_appeals")
        ALL_REMOVED=true
        
        for table in "${TABLES[@]}"; do
            if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" 2>/dev/null | grep -q 't'; then
                log_warn "Table still exists: $table"
                ALL_REMOVED=false
            fi
        done
        
        if [ "$ALL_REMOVED" = true ]; then
            log_info "All moderation tables removed successfully"
        else
            log_warn "Some moderation tables still exist (may be expected)"
        fi
    fi
    
    # Check database is in clean state
    log_step "Checking migration state..."
    
    IS_DIRTY=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT dirty FROM schema_migrations WHERE version = '$(get_current_version)';" 2>/dev/null | xargs)
    
    if [ "$IS_DIRTY" = "f" ]; then
        log_info "Migration state: clean"
    else
        log_error "Migration state: dirty (manual intervention required)"
    fi
    
    unset PGPASSWORD
}

# Generate summary
generate_summary() {
    log_header "Rollback Summary"
    
    echo ""
    echo "Environment: $ENVIRONMENT"
    echo "Dry Run: $DRY_RUN"
    echo ""
    
    if [ "$DRY_RUN" = true ]; then
        echo -e "${YELLOW}This was a DRY RUN - no changes were made${NC}"
        echo "Run without --dry-run to execute rollback"
    else
        FINAL_VERSION=$(get_current_version)
        echo -e "${GREEN}✓ Rollback completed successfully${NC}"
        echo "Current migration version: $FINAL_VERSION"
        echo ""
        echo "Next steps:"
        echo "1. Verify application still functions correctly"
        echo "2. Check logs for any errors"
        echo "3. If needed, restore from backup: ${BACKUP_DIR}/moderation-rollback-*.sql.gz"
    fi
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════╗"
    echo "║  Moderation Rollback Script v1.0      ║"
    echo "╚════════════════════════════════════════╝"
    echo -e "${NC}"
    
    # Run rollback steps
    load_environment
    check_migration_tool
    confirm_rollback
    backup_moderation_data
    run_rollback
    verify_rollback
    generate_summary
    
    echo ""
    log_info "Rollback process complete"
}

# Run main function
main
