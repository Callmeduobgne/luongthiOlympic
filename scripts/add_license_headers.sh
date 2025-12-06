#!/bin/bash

# Define the license header for C-style comments (Go, JS, TS)
read -r -d '' C_STYLE_HEADER << EOM
/*
Copyright (c) 2025 IBN Network

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
EOM

# Define the license header for Hash-style comments (Shell, Dockerfile)
read -r -d '' HASH_STYLE_HEADER << EOM
# Copyright (c) 2025 IBN Network
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
EOM

# Function to add header
add_header() {
    local file="$1"
    local type="$2"
    
    if [ ! -f "$file" ]; then
        return
    fi

    # Check if header already exists (look for "Copyright (c) 2025 IBN Network")
    if grep -q "Copyright (c) 2025 IBN Network" "$file"; then
        echo "Skipping $file (Header already present)"
        return
    fi

    echo "Adding header to $file"
    
    local temp_file=$(mktemp)
    
    if [ "$type" == "c-style" ]; then
        echo "$C_STYLE_HEADER" > "$temp_file"
        echo "" >> "$temp_file"
        cat "$file" >> "$temp_file"
    elif [ "$type" == "hash-style" ]; then
        # Check for shebang
        if head -n 1 "$file" | grep -q "^#!"; then
            head -n 1 "$file" > "$temp_file"
            echo "" >> "$temp_file"
            echo "$HASH_STYLE_HEADER" >> "$temp_file"
            echo "" >> "$temp_file"
            tail -n +2 "$file" >> "$temp_file"
        else
            echo "$HASH_STYLE_HEADER" > "$temp_file"
            echo "" >> "$temp_file"
            cat "$file" >> "$temp_file"
        fi
    fi
    
    mv "$temp_file" "$file"
}

# Export function and variables for find -exec (not strictly needed with loop)

# Process Go files
echo "Processing Go files..."
find backend api-gateway -name "*.go" | while read file; do add_header "$file" "c-style"; done

# Process TS/JS files
echo "Processing TS/JS files..."
find frontend/src teaTraceCC/src -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" | while read file; do add_header "$file" "c-style"; done

# Process Shell scripts
echo "Processing Shell scripts..."
find scripts -name "*.sh" | while read file; do add_header "$file" "hash-style"; done

echo "Done."
