#!/bin/bash

# Change to project root directory
cd "$(dirname "$0")/.." || exit

# Simple approach: just run the app with go run
# Use fswatch if available, otherwise just run once
if command -v fswatch &> /dev/null; then
    fswatch -o -r --exclude '\.git' --exclude 'tmp' . | while read; do
        clear
        go run .
    done
else
    # Just run the app
    echo "For hot-reload, install fswatch: brew install fswatch"
    echo "Running app once..."
    go run .
fi

