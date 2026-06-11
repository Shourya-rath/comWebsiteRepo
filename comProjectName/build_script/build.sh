#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Step 1: Generating Templ components..."
templ generate

echo "Step 2: Building Go binary..."
go build -o project_store_server

echo "Step 3: Starting server..."
echo "---------------------------"
./project_store_server