#!/bin/bash

# Find all folders containing .proto files
folders=$(find . -type f -name "*.proto" -exec dirname {} \; | sort -u)

# Loop through each unique folder
for folder in $folders; do
    echo "Processing .proto files in folder: $folder"
    for proto_file in "$folder"/*.proto; do
        # Check if there are any .proto files in the folder
        if [[ -e "$proto_file" ]]; then
            echo "Compiling $proto_file..."
            # Run protoc command and check for errors
            if ! protoc --go_out=. --go_opt=paths=source_relative \
                         --go-grpc_out=. --go-grpc_opt=paths=source_relative \
                         "$proto_file"; then
                echo "Error: Failed to compile $proto_file" >&2
            else
                echo "Successfully compiled $proto_file"
            fi
        else
            echo "No .proto files found in $folder"
        fi
    done
done
