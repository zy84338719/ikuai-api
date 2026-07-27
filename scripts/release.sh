#!/usr/bin/env bash
# Release helper. Run this on a machine that can reach GitHub
# (the corporate firewall currently blocks 22 + 443 to github.com).
#
# Usage:
#   bash scripts/release.sh
#
# What it does:
#   1. Verifies the working tree is clean.
#   2. Runs the full check (fmt + vet + test).
#   3. Pushes main to origin.
#   4. Pushes the v1.0.0 tag.
#   5. Creates a GitHub release for v1.0.0 using the bundled
#      RELEASE_NOTES_v1.0.0.md as the body.
#
# Requires: gh CLI authenticated against the zy84338719 user, or
# pre-set GITHUB_TOKEN in the environment.

set -euo pipefail

cd "$(dirname "$0")/.."

VERSION="${VERSION:-v1.0.0}"
BRANCH="${BRANCH:-main}"
NOTES="RELEASE_NOTES_${VERSION}.md"

echo "=== pre-flight ==="
if ! git diff --quiet HEAD; then
    echo "working tree is dirty; commit or stash first" >&2
    exit 1
fi
if ! git diff --quiet "$BRANCH" origin/"$BRANCH" 2>/dev/null; then
    echo "local $BRANCH and origin/$BRANCH differ; pull --rebase first" >&2
    exit 1
fi

echo "=== make check ==="
make check

echo "=== push $BRANCH ==="
git push origin "$BRANCH"

echo "=== push tag $VERSION ==="
git push origin "$VERSION"

if [ -f "$NOTES" ] && command -v gh >/dev/null 2>&1; then
    echo "=== create GitHub release ==="
    gh release create "$VERSION" \
        --title "$VERSION — v4-only stable" \
        --notes-file "$NOTES" \
        --target "$BRANCH"
    echo "release published: https://github.com/zy84338719/ikuai-api/releases/tag/$VERSION"
else
    echo "=== skipped gh release (gh not installed or $NOTES missing) ==="
    echo "to publish manually:"
    echo "  gh release create $VERSION --title '$VERSION — v4-only stable' --notes-file $NOTES --target $BRANCH"
fi

echo "=== done ==="
