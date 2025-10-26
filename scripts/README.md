# Deployment and Release Scripts

This directory contains scripts for automated deployment and versioning of the `gostratum/core` module.

## Scripts Overview

### deploy.sh
Example deployment script demonstrating ENV_PREFIX usage.

**Usage:**
```bash
./deploy.sh <app-name> <environment> <env-prefix>
```

**Examples:**
```bash
# Deploy with STRATUM prefix (default)
./deploy.sh myapp prod STRATUM

# Deploy with custom prefix
./deploy.sh myapp prod MYAPP

# Deploy to dev environment
./deploy.sh myapp dev MYAPP
```

**Environments:**
- `prod` - Production configuration (port 80, warn logging)
- `staging` - Staging configuration (port 8080, info logging)
- `dev` - Development configuration (port 8080, debug logging)

### release.sh
Automated release process with version bumping and changelog updates.

**Usage:**
```bash
# Patch release (0.1.8 → 0.1.9)
./release.sh patch

# Minor release (0.1.8 → 0.2.0)
./release.sh minor

# Major release (0.1.8 → 1.0.0)
./release.sh major

# Dry run
DRY_RUN=true ./release.sh patch
```

**Process:**
1. Validates version file
2. Checks git status
3. Updates dependencies
4. Runs tests
5. Bumps version
6. Updates CHANGELOG.md
7. Commits and tags

### Other Scripts

- `bump-version.sh` - Increment version number
- `update-changelog.sh` - Update CHANGELOG.md
- `update-deps.sh` - Update gostratum dependencies
- `validate-version.sh` - Validate .version file format

## ENV_PREFIX Feature (v0.1.9+)

The `ENV_PREFIX` environment variable allows you to set a global environment variable prefix without code changes.

### Precedence

1. **WithEnvPrefix() option** (code) - Highest priority
2. **ENV_PREFIX environment variable** - Global default
3. **STRATUM** - Hardcoded default

### Usage Examples

#### Option 1: Via Environment Variable (Recommended for Deployment)

```bash
#!/bin/bash
# deploy-prod.sh

export ENV_PREFIX=MYAPP
export MYAPP_APP_PORT=8080
export MYAPP_DB_HOST=prod-db.example.com
export MYAPP_DB_PORT=5432

./myapp
```

#### Option 2: Via Code (Recommended for Development)

```go
package main

import "github.com/gostratum/core/configx"

func main() {
    loader := configx.New(
        configx.WithEnvPrefix("MYAPP"),
    )
    // Uses MYAPP_* environment variables
}
```

#### Option 3: Default Behavior

```bash
# No ENV_PREFIX set, no WithEnvPrefix() option
export STRATUM_APP_PORT=8080
./myapp
# Uses STRATUM_* environment variables
```

## Deployment Workflow

### 1. Development

```bash
# Use default STRATUM prefix
export STRATUM_APP_PORT=8080
export STRATUM_DB_HOST=localhost
go run .
```

### 2. Staging/Production Deployment

```bash
# Set global prefix via ENV_PREFIX
export ENV_PREFIX=MYAPP
export APP_ENV=prod
export MYAPP_APP_PORT=80
export MYAPP_DB_HOST=prod-db.example.com

# Or use deploy.sh script
./scripts/deploy.sh myapp prod MYAPP
```

### 3. Docker Deployment

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o myapp

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/myapp .

# Set environment prefix
ENV ENV_PREFIX=MYAPP
ENV APP_ENV=prod
ENV MYAPP_APP_PORT=8080

CMD ["./myapp"]
```

```bash
# Run with overrides
docker run -e ENV_PREFIX=MYAPP -e MYAPP_APP_PORT=9000 myapp
```

### 4. Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        env:
        - name: ENV_PREFIX
          value: "MYAPP"
        - name: APP_ENV
          value: "prod"
        - name: MYAPP_APP_PORT
          valueFrom:
            configMapKeyRef:
              name: myapp-config
              key: port
        - name: MYAPP_DB_HOST
          valueFrom:
            secretKeyRef:
              name: myapp-secrets
              key: db_host
```

## Release Workflow

### Manual Release

```bash
# 1. Update code and tests
# 2. Update CHANGELOG.md manually
# 3. Update .version file
echo "0.1.9" > .version

# 4. Run release script
./scripts/release.sh

# 5. Push to remote
git push origin main
git push origin v0.1.9
```

### Automated Release

```bash
# Uses scripts to automate everything
./scripts/release.sh patch

# Follow prompts and review changes
git show HEAD
git push origin main
git push origin v0.1.9
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Run tests
        run: make test
      
      - name: Extract version
        id: version
        run: echo "VERSION=$(cat .version)" >> $GITHUB_OUTPUT
      
      - name: Create Release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: v${{ steps.version.outputs.VERSION }}
          release_name: Release v${{ steps.version.outputs.VERSION }}
          body_path: CHANGELOG.md
```

## Best Practices

1. **Always run tests before releasing**
   ```bash
   make test
   ```

2. **Use semantic versioning**
   - MAJOR: Breaking changes
   - MINOR: New features (backward compatible)
   - PATCH: Bug fixes

3. **Update CHANGELOG.md before releasing**
   - Follow [Keep a Changelog](https://keepachangelog.com/) format
   - Document all changes clearly

4. **Use ENV_PREFIX for deployment flexibility**
   - Set globally via environment variable
   - Override in code only when necessary

5. **Tag releases properly**
   ```bash
   git tag -a v0.1.9 -m "Release v0.1.9"
   git push origin v0.1.9
   ```

## Troubleshooting

### Version mismatch
```bash
# Validate version file
./scripts/validate-version.sh

# Check CHANGELOG.md has matching version
grep "## \[$(cat .version)\]" CHANGELOG.md
```

### Tests fail
```bash
# Run tests with verbose output
go test -v ./...

# Check specific package
go test -v ./configx/...
```

### ENV_PREFIX not working
```bash
# Verify environment variable is set
echo $ENV_PREFIX

# Check if explicitly bound (required for some cases)
loader.BindEnv("app.port")

# Debug: Check what Viper sees
loader.BindEnv("app.port")
config := &AppConfig{}
loader.Bind(config)
```

## Support

For issues or questions:
- GitHub Issues: https://github.com/gostratum/core/issues
- Documentation: https://github.com/gostratum/core/blob/main/README.md
- CHANGELOG: https://github.com/gostratum/core/blob/main/CHANGELOG.md
