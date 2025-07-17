#!/bin/bash

# Test the assets CLI

echo "=== Testing Assets CLI ==="

# Test help
echo "1. Testing help command:"
./bin/assets --help

echo -e "\n2. Testing config show (should show workspace ID error):"
./bin/assets config show

echo -e "\n3. Testing create command structure:"
./bin/assets create --help

echo -e "\n4. Testing schema command structure:"
./bin/assets schema --help

echo -e "\n5. Testing with workspace ID (placeholder):"
ATLASSIAN_ASSETS_WORKSPACE_ID=test-workspace-id ./bin/assets config show

echo -e "\n=== CLI Tests Complete ==="