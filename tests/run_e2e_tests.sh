#!/bin/bash

# Build the Newman Docker image
docker build -t alchemorsel-e2e-tests -f Dockerfile.newman .

# Run the Newman container
docker run --rm \
  --network alchemorsel-v1_appnet \
  -e "base_url=http://app:8080" \
  alchemorsel-e2e-tests 