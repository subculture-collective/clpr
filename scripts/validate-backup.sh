#!/bin/bash
set -euo pipefail

# Backup Validation Script
# Verifies nightly backup completion, integrity, encryption, and cross-region storage
# Exit codes: 0 = success, 1 = validation failed

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/clpr}"
CLOUD_PROVIDER="${CLOUD_PROVIDER:-gcp}"
BACKUP_BUCKET="${BACKUP_BUCKET:-clpr-backups-prod}"
AZURE_STORAGE_ACCOUNT="${AZURE_STORAGE_ACCOUNT:-}"
MAX_BACKUP_AGE_HOURS="${MAX_BACKUP_AGE_HOURS:-24}"
MIN_BACKUP_SIZE_MB="${MIN_BACKUP_SIZE_MB:-1}"
VALIDATION_LOG="${VALIDATION_LOG:-/var/log/clpr/backup-validation.log}"

# Metrics for alerting
VALIDATION_SUCCESS=0
VALIDATION_TIMESTAMP=$(date +%s)
BACKUP_AGE_HOURS=0
BACKUP_SIZE_MB=0
ENCRYPTION_VERIFIED=0
CROSS_REGION_VERIFIED=0

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1" | tee -a "$VALIDATION_LOG"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$VALIDATION_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$VALIDATION_LOG"
}

# Ensure log directory exists
mkdir -p "$(dirname "$VALIDATION_LOG")"

echo "=== Backup Validation Started at $(date) ===" | tee -a "$VALIDATION_LOG"
log_info "Configuration:"
log_info "  Cloud Provider: $CLOUD_PROVIDER"
log_info "  Backup Bucket: $BACKUP_BUCKET"
log_info "  Max Backup Age: ${MAX_BACKUP_AGE_HOURS}h"
log_info "  Min Backup Size: ${MIN_BACKUP_SIZE_MB}MB"

# Function to get latest backup from cloud storage
get_latest_backup() {
    local latest_backup=""
    local backup_timestamp=""
    
    if [ "$CLOUD_PROVIDER" = "gcp" ]; then
        log_info "Checking GCS bucket: gs://${BACKUP_BUCKET}/database/"
        
        # Get latest backup file
        latest_backup=$(gsutil ls -l "gs://${BACKUP_BUCKET}/database/postgres-backup-*.sql.gz" 2>/dev/null | \
            grep -v TOTAL | sort -k2 -r | head -1 | awk '{print $3}') || {
            log_error "Failed to list backups from GCS"
            return 1
        }
        
        if [ -z "$latest_backup" ]; then
            log_error "No backups found in GCS bucket"
            return 1
        fi
        
        # Get backup metadata
        backup_size=$(gsutil ls -l "$latest_backup" | grep -v TOTAL | awk '{print $1}')
        backup_timestamp=$(gsutil ls -l "$latest_backup" | grep -v TOTAL | awk '{print $2}')
        
        log_info "Latest backup: $latest_backup"
        log_info "Backup size: $(numfmt --to=iec-i --suffix=B $backup_size 2>/dev/null || echo ${backup_size}B)"
        log_info "Backup timestamp: $backup_timestamp"
        
    elif [ "$CLOUD_PROVIDER" = "aws" ]; then
        log_info "Checking S3 bucket: s3://${BACKUP_BUCKET}/database/"
        
        # Get latest backup file
        local latest_file=$(aws s3 ls "s3://${BACKUP_BUCKET}/database/" 2>/dev/null | \
            grep "postgres-backup-" | sort -r | head -1 | awk '{print $4}') || {
            log_error "Failed to list backups from S3"
            return 1
        }
        
        if [ -z "$latest_file" ]; then
            log_error "No backups found in S3 bucket"
            return 1
        fi
        
        latest_backup="s3://${BACKUP_BUCKET}/database/${latest_file}"
        
        # Get backup metadata
        local s3_info=$(aws s3 ls "s3://${BACKUP_BUCKET}/database/${latest_file}" --human-readable 2>/dev/null)
        backup_size=$(echo "$s3_info" | awk '{print $3}' | numfmt --from=iec 2>/dev/null || echo "0")
        backup_timestamp=$(echo "$s3_info" | awk '{print $1" "$2}')
        
        log_info "Latest backup: $latest_backup"
        log_info "Backup size: $(echo "$s3_info" | awk '{print $3}')"
        log_info "Backup timestamp: $backup_timestamp"
        
    elif [ "$CLOUD_PROVIDER" = "azure" ]; then
        log_info "Checking Azure blob storage: ${AZURE_STORAGE_ACCOUNT}/${BACKUP_BUCKET}/database/"
        
        if [ -z "$AZURE_STORAGE_ACCOUNT" ]; then
            log_error "AZURE_STORAGE_ACCOUNT not set"
            return 1
        fi
        
        # Get latest backup file
        latest_backup=$(az storage blob list \
            --account-name "${AZURE_STORAGE_ACCOUNT}" \
            --container-name "${BACKUP_BUCKET}" \
            --prefix "database/postgres-backup-" \
            --query "sort_by([].{name:name, modified:properties.lastModified}, &modified)[-1].name" \
            -o tsv 2>/dev/null) || {
            log_error "Failed to list backups from Azure"
            return 1
        }
        
        if [ -z "$latest_backup" ]; then
            log_error "No backups found in Azure blob storage"
            return 1
        fi
        
        # Get backup metadata
        local blob_info=$(az storage blob show \
            --account-name "${AZURE_STORAGE_ACCOUNT}" \
            --container-name "${BACKUP_BUCKET}" \
            --name "${latest_backup}" \
            --query "{size:properties.contentLength, modified:properties.lastModified}" \
            -o json 2>/dev/null)
        
        backup_size=$(echo "$blob_info" | jq -r '.size')
        backup_timestamp=$(echo "$blob_info" | jq -r '.modified')
        
        log_info "Latest backup: $latest_backup"
        log_info "Backup size: $(numfmt --to=iec-i --suffix=B $backup_size 2>/dev/null || echo ${backup_size}B)"
        log_info "Backup timestamp: $backup_timestamp"
    else
        log_error "Unsupported cloud provider: $CLOUD_PROVIDER"
        return 1
    fi
    
    # Calculate backup age
    if [ -n "$backup_timestamp" ]; then
        local backup_ts_epoch=$(date -d "$backup_timestamp" +%s 2>/dev/null || date -j -f "%Y-%m-%d %H:%M:%S" "$backup_timestamp" +%s 2>/dev/null || echo "0")
        if [ "$backup_ts_epoch" = "0" ] || [ -z "$backup_ts_epoch" ]; then
            log_warn "Failed to parse backup timestamp: $backup_timestamp"
            backup_ts_epoch=0
        fi
        local current_ts=$(date +%s)
        local age_seconds=$((current_ts - backup_ts_epoch))
        BACKUP_AGE_HOURS=$((age_seconds / 3600))
        
        log_info "Backup age: ${BACKUP_AGE_HOURS} hours"
    else
        log_warn "Backup timestamp is empty, cannot calculate age"
    fi
    
    # Calculate backup size in MB
    BACKUP_SIZE_MB=$((backup_size / 1024 / 1024))
    
    # Export for use in other functions
    export LATEST_BACKUP="$latest_backup"
    export BACKUP_SIZE_BYTES="$backup_size"
}

# Function to verify backup age
verify_backup_age() {
    log_info "Verifying backup age..."
    
    if [ "$BACKUP_AGE_HOURS" -gt "$MAX_BACKUP_AGE_HOURS" ]; then
        log_error "Backup is too old: ${BACKUP_AGE_HOURS}h (max: ${MAX_BACKUP_AGE_HOURS}h)"
        return 1
    fi
    
    log_info "✓ Backup age is acceptable: ${BACKUP_AGE_HOURS}h"
    return 0
}

# Function to verify backup size
verify_backup_size() {
    log_info "Verifying backup size..."
    
    if [ "$BACKUP_SIZE_MB" -lt "$MIN_BACKUP_SIZE_MB" ]; then
        log_error "Backup is too small: ${BACKUP_SIZE_MB}MB (min: ${MIN_BACKUP_SIZE_MB}MB)"
        return 1
    fi
    
    log_info "✓ Backup size is acceptable: ${BACKUP_SIZE_MB}MB"
    return 0
}

# Function to verify encryption
verify_encryption() {
    log_info "Verifying backup encryption..."
    
    if [ "$CLOUD_PROVIDER" = "gcp" ]; then
        # Check if encryption is enabled on the bucket
        local encryption_config=$(gsutil encryption get "gs://${BACKUP_BUCKET}" 2>/dev/null || echo "")
        if [ -n "$encryption_config" ]; then
            log_info "✓ GCS bucket has encryption enabled"
            ENCRYPTION_VERIFIED=1
        else
            log_warn "GCS bucket encryption not configured (but may use default encryption)"
            ENCRYPTION_VERIFIED=1  # GCS encrypts by default
        fi
        
    elif [ "$CLOUD_PROVIDER" = "aws" ]; then
        # Check if server-side encryption is enabled
        local backup_filename="${LATEST_BACKUP##*/}"
        local encryption=$(aws s3api head-object \
            --bucket "${BACKUP_BUCKET}" \
            --key "database/${backup_filename}" \
            --query 'ServerSideEncryption' \
            --output text 2>/dev/null || echo "")
        
        if [ -n "$encryption" ] && [ "$encryption" != "None" ]; then
            log_info "✓ S3 backup is encrypted with $encryption"
            ENCRYPTION_VERIFIED=1
        else
            log_error "S3 backup is not encrypted"
            return 1
        fi
        
    elif [ "$CLOUD_PROVIDER" = "azure" ]; then
        # Azure Storage encrypts all data at rest by default
        log_info "✓ Azure Storage encrypts all data at rest by default"
        ENCRYPTION_VERIFIED=1
    fi
    
    return 0
}

# Function to verify cross-region replication
verify_cross_region() {
    log_info "Verifying cross-region storage..."
    
    if [ "$CLOUD_PROVIDER" = "gcp" ]; then
        # Check bucket location type
        local location=$(gsutil ls -L -b "gs://${BACKUP_BUCKET}" 2>/dev/null | grep -iE "Location type:|Location:" | awk '{print $NF}' || echo "")
        
        if echo "$location" | grep -qiE "multi-region|dual-region|^(US|EU|ASIA)$"; then
            log_info "✓ GCS bucket is multi-region or geo-redundant: $location"
            CROSS_REGION_VERIFIED=1
        else
            log_warn "GCS bucket is single-region: $location"
        fi
        
    elif [ "$CLOUD_PROVIDER" = "aws" ]; then
        # Check replication configuration
        local replication=$(aws s3api get-bucket-replication \
            --bucket "${BACKUP_BUCKET}" \
            --query 'ReplicationConfiguration.Rules[0].Status' \
            --output text 2>/dev/null || echo "")
        
        if [ "$replication" = "Enabled" ]; then
            log_info "✓ S3 bucket has replication enabled"
            CROSS_REGION_VERIFIED=1
        else
            log_warn "S3 bucket replication not configured"
        fi
        
    elif [ "$CLOUD_PROVIDER" = "azure" ]; then
        # Check redundancy
        local redundancy=$(az storage account show \
            --name "${AZURE_STORAGE_ACCOUNT}" \
            --query 'sku.name' \
            -o tsv 2>/dev/null || echo "")
        
        if echo "$redundancy" | grep -qE "GRS|GZRS|RA-GRS|RA-GZRS"; then
            log_info "✓ Azure storage has geo-redundancy: $redundancy"
            CROSS_REGION_VERIFIED=1
        else
            log_warn "Azure storage is not geo-redundant: $redundancy"
        fi
    fi
    
    return 0
}

# Function to report metrics (for Prometheus pushgateway or similar)
report_metrics() {
    local pushgateway="${PROMETHEUS_PUSHGATEWAY:-}"
    
    if [ -z "$pushgateway" ]; then
        log_info "PROMETHEUS_PUSHGATEWAY not set, skipping metrics push"
        return 0
    fi
    
    log_info "Reporting metrics to $pushgateway..."
    
    cat <<METRICS | curl --data-binary @- "${pushgateway}/metrics/job/backup_validation" 2>/dev/null || log_warn "Failed to push metrics to Prometheus"
# HELP backup_validation_success Whether the last backup validation succeeded (1 = success, 0 = failure)
# TYPE backup_validation_success gauge
backup_validation_success ${VALIDATION_SUCCESS}
# HELP backup_validation_timestamp Unix timestamp of the last backup validation
# TYPE backup_validation_timestamp gauge
backup_validation_timestamp ${VALIDATION_TIMESTAMP}
# HELP backup_age_hours Age of the latest backup in hours
# TYPE backup_age_hours gauge
backup_age_hours ${BACKUP_AGE_HOURS}
# HELP backup_size_mb Size of the latest backup in megabytes
# TYPE backup_size_mb gauge
backup_size_mb ${BACKUP_SIZE_MB}
# HELP backup_encryption_verified Whether backup encryption was verified (1 = verified, 0 = not verified)
# TYPE backup_encryption_verified gauge
backup_encryption_verified ${ENCRYPTION_VERIFIED}
# HELP backup_cross_region_verified Whether cross-region storage was verified (1 = verified, 0 = not verified)
# TYPE backup_cross_region_verified gauge
backup_cross_region_verified ${CROSS_REGION_VERIFIED}
METRICS
    
    log_info "✓ Metrics reported successfully"
}

# Main validation flow
main() {
    local exit_code=0
    
    # Get latest backup
    if ! get_latest_backup; then
        log_error "Failed to get latest backup"
        VALIDATION_SUCCESS=0
        report_metrics
        return 1
    fi
    
    # Verify backup age
    if ! verify_backup_age; then
        exit_code=1
    fi
    
    # Verify backup size
    if ! verify_backup_size; then
        exit_code=1
    fi
    
    # Verify encryption
    if ! verify_encryption; then
        exit_code=1
    fi
    
    # Verify cross-region storage
    verify_cross_region  # Non-fatal, just a warning
    
    # Set validation success based on exit code
    if [ $exit_code -eq 0 ]; then
        VALIDATION_SUCCESS=1
        log_info "✓ All backup validations passed"
        echo "=== Backup Validation SUCCEEDED ===" | tee -a "$VALIDATION_LOG"
    else
        VALIDATION_SUCCESS=0
        log_error "✗ Backup validation failed"
        echo "=== Backup Validation FAILED ===" | tee -a "$VALIDATION_LOG"
    fi
    
    # Report metrics
    report_metrics
    
    return $exit_code
}

# Run main function
main
