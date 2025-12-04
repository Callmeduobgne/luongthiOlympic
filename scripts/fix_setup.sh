#!/bin/bash

# Script to fix setup.sh by removing duplicate code

echo "Fixing setup.sh..."

# Backup original file
cp scripts/setup.sh scripts/setup.sh.backup

# Use sed to delete lines 1234 to 1770 (duplicate code)
sed -i '1234,1770d' scripts/setup.sh

echo "Done! Backup saved to scripts/setup.sh.backup"
echo "Verifying syntax..."
bash -n scripts/setup.sh && echo "✓ Syntax OK" || echo "✗ Syntax error - restoring backup"
