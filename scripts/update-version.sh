#!/bin/bash

# Script to update version across all files
# Usage: ./scripts/update-version.sh 1.2.1

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.2.1"
    exit 1
fi

NEW_VERSION="$1"

# Validate version format (semantic versioning)
if ! echo "$NEW_VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    echo "Error: Version must be in format X.Y.Z (e.g., 1.2.1)"
    exit 1
fi

echo "Updating version to $NEW_VERSION..."

# Update version.go
echo "Updating version.go..."
sed -i.bak "s/VERSION  = \".*\"/VERSION  = \"$NEW_VERSION\"/" version.go

# Update Makefile
echo "Updating Makefile..."
sed -i.bak "s/VERSION=.*/VERSION=$NEW_VERSION/" Makefile

# Update debian/changelog
echo "Updating debian/changelog..."
MAINTAINER=$(git config user.name)" <"$(git config user.email)">"
TIMESTAMP=$(date -R)

# Create new changelog entry
cat > debian/changelog.new << EOF
mlogtail ($NEW_VERSION-1) unstable; urgency=medium

  * Release version $NEW_VERSION

 -- $MAINTAINER  $TIMESTAMP

EOF

# Append old changelog
cat debian/changelog >> debian/changelog.new
mv debian/changelog.new debian/changelog

# Clean up backup files
rm -f version.go.bak Makefile.bak

echo "Version updated successfully to $NEW_VERSION"
echo ""
echo "Files updated:"
echo "  - version.go"
echo "  - Makefile"
echo "  - debian/changelog"
echo ""
echo "Next steps:"
echo "  1. Review changes: git diff"
echo "  2. Commit changes: git add . && git commit -m 'Bump version to $NEW_VERSION'"
echo "  3. Create tag: git tag v$NEW_VERSION"
echo "  4. Push: git push origin master && git push origin v$NEW_VERSION"