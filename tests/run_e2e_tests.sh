#!/bin/bash

# Wait for the app to be ready
echo "Waiting for app to be ready..."
while ! curl -s http://localhost:8080/v1/health > /dev/null; do
    echo "Waiting for app to be ready..."
    sleep 1
done
echo "App is ready!"

# Build the Newman Docker image
echo "Building Newman Docker image..."
cd tests && docker build -t alchemorsel-e2e-tests -f Dockerfile.newman . && cd ..

# Run the Newman container
echo "Running Newman tests..."
docker run --rm \
  --network alchemorsel-v1_appnet \
  -e "base_url=http://app:8080" \
  -e "JWT_SECRET=your-super-secret-key-that-is-at-least-32-characters-long" \
  alchemorsel-e2e-tests 