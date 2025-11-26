#!/bin/bash

# Script to add GetPackage route to backend router
# This script adds the route registration if it doesn't already exist

set -e

ROUTER_FILE="/home/exp2/ibn/backend/cmd/server/main.go"

echo "🔍 Checking if GetPackage route already exists..."

if grep -q "GetPackage" "$ROUTER_FILE" 2>/dev/null; then
    echo "✅ GetPackage route already registered!"
    exit 0
fi

echo "📝 Adding GetPackage route to router..."

# Find the line with CreatePackage route and add GetPackage after it
# Using sed to insert the new route
sed -i '/teatrace.*packages.*CreatePackage/a \        r.Get("/api/v1/teatrace/packages/{packageId}", teaTraceHandler.GetPackage)' "$ROUTER_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Route added successfully!"
    echo ""
    echo "Added line:"
    echo '        r.Get("/api/v1/teatrace/packages/{packageId}", teaTraceHandler.GetPackage)'
    echo ""
    echo "📋 Next step: Restart backend with:"
    echo "   docker-compose restart ibn-backend"
else
    echo "❌ Failed to add route automatically"
    echo ""
    echo "Please add this line manually to $ROUTER_FILE:"
    echo '        r.Get("/api/v1/teatrace/packages/{packageId}", teaTraceHandler.GetPackage)'
    echo ""
    echo "Add it after the CreatePackage route line."
    exit 1
fi
