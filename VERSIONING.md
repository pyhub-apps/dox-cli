# HeadVer Versioning System

This project uses the **HeadVer** versioning system, a modern alternative to Semantic Versioning designed for agile product teams.

## Version Format

```
{head}.{yearweek}.{build}
```

- **`{head}`**: Major version number (manually controlled)
- **`{yearweek}`**: 2-digit year + 2-digit week number (automatically generated)
- **`{build}`**: Build number within the week (automatically generated)

### Examples
- `1.2534.0` - Version 1, created in week 34 of 2025, first build
- `1.2534.23` - Version 1, created in week 34 of 2025, 23rd build
- `2.2601.5` - Version 2, created in week 1 of 2026, 5th build

## Why HeadVer?

### Advantages over SemVer
1. **Simplicity**: Only one number (`head`) needs manual management
2. **Time Awareness**: The `yearweek` component shows the age of a version at a glance
3. **Unique Builds**: Every release gets a unique version automatically
4. **No Forgotten Bumps**: Version updates happen automatically with each build

### When to Change the Head Version
Increment the `{head}` version only for:
- **Breaking Changes**: Incompatible API changes
- **Major Rewrites**: Significant architectural changes
- **Product Pivots**: Major shifts in product direction

## Using the HeadVer Script

### Basic Usage
```bash
# Generate current version
./scripts/headver.sh
# Output: 1.2534.23

# Get specific components
./scripts/headver.sh --head      # Output: 1
./scripts/headver.sh --yearweek  # Output: 2534
./scripts/headver.sh --build     # Output: 23
```

### Managing Head Version
```bash
# Set head version
./scripts/headver.sh --set-head 2

# Increment head version
./scripts/headver.sh --bump-head
```

### CI/CD Integration
The build number can be overridden using environment variables:
```bash
# GitHub Actions
BUILD_NUMBER=$GITHUB_RUN_NUMBER ./scripts/headver.sh

# Jenkins
BUILD_NUMBER=$BUILD_NUMBER ./scripts/headver.sh
```

## Building with HeadVer

### Local Build
```bash
# Version is automatically generated
make build
```

### Manual Build with Version
```bash
VERSION=$(./scripts/headver.sh)
go build -ldflags="-X main.version=$VERSION" -o pyhub-docs
```

## Creating Releases

### GitHub Release
```bash
# Generate version
VERSION=$(./scripts/headver.sh)

# Create and push tag
git tag "v$VERSION" -m "Release $VERSION"
git push origin "v$VERSION"
```

The GitHub Actions workflow will automatically:
1. Use the HeadVer script to generate consistent versions
2. Build binaries with the correct version embedded
3. Create a GitHub release

## Migration from SemVer

We migrated from Semantic Versioning (v1.3.0) to HeadVer in August 2025:
- Last SemVer version: v1.3.0
- First HeadVer version: 1.2534.0

The `{head}` version started at 1 to maintain continuity with the major version from SemVer.

## Version History

| Version Type | Format | Example | Use Case |
|-------------|---------|---------|-----------|
| SemVer (old) | MAJOR.MINOR.PATCH | 1.3.0 | Used until August 2025 |
| HeadVer (current) | HEAD.YEARWEEK.BUILD | 1.2534.23 | Current versioning system |

## References

- [HeadVer Official Repository](https://github.com/line/headver)
- [HeadVer Blog Post (Korean)](https://techblog.lycorp.co.jp/ko/headver-new-versioning-system-for-product-teams)
- [HeadVer Blog Post (Japanese)](https://techblog.lycorp.co.jp/ja/headver-new-versioning-system-for-product-teams)

## FAQ

**Q: When should I bump the head version?**
A: Only for breaking changes or major product shifts. Most releases will only increment the build number.

**Q: What if I release multiple times in the same week?**
A: Each release gets a unique build number. For example: 1.2534.0, 1.2534.1, 1.2534.2

**Q: How do I know when a version was released?**
A: The yearweek component tells you. For example, 2534 means week 34 of 2025.

**Q: Can I use this with Docker tags?**
A: Yes! HeadVer versions work well as Docker tags: `pyhub-docs:1.2534.23`