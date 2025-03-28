#!/bin/bash

# Wait for the app to be ready
echo "Waiting for app to be ready..."
while ! curl -s http://app:8080/v1/health > /dev/null; do
    sleep 1
done
echo "App is ready!"

# Build the Newman Docker image
docker build -t alchemorsel-e2e-tests -f Dockerfile.newman .

# Run the Newman container
docker run --rm \
  --network alchemorsel-v1_appnet \
  -e "base_url=http://app:8080" \
  -e "JWT_SECRET=your-super-secret-key-that-is-at-least-32-characters-long" \
  alchemorsel-e2e-tests 