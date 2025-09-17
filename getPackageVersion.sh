#!/bin/bash

# Check if package.json exists
if [ ! -f "package.json" ]; then
    echo "Error: package.json not found in current directory" >&2
    exit 1
fi

# Extract version using only native Unix tools (grep, sed, awk)
# This handles various formatting styles in package.json
grep '"version"' package.json | head -1 | sed 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/'
