#!/bin/bash

# Run Go tests and capture the output
TEST_OUTPUT=$(go test -cover ./...)

# Extract the coverage percentage from the output
COVERAGE=$(echo "$TEST_OUTPUT" | grep -oP 'coverage: \K[0-9.]+(?=%)')

# Determine the badge color based on the coverage percentage
if (( $(echo "$COVERAGE < 50" | bc -l) )); then
    COLOR="red"
elif (( $(echo "$COVERAGE < 80" | bc -l) )); then
    COLOR="yellow"
else
    COLOR="green"
fi

# Generate the badge URL using shields.io
BADGE_URL="https://img.shields.io/badge/coverage-${COVERAGE}%25-${COLOR}.svg"

# Write the badge markdown to README.md
echo -e "\n![Coverage Badge](${BADGE_URL})" >> README.md
