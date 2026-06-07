#!/bin/bash
set -e

# Pre-flight Check Script for Moderation System Migration
# Validates that the system is ready for moderation feature deployment

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
REPORT_FILE=""

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Moderation-specific migrations to check
MODERATION_MIGRATIONS=(
    "000011_add_moderation_audit_logs"
    "000049_add_moderation_queue_system"
    "000050_add_moderation_appeals"
    "000069_add_forum_moderation"
    "000097_update_moderation_audit_logs"
)

# Usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Pre-flight check script for moderation system deployment validation.

OPTIONS:
    -e, --env ENV          Environment to check (staging|production) [default: staging]
    -r, --report FILE      Generate report to file
    -h, --help             Show this help message

EXAMPLES:
    $0 --env production
    $0 --env staging --report preflight-moderation.txt

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
        -r|--report)
            REPORT_FILE="$2"
            shift 2
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
    echo -e "\n${BLUE}=== $1 ===${NC}"
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

log_check() {
    echo -e "${BLUE}[•]${NC} Checking: $1"
}

# Check result tracking
check_pass() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
    log_info "$1"
}

check_fail() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
    log_error "$1"
}

check_warn() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    WARNING_CHECKS=$((WARNING_CHECKS + 1))
    log_warn "$1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Load environment variables
load_environment() {
    log_header "Loading Environment"
    
    # Check for .env file
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
        check_warn "No .env file found, using system environment variables"
    fi
    
    log_info "Environment: $ENVIRONMENT"
}

# Check moderation migration files exist
check_migration_files() {
    log_header "Migration Files"
    
    MIGRATIONS_DIR="$PROJECT_ROOT/backend/migrations"
    
    if [ ! -d "$MIGRATIONS_DIR" ]; then
        check_fail "Migrations directory not found: $MIGRATIONS_DIR"
        return
    fi
    
    check_pass "Migrations directory exists: $MIGRATIONS_DIR"
    
    # Check each moderation migration
    for migration in "${MODERATION_MIGRATIONS[@]}"; do
        log_check "Migration $migration"
        
        up_file=$(ls "$MIGRATIONS_DIR"/*"$migration.up.sql" 2>/dev/null | head -1)
        down_file=$(ls "$MIGRATIONS_DIR"/*"$migration.down.sql" 2>/dev/null | head -1)
        
        if [ -f "$up_file" ] && [ -f "$down_file" ]; then
            check_pass "Migration files present: $migration"
        else
            check_fail "Migration files missing: $migration"
        fi
    done
}

# Check database prerequisites
check_database_prerequisites() {
    log_header "Database Prerequisites"
    
    if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_USER" ] || [ -z "$DB_NAME" ]; then
        check_fail "Database configuration incomplete, skipping checks"
        return
    fi
    
    # Check if psql is available
    if ! command_exists psql; then
        check_warn "psql not installed, skipping database prerequisite checks"
        return
    fi
    
    # Build connection string
    export PGPASSWORD="$DB_PASSWORD"
    
    # Test connection
    log_check "Database connection"
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
        check_pass "Database connection successful"
    else
        check_fail "Database connection failed"
        unset PGPASSWORD
        return
    fi
    
    # Check PostgreSQL version (need 12+ for better JSON support)
    log_check "PostgreSQL version"
    DB_VERSION=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SHOW server_version;" 2>/dev/null | xargs | cut -d' ' -f1 | cut -d'.' -f1)
    if [ -n "$DB_VERSION" ] && [ "$DB_VERSION" -ge 12 ]; then
        check_pass "PostgreSQL version: $DB_VERSION (>= 12 required)"
    else
        check_fail "PostgreSQL version too old: $DB_VERSION (>= 12 required)"
    fi
    
    # Check if users table exists (dependency)
    log_check "Users table exists"
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users');" 2>/dev/null | grep -q 't'; then
        check_pass "Users table exists"
    else
        check_fail "Users table missing (required for moderation system)"
    fi
    
    # Check if clips table exists (dependency)
    log_check "Clips table exists"
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'clips');" 2>/dev/null | grep -q 't'; then
        check_pass "Clips table exists"
    else
        check_warn "Clips table missing (moderation system works with clips)"
    fi
    
    # Check if comments table exists (dependency)
    log_check "Comments table exists"
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'comments');" 2>/dev/null | grep -q 't'; then
        check_pass "Comments table exists"
    else
        check_warn "Comments table missing (moderation system works with comments)"
    fi
    
    # Check current migration status
    log_check "Migration status"
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'schema_migrations');" 2>/dev/null | grep -q 't'; then
        CURRENT_VERSION=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null | xargs)
        if [ -n "$CURRENT_VERSION" ]; then
            check_pass "Current migration version: $CURRENT_VERSION"
            
            # Check if dirty
            IS_DIRTY=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT dirty FROM schema_migrations WHERE version = '$CURRENT_VERSION';" 2>/dev/null | xargs)
            if [ "$IS_DIRTY" = "t" ]; then
                check_fail "Database migration is in dirty state! Must be fixed before proceeding"
            else
                check_pass "Migration state: clean"
            fi
        fi
    else
        check_warn "schema_migrations table not found (may be fresh database)"
    fi
    
    # Check if moderation tables already exist
    log_check "Existing moderation tables"
    EXISTING_MOD_TABLES=0
    
    for table in "moderation_audit_logs" "moderation_queue" "moderation_decisions" "moderation_appeals"; do
        if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" 2>/dev/null | grep -q 't'; then
            EXISTING_MOD_TABLES=$((EXISTING_MOD_TABLES + 1))
        fi
    done
    
    if [ $EXISTING_MOD_TABLES -eq 0 ]; then
        check_pass "No existing moderation tables (clean state for migration)"
    elif [ $EXISTING_MOD_TABLES -eq 4 ]; then
        check_warn "All moderation tables already exist (may be already migrated)"
    else
        check_warn "Partial moderation tables exist ($EXISTING_MOD_TABLES/4) - check migration state"
    fi
    
    unset PGPASSWORD
}

# Check golang-migrate tool
check_migration_tool() {
    log_header "Migration Tool"
    
    log_check "golang-migrate installation"
    if command_exists migrate; then
        MIGRATE_VERSION=$(migrate -version 2>/dev/null | head -1 || echo "unknown")
        check_pass "golang-migrate installed: $MIGRATE_VERSION"
    else
        check_fail "golang-migrate not installed (required for running migrations)"
        log_error "Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    fi
}

# Check disk space for migration
check_disk_space() {
    log_header "System Resources"
    
    log_check "Disk space"
    DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    DISK_AVAILABLE=$(df -h / | awk 'NR==2 {print $4}')
    
    if [ "$DISK_USAGE" -lt 90 ]; then
        check_pass "Disk space: ${DISK_USAGE}% used, ${DISK_AVAILABLE} available"
    else
        check_fail "Disk space critically low: ${DISK_USAGE}% used (cleanup required before migration)"
    fi
}

# Check backup capability
check_backup_capability() {
    log_header "Backup Capability"
    
    log_check "Backup script"
    BACKUP_SCRIPT="$SCRIPT_DIR/backup.sh"
    if [ -f "$BACKUP_SCRIPT" ] && [ -x "$BACKUP_SCRIPT" ]; then
        check_pass "Backup script found and executable: $BACKUP_SCRIPT"
    else
        check_warn "Backup script not found or not executable: $BACKUP_SCRIPT"
    fi
    
    log_check "Backup directory"
    BACKUP_DIR="${BACKUP_DIR:-/var/backups/clpr}"
    if [ -d "$BACKUP_DIR" ] && [ -w "$BACKUP_DIR" ]; then
        check_pass "Backup directory writable: $BACKUP_DIR"
    elif [ ! -d "$BACKUP_DIR" ]; then
        check_warn "Backup directory doesn't exist: $BACKUP_DIR (will be created)"
    else
        check_warn "Backup directory not writable: $BACKUP_DIR"
    fi
}

# Check for data conflicts
check_data_conflicts() {
    log_header "Data Conflict Checks"
    
    if [ -z "$DB_HOST" ] || ! command_exists psql; then
        check_warn "Skipping data conflict checks (database not accessible)"
        return
    fi
    
    export PGPASSWORD="$DB_PASSWORD"
    
    # Check for users with moderator role
    log_check "Moderator users"
    MOD_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE role IN ('moderator', 'admin') AND is_banned = false;" 2>/dev/null | xargs || echo "0")
    if [ "$MOD_COUNT" -gt 0 ]; then
        check_pass "Found $MOD_COUNT active moderator/admin users"
    else
        check_warn "No active moderator/admin users found (moderation system requires moderators)"
    fi
    
    # Check for existing reports table (might conflict)
    log_check "Existing reports"
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'reports');" 2>/dev/null | grep -q 't'; then
        REPORT_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM reports;" 2>/dev/null | xargs || echo "0")
        if [ "$REPORT_COUNT" -gt 0 ]; then
            check_pass "Found $REPORT_COUNT existing reports (will need to be migrated to moderation queue)"
        else
            check_pass "Reports table exists but is empty"
        fi
    fi
    
    unset PGPASSWORD
}

# Generate summary
generate_summary() {
    log_header "Pre-flight Check Summary"
    
    echo ""
    echo "Environment: $ENVIRONMENT"
    echo "Target: Moderation System Migration"
    echo ""
    echo "Total Checks: $TOTAL_CHECKS"
    echo -e "  ${GREEN}Passed: $PASSED_CHECKS${NC}"
    echo -e "  ${YELLOW}Warnings: $WARNING_CHECKS${NC}"
    echo -e "  ${RED}Failed: $FAILED_CHECKS${NC}"
    echo ""
    
    if [ $FAILED_CHECKS -eq 0 ]; then
        if [ $WARNING_CHECKS -eq 0 ]; then
            echo -e "${GREEN}✓ All pre-flight checks passed!${NC}"
            echo "Moderation system migration may proceed."
            return 0
        else
            echo -e "${YELLOW}⚠ Pre-flight checks passed with warnings.${NC}"
            echo "Review warnings before proceeding with migration."
            return 0
        fi
    else
        echo -e "${RED}✗ Pre-flight checks failed!${NC}"
        echo "Fix all failed checks before deploying moderation system to $ENVIRONMENT."
        return 1
    fi
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════╗"
    echo "║ Moderation Migration Pre-flight v1.0  ║"
    echo "╚════════════════════════════════════════╝"
    echo -e "${NC}"
    
    # Redirect output to report file if specified
    if [ -n "$REPORT_FILE" ]; then
        exec > >(tee "$REPORT_FILE")
    fi
    
    # Run checks
    load_environment
    check_migration_files
    check_migration_tool
    check_database_prerequisites
    check_disk_space
    check_backup_capability
    check_data_conflicts
    
    # Generate summary
    generate_summary
    EXIT_CODE=$?
    
    if [ -n "$REPORT_FILE" ]; then
        echo ""
        echo "Report saved to: $REPORT_FILE"
    fi
    
    exit $EXIT_CODE
}

# Run main function
main
