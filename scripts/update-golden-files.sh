#!/bin/bash
# This script updates the golden files for unit tests that import the 'gotest.tools/v3/golden' dependency in a Go project.
# It lists all packages in the project, checks for the dependency in test imports, and runs unit tests with '-update' to update golden files.

dependency="gotest.tools/v3/golden"

# List all packages in the project
packages=$(go list ./...)

# Loop through each package and check if it imports the specific dependency
for pkg in $packages; do
    # Use 'go list' with 'XTestImports' template to get the imports from test binaries
    imports=$(go list -f '{{join .TestImports "\n"}}{{"\n"}}{{join .XTestImports "\n"}}' "$pkg")
    
    # Check if the dependency is in the imports
    if echo "$imports" | grep -q "$dependency"; then
        # If the dependency is found, run the unit tests updating the golden files
        go test "$pkg" -update -timeout 30s
    fi
done
