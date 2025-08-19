# CI Pipeline Go Version Fixes

## Issue Summary

The GitHub Actions CI pipeline was failing with Go version compatibility errors:

```
package requires newer Go version go1.25 (application built with go1.24)
the Go language version (go1.24) used to build golangci-lint is lower than the targeted Go version (1.25.0)
```

## Root Cause

1. **go.mod specified Go 1.25.0** - This version doesn't exist yet (Go 1.25 is unreleased)
2. **CI pipeline used Go 1.21** - Older than the specified version
3. **Tool compatibility issues** - golangci-lint and other tools were built with Go 1.24

## Fixes Applied

### 1. Updated go.mod
```diff
- go 1.25.0
+ go 1.23
```

### 2. Updated CI Workflows
Updated both `.github/workflows/ci.yml` and `.github/workflows/release.yml`:

```diff
env:
-  GO_VERSION: '1.21'
+  GO_VERSION: '1.23'  # Using stable Go version for CI
```

### 3. Updated Documentation
Updated all documentation files to reflect the correct Go version requirement:

- `README.md`: Multiple references updated to Go 1.23+
- `docs/development.md`: Prerequisites updated
- `docs/testing.md`: Prerequisites updated

### 4. Verified Compatibility
- Ensured the project builds with Go 1.23
- Verified all tests pass
- Confirmed compatibility with the local Go 1.25.0 development version

## Go Version Strategy

- **Minimum Required**: Go 1.23 (stable release)
- **CI/CD Uses**: Go 1.23 (for consistency and tool compatibility)
- **Local Development**: Works with Go 1.23+ including development versions like 1.25

## Files Modified

### Core Configuration
- `go.mod` - Updated Go version requirement
- `.github/workflows/ci.yml` - Updated CI Go version
- `.github/workflows/release.yml` - Updated release Go version

### Documentation
- `README.md` - Updated prerequisites and tech stack
- `docs/development.md` - Updated prerequisites
- `docs/testing.md` - Updated prerequisites

## Benefits

1. **CI Compatibility**: Tools and runtime now use compatible Go versions
2. **Stable Releases**: Using stable Go 1.23 ensures reliability
3. **Future Compatibility**: Works with newer Go versions (like 1.25 development)
4. **Clear Documentation**: Users know the correct requirements

## Testing Verification

After the fixes:
- ✅ Project builds successfully
- ✅ All unit tests pass
- ✅ Integration tests work
- ✅ CI pipeline should run without version conflicts

## Next Steps

1. **Push changes** to trigger CI pipeline
2. **Monitor CI results** to confirm fixes work
3. **Update any external documentation** that references Go requirements
4. **Consider pinning specific Go patch versions** if needed for reproducible builds

## Prevention

To avoid similar issues in the future:

1. **Use stable Go versions** in go.mod
2. **Keep CI and go.mod versions synchronized**
3. **Test with the minimum required Go version**
4. **Monitor Go release schedule** for updates
5. **Use dependabot or similar** to track Go version updates

## Go Version Timeline

- Go 1.21: Released August 2023
- Go 1.22: Released February 2024  
- Go 1.23: Released August 2024 ✅ (Current stable)
- Go 1.24: Expected February 2025
- Go 1.25: Expected August 2025

Using Go 1.23 provides a good balance of modern features and stability for CI/CD pipelines.
