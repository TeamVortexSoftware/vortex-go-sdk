# Publishing the Vortex Go SDK

This guide walks you through publishing the Vortex Go SDK so users can install it with `go get`.

## Overview

Go modules are published via Git tags. There's no central registry like npm or PyPI - Go fetches packages directly from version control (GitHub).

## Prerequisites

1. **GitHub Repository**: Ensure the SDK is in a public GitHub repository
2. **Git Access**: Push access to the repository
3. **Go Module Path**: The module path in `go.mod` must match the repository structure

## Current Configuration

The SDK is configured as:
```go
module github.com/teamvortexsoftware/vortex-go-sdk
```

This means:
- Repository: `https://github.com/teamvortexsoftware/vortex-go-sdk`
- Import path: `github.com/teamvortexsoftware/vortex-go-sdk`

## Publishing Process

### Option 1: Standalone Repository (Recommended for Public SDKs)

For independent versioning and cleaner distribution:

1. **Create a separate repository** for the Go SDK:
   ```bash
   # Create new repository at github.com/teamvortexsoftware/vortex-go-sdk
   # Then copy the SDK files
   cp -r packages/vortex-go-sdk/* /path/to/vortex-go-sdk/
   ```

2. **Initialize Git and push**:
   ```bash
   cd /path/to/vortex-go-sdk
   git init
   git add .
   git commit -m "Initial commit"
   git remote add origin https://github.com/teamvortexsoftware/vortex-go-sdk.git
   git push -u origin main
   ```

3. **Create a version tag**:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

4. **Users can now install**:
   ```bash
   go get github.com/teamvortexsoftware/vortex-go-sdk@v1.0.0
   ```

### Option 2: Monorepo with Submodules

To keep the SDK in your existing monorepo:

1. **Update the module path** in `go.mod`:
   ```go
   module github.com/teamvortexsoftware/vortex/packages/vortex-go-sdk
   ```

2. **Create a version tag with path prefix**:
   ```bash
   git tag packages/vortex-go-sdk/v1.0.0
   git push origin packages/vortex-go-sdk/v1.0.0
   ```

3. **Users install with the full path**:
   ```bash
   go get github.com/teamvortexsoftware/vortex/packages/vortex-go-sdk@v1.0.0
   ```

**Note**: Monorepo approach is more complex. Standalone is recommended for public SDKs.

## Version Management

Go uses Semantic Versioning with the `v` prefix:

- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.0.1` - Patch release (bug fixes)
- `v2.0.0` - Breaking changes (requires new import path for major versions ≥ 2)

### Major Version 2+

For major versions ≥ 2, Go requires updating the module path:

```go
// go.mod for v2
module github.com/teamvortexsoftware/vortex-go-sdk/v2
```

Users import as:
```go
import "github.com/teamvortexsoftware/vortex-go-sdk/v2"
```

## Release Checklist

### 1. Update Version Information

Update `README.md` with the new version in installation examples.

### 2. Update CHANGELOG

Document changes in a `CHANGELOG.md` file.

### 3. Run Tests

```bash
cd packages/vortex-go-sdk
go test ./...
```

### 4. Commit Changes

```bash
git add .
git commit -m "Release v1.0.0"
git push
```

### 5. Create and Push Tag

```bash
# For standalone repository
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# For monorepo
git tag -a packages/vortex-go-sdk/v1.0.0 -m "Release v1.0.0"
git push origin packages/vortex-go-sdk/v1.0.0
```

### 6. Verify on pkg.go.dev

After pushing the tag, the package should appear on [pkg.go.dev](https://pkg.go.dev) within a few minutes:
- Standalone: `https://pkg.go.dev/github.com/teamvortexsoftware/vortex-go-sdk`
- Monorepo: `https://pkg.go.dev/github.com/teamvortexsoftware/vortex/packages/vortex-go-sdk`

If it doesn't appear, request indexing at: `https://pkg.go.dev/github.com/teamvortexsoftware/vortex-go-sdk@v1.0.0`

## Automated Publishing with GitHub Actions

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'
      - 'packages/vortex-go-sdk/v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.18'

      - name: Run tests
        run: |
          cd packages/vortex-go-sdk
          go test ./...

      - name: Create GitHub Release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref }}
          release_name: Release ${{ github.ref }}
          draft: false
          prerelease: false
```

## Testing Installation

After publishing, test the installation:

```bash
# Create a test directory
mkdir test-vortex-go-sdk
cd test-vortex-go-sdk

# Initialize a Go module
go mod init test

# Install the SDK
go get github.com/teamvortexsoftware/vortex-go-sdk@v1.0.0

# Create a test file
cat > main.go << 'EOF'
package main

import (
    "fmt"
    "github.com/teamvortexsoftware/vortex-go-sdk"
)

func main() {
    client := vortex.NewClient("test-key")
    fmt.Println("SDK loaded successfully!")
}
EOF

# Run it
go run main.go
```

## Versioning Best Practices

1. **Use annotated tags**: `git tag -a v1.0.0 -m "message"` (not lightweight tags)
2. **Never delete published tags**: This breaks existing dependencies
3. **Follow semantic versioning**: Major.Minor.Patch
4. **Document breaking changes**: Especially for major version bumps
5. **Test before tagging**: Ensure tests pass

## Module Proxy and Checksums

Go uses:
- **Module Proxy** (proxy.golang.org): Caches modules for availability
- **Checksum Database** (sum.golang.org): Ensures integrity

After publishing:
1. First download is fetched from source (GitHub)
2. Module is cached in the proxy
3. Checksum is recorded in the database
4. Future downloads use the cached version

## Troubleshooting

### Package not appearing on pkg.go.dev

- Wait 10-15 minutes for indexing
- Verify the tag is visible on GitHub
- Ensure `go.mod` is in the repository root (or appropriate subdirectory for monorepos)
- Check that the module path in `go.mod` matches the repository structure

### Users getting "module not found"

- Verify the tag exists: `git tag -l`
- Check the module path matches the import path
- Ensure the repository is public
- Try requesting indexing manually on pkg.go.dev

### Import path issues

- The import path must exactly match the module path in `go.mod`
- For v2+, must include `/v2` suffix

## Resources

- [Go Modules Reference](https://go.dev/ref/mod)
- [Publishing Go Modules](https://go.dev/doc/modules/publishing)
- [Semantic Versioning](https://semver.org/)
- [pkg.go.dev](https://pkg.go.dev)
- [Go Module Proxy](https://proxy.golang.org/)

## Recommended Setup

For this SDK, I recommend:

1. **Create a standalone repository**: `github.com/teamvortexsoftware/vortex-go-sdk`
2. **Keep module path simple**: `github.com/teamvortexsoftware/vortex-go-sdk`
3. **Use standard versioning**: `v1.0.0`, `v1.1.0`, etc.
4. **Set up GitHub Actions**: Automate testing and releases
5. **Document in README**: Clear installation and usage instructions

This provides the best experience for Go developers and follows Go ecosystem conventions.
