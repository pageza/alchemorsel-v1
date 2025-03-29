#!/bin/bash

# Create secrets directory if it doesn't exist
mkdir -p secrets

# Set up JWT secret (64 characters)
echo "this-is-a-very-long-and-secure-jwt-secret-key-for-testing-purposes-only-1234" > secrets/jwt_secret.txt

# Wait for the app to be ready
echo "Waiting for app to be ready..."
while ! curl -s http://localhost:8080/v1/health > /dev/null; do
    echo "Waiting for app to be ready..."
    sleep 1
done
echo "App is ready!"

# Build the Newman Docker image
echo "Building Newman Docker image..."
cd tests && docker build -t alchemorsel-e2e-tests -f e2e/Dockerfile.newman . && cd ..

# Run the Newman container
echo "Running Newman tests..."
docker run --rm \
  --network alchemorsel-v1_appnet \
  -e "base_url=http://app:8080" \
  -e "JWT_SECRET=this-is-a-very-long-and-secure-jwt-secret-key-for-testing-purposes-only-1234" \
  -e "INTEGRATION_TEST=true" \
  alchemorsel-e2e-tests 