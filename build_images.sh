#!/bin/bash

set -e

# --- Configuration ---
USER="deskchen"
TAG="latest"
IMAGE="onlineboutique-grpc"
YAML_DIR="kubernetes/apply"
# ---

# 1. Update Kubernetes YAML files to use the new image
echo "Updating YAML files in $YAML_DIR..."
NEW_IMAGE="$USER/$IMAGE:$TAG"
# Find and replace the image line in all .yaml files in the directory
sed -i.bak "s|image: .*/$IMAGE:.*|image: $NEW_IMAGE|g" $YAML_DIR/*.yaml
# Remove the backup files created by sed
rm -f $YAML_DIR/*.yaml.bak

# 2. Build the Docker image
echo "Building Docker image: $NEW_IMAGE"
docker build -t "$NEW_IMAGE" -f Dockerfile .

# 3. Push the Docker image
echo "Pushing Docker image: $NEW_IMAGE"
docker push "$NEW_IMAGE"

echo "✅ Process complete."