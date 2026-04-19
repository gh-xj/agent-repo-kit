#!/bin/sh
# Usage: scripts/tag-release.sh <version>
# Example: scripts/tag-release.sh 0.4.0
set -eu
VERSION="${1:?usage: tag-release.sh <version>}"
git tag -a "v${VERSION}" -m "release v${VERSION}"
git push origin "v${VERSION}"
printf '[tag-release] pushed v%s\n' "$VERSION"
