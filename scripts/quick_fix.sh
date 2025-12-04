#!/bin/bash

# Quick fix for setup.sh - remove lines 1236-1770

echo "Backing up setup.sh..."
cp setup.sh setup.sh.broken_backup

echo "Removing duplicate code (lines 1236-1770)..."
sed -i '1236,1770d' setup.sh

echo "Checking syntax..."
if bash -n setup.sh 2>/dev/null; then
    echo "✓ Syntax OK!"
    echo "✓ File fixed successfully"
    echo ""
    echo "You can now run: ./setup.sh"
else
    echo "✗ Still has syntax errors"
    echo "Showing errors:"
    bash -n setup.sh
fi
