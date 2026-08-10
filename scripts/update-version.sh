#!/usr/bin/env bash

set -euo pipefail

#############################################
# Airtel Cloud Terraform Provider
# Version Update Utility
#
# Usage:
#   ./scripts/update-version.sh 1.1.5
#############################################

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <major.minor.patch>"
    exit 1
fi

VERSION="$1"

if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "ERROR: Invalid version '$VERSION'"
    echo "Expected format: X.Y.Z (example: 1.1.5)"
    exit 1
fi

echo ""
echo "==============================================="
echo " Airtel Cloud Terraform Provider"
echo " Version Update Utility"
echo "==============================================="
echo " Target Version : ${VERSION}"
echo ""

#############################################
# Update Terraform example files
#############################################

echo "Updating Terraform examples..."

find examples -type f -name "*.tf" | while read -r file; do
    sed -i.bak -E \
    "s|version *= *\"(~> )?[0-9]+\.[0-9]+\.[0-9]+\"|version = \"${VERSION}\"|g" \
    "$file"

    rm -f "${file}.bak"
    echo "  ✓ $file"
done

#############################################
# Update Markdown documentation
#############################################

echo ""
echo "Updating Markdown documentation..."

find . -type f -name "*.md" | while read -r file; do

    # Provider download path
    sed -i.bak -E \
    "s|terraform-provider-airtelcloud/[0-9]+\.[0-9]+\.[0-9]+|terraform-provider-airtelcloud/${VERSION}|g" \
    "$file"

    # Registry path
    sed -i.bak -E \
    "s|airtelcloud/[0-9]+\.[0-9]+\.[0-9]+/|airtelcloud/${VERSION}/|g" \
    "$file"

    # Terraform provider version blocks
    sed -i.bak -E \
    "s|version *= *\"(~> )?[0-9]+\.[0-9]+\.[0-9]+\"|version = \"${VERSION}\"|g" \
    "$file"

    # Provider version line
    sed -i.bak -E \
    "s|Provider version\*\*: [0-9]+\.[0-9]+\.[0-9]+|Provider version**: ${VERSION}|g" \
    "$file"

    rm -f "${file}.bak"

    echo "  ✓ $file"

done

echo ""
echo "==============================================="
echo " Version update completed successfully."
echo "==============================================="
echo ""

echo "Modified files:"
git diff --name-only
