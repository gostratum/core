#!/bin/bash
# Example deployment script for gostratum/core applications
# This script demonstrates how to use ENV_PREFIX for configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
APP_NAME="${1:-myapp}"
ENVIRONMENT="${2:-dev}"
ENV_PREFIX="${3:-STRATUM}"

log_info "Deploying $APP_NAME in $ENVIRONMENT environment with prefix $ENV_PREFIX"

# Set environment-specific variables
case $ENVIRONMENT in
    prod)
        log_info "Setting production configuration..."
        export APP_ENV=prod
        export ENV_PREFIX=$ENV_PREFIX
        export ${ENV_PREFIX}_APP_PORT=80
        export ${ENV_PREFIX}_APP_HOST=0.0.0.0
        export ${ENV_PREFIX}_APP_LOG_LEVEL=warn
        ;;
    staging)
        log_info "Setting staging configuration..."
        export APP_ENV=staging
        export ENV_PREFIX=$ENV_PREFIX
        export ${ENV_PREFIX}_APP_PORT=8080
        export ${ENV_PREFIX}_APP_HOST=0.0.0.0
        export ${ENV_PREFIX}_APP_LOG_LEVEL=info
        ;;
    dev)
        log_info "Setting development configuration..."
        export APP_ENV=dev
        export ENV_PREFIX=$ENV_PREFIX
        export ${ENV_PREFIX}_APP_PORT=8080
        export ${ENV_PREFIX}_APP_HOST=localhost
        export ${ENV_PREFIX}_APP_LOG_LEVEL=debug
        ;;
    *)
        log_error "Unknown environment: $ENVIRONMENT"
        log_error "Valid environments: prod, staging, dev"
        exit 1
        ;;
esac

# Display configuration
log_info "Configuration:"
echo "  APP_ENV: $APP_ENV"
echo "  ENV_PREFIX: $ENV_PREFIX"
env | grep "^${ENV_PREFIX}_" | sort

# Check if application binary exists
if [ ! -f "./$APP_NAME" ]; then
    log_error "Application binary ./$APP_NAME not found"
    exit 1
fi

# Start application
log_info "Starting $APP_NAME..."
exec ./$APP_NAME
