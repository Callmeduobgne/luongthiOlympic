#!/bin/bash

# Copyright 2025 IBN Network (ICTU Blockchain Network)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -e

echo "Running database migrations..."

# Iterate through all .up.sql files in the /migrations directory
for file in /migrations/*.up.sql; do
    if [ -f "$file" ]; then
        echo "Applying migration: $file"
        psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$file"
    fi
done

echo "Database migrations completed successfully."
