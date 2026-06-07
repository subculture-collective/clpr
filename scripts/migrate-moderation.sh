#!/bin/bash
set -e

# Moderation System Migration Runner
# Runs moderation-specific migrations with validation and rollback support

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
SKIP_VALIDATION="${SKIP_VALIDATION:-false}"

# Moderation migrations in order
MODERATION_MIGRATIONS=(
    "11"  # add_moderation_audit_logs
    "49"  # add_moderation_queue_system
    "50"  # add_moderation_appeals
    "69"  # add_forum_moderation
    "97"  # update_moderation_audit_logs
)

# Usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Moderation system migration runner with validation and rollback support.

OPTIONS:
    -e, --env ENV          Environment (staging|production) [default: staging]
    --dry-run              Show what would be done without executing
    --skip-backup          Skip database backup before migration
    --skip-validation      Skip post-migration validation
    -h, --help             Show this help message

EXAMPLES:
    $0 --env staging --dry-run
    $0 --env production
    $0 --skip-backup  # Not recommended for production!

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
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --skip-backup)
            SKIP_BACKUP=true
            shift
            ;;
        --skip-validation)
            SKIP_VALIDATION=true
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

# Run pre-flight checks
run_preflight_checks() {
    log_header "Pre-flight Checks"
    
    PREFLIGHT_SCRIPT="$SCRIPT_DIR/preflight-moderation.sh"
    
    if [ ! -f "$PREFLIGHT_SCRIPT" ]; then
        log_error "Pre-flight check script not found: $PREFLIGHT_SCRIPT"
        exit 1
    fi
    
    log_step "Running pre-flight checks..."
    
    if [ "$DRY_RUN" = true ]; then
        log_warn "Skipping pre-flight checks in dry-run mode"
        return
    fi
    
    if ! bash "$PREFLIGHT_SCRIPT" --env "$ENVIRONMENT"; then
        log_error "Pre-flight checks failed"
        exit 1
    fi
    
    log_info "Pre-flight checks passed"
}

# Create backup
create_backup() {
    if [ "$SKIP_BACKUP" = true ]; then
        log_warn "Skipping backup (--skip-backup flag set)"
        return
    fi
    
    log_header "Database Backup"
    
    BACKUP_SCRIPT="$SCRIPT_DIR/backup.sh"
    
    if [ "$DRY_RUN" = true ]; then
        log_warn "Would create backup with: $BACKUP_SCRIPT"
        return
    fi
    
    if [ ! -f "$BACKUP_SCRIPT" ]; then
        log_warn "Backup script not found: $BACKUP_SCRIPT"
        log_warn "Creating manual backup..."
        
        # Manual backup
        BACKUP_DIR="${BACKUP_DIR:-/tmp/clpr-backups}"
        mkdir -p "$BACKUP_DIR"
        BACKUP_FILE="$BACKUP_DIR/moderation-pre-migration-$(date +%Y%m%d-%H%M%S).sql"
        
        export PGPASSWORD="$DB_PASSWORD"
        log_step "Creating backup: $BACKUP_FILE"
        
        if pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -F c -f "$BACKUP_FILE"; then
            log_info "Backup created: $BACKUP_FILE"
            log_info "Backup size: $(du -h "$BACKUP_FILE" | cut -f1)"
        else
            log_error "Backup failed"
            unset PGPASSWORD
            exit 1
        fi
        
        unset PGPASSWORD
    else
        log_step "Creating backup with backup.sh..."
        if bash "$BACKUP_SCRIPT"; then
            log_info "Backup created successfully"
        else
            log_error "Backup failed"
            exit 1
        fi
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
        echo "000000"
        return
    fi
    
    CURRENT_VERSION=$(migrate -path "$PROJECT_ROOT/backend/migrations" -database "$DB_URL" version 2>/dev/null | awk '{print $1}' || echo "0")
    echo "$CURRENT_VERSION"
}

# Check if migration is already applied
is_migration_applied() {
    local migration_num=$1
    local current_version=$(get_current_version)
    
    # Pad migration number to 6 digits
    local migration_version=$(printf "%06d" "$migration_num")
    
    if [ "$current_version" -ge "$migration_version" ]; then
        return 0  # Already applied
    else
        return 1  # Not applied
    fi
}

# Run migrations
run_migrations() {
    log_header "Running Moderation Migrations"
    
    MIGRATIONS_PATH="$PROJECT_ROOT/backend/migrations"
    
    if [ ! -d "$MIGRATIONS_PATH" ]; then
        log_error "Migrations directory not found: $MIGRATIONS_PATH"
        exit 1
    fi
    
    CURRENT_VERSION=$(get_current_version)
    log_info "Current migration version: $CURRENT_VERSION"
    
    # Check which moderation migrations need to be applied
    MIGRATIONS_TO_APPLY=()
    for migration_num in "${MODERATION_MIGRATIONS[@]}"; do
        if ! is_migration_applied "$migration_num"; then
            MIGRATIONS_TO_APPLY+=("$migration_num")
        fi
    done
    
    if [ ${#MIGRATIONS_TO_APPLY[@]} -eq 0 ]; then
        log_info "All moderation migrations already applied"
        return
    fi
    
    log_info "Migrations to apply: ${MIGRATIONS_TO_APPLY[*]}"
    
    if [ "$DRY_RUN" = true ]; then
        log_warn "Would run migrations up to include all moderation migrations"
        return
    fi
    
    # Calculate target version (highest moderation migration)
    TARGET_VERSION=$(printf "%06d" "${MODERATION_MIGRATIONS[-1]}")
    
    log_step "Running migrations up to version $TARGET_VERSION..."
    
    if migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" goto "$TARGET_VERSION"; then
        log_info "Migrations completed successfully"
    else
        log_error "Migration failed!"
        log_error "Database may be in dirty state. Check with: migrate -database '$DB_URL' -path '$MIGRATIONS_PATH' version"
        exit 1
    fi
    
    # Verify final version
    FINAL_VERSION=$(get_current_version)
    log_info "Final migration version: $FINAL_VERSION"
}

# Validate schema
validate_schema() {
    if [ "$SKIP_VALIDATION" = true ]; then
        log_warn "Skipping validation (--skip-validation flag set)"
        return
    fi
    
    log_header "Schema Validation"
    
    if [ "$DRY_RUN" = true ]; then
        log_warn "Would validate schema"
        return
    fi
    
    VALIDATION_SCRIPT="$SCRIPT_DIR/validate-moderation.sh"
    
    if [ ! -f "$VALIDATION_SCRIPT" ]; then
        log_warn "Validation script not found: $VALIDATION_SCRIPT"
        log_step "Running basic validation..."
        
        export PGPASSWORD="$DB_PASSWORD"
        
        # Check moderation tables exist
        TABLES=("moderation_audit_logs" "moderation_queue" "moderation_decisions" "moderation_appeals")
        
        for table in "${TABLES[@]}"; do
            if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" 2>/dev/null | grep -q 't'; then
                log_info "Table exists: $table"
            else
                log_error "Table missing: $table"
                unset PGPASSWORD
                exit 1
            fi
        done
        
        unset PGPASSWORD
        log_info "Basic validation passed"
    else
        log_step "Running validation script..."
        if bash "$VALIDATION_SCRIPT" --env "$ENVIRONMENT"; then
            log_info "Validation passed"
        else
            log_error "Validation failed"
            exit 1
        fi
    fi
}

# Generate summary
generate_summary() {
    log_header "Migration Summary"
    
    echo ""
    echo "Environment: $ENVIRONMENT"
    echo "Dry Run: $DRY_RUN"
    echo ""
    
    if [ "$DRY_RUN" = true ]; then
        echo -e "${YELLOW}This was a DRY RUN - no changes were made${NC}"
        echo "Run without --dry-run to apply migrations"
    else
        FINAL_VERSION=$(get_current_version)
        echo -e "${GREEN}✓ Moderation migrations completed successfully${NC}"
        echo "Current migration version: $FINAL_VERSION"
        echo ""
        echo "Next steps:"
        echo "1. Run smoke tests on moderation features"
        echo "2. Monitor application logs for errors"
        echo "3. Verify moderation queue is accessible"
    fi
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════╗"
    echo "║  Moderation Migration Runner v1.0     ║"
    echo "╚════════════════════════════════════════╝"
    echo -e "${NC}"
    
    # Confirm production deployment
    if [ "$ENVIRONMENT" = "production" ] && [ "$DRY_RUN" != true ]; then
        echo ""
        echo -e "${RED}WARNING: You are about to run migrations in PRODUCTION${NC}"
        echo ""
        read -p "Type 'yes' to continue: " confirm
        if [ "$confirm" != "yes" ]; then
            echo "Aborted"
            exit 1
        fi
    fi
    
    # Run migration steps
    load_environment
    run_preflight_checks
    check_migration_tool
    create_backup
    run_migrations
    validate_schema
    generate_summary
    
    echo ""
    log_info "Migration process complete"
}

# Run main function
main
