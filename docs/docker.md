# Docker Configuration Documentation

## Overview

This document provides comprehensive documentation for the Docker configuration in the Alchemorsel project, including security scanning, secrets management, and environment profiles.

## Table of Contents
1. [Security Scanning](#security-scanning)
2. [Secrets Management](#secrets-management)
3. [Environment Profiles](#environment-profiles)
4. [Getting Started](#getting-started)

## Security Scanning

The project implements multi-layer security scanning using Trivy and Snyk:

### Trivy Scanning
- Scans for vulnerabilities in base images and dependencies
- Configured to fail on high and critical severity issues
- Runs during the build process
- Command: `trivy fs --exit-code 1 --severity HIGH,CRITICAL .`

### Snyk Scanning
- Provides additional security scanning capabilities
- Requires SNYK_TOKEN environment variable
- Scans for vulnerabilities in dependencies
- Command: `snyk test --severity-threshold=high`

### Running Security Scans
```bash
# Run security scanning service
docker-compose --profile security up security-scan
```

## Secrets Management

The project uses Docker secrets for secure credential management:

### Available Secrets
1. `postgres_user`: PostgreSQL database username
2. `postgres_password`: PostgreSQL database password
3. `jwt_secret`: Secret key for JWT token generation

### Secret Files
- Location: `secrets/` directory
- Example files: `*.txt.example`
- Actual secret files: `*.txt` (gitignored)

### Setting Up Secrets
1. Copy example files:
```bash
cp secrets/*.txt.example secrets/*.txt
```

2. Update secret values in the actual files:
```bash
# secrets/postgres_user.txt
your_actual_db_user

# secrets/postgres_password.txt
your_actual_db_password

# secrets/jwt_secret.txt
your_actual_jwt_secret
```

### Accessing Secrets in Containers
Secrets are mounted as files in `/run/secrets/` and accessed via environment variables:
- `POSTGRES_USER_FILE`
- `POSTGRES_PASSWORD_FILE`
- `JWT_SECRET_FILE`

## Environment Profiles

The project uses Docker Compose profiles to manage different environments:

### Available Profiles

1. **Development** (`development`)
   - For local development
   - Includes hot-reload
   - Exposes development ports
   - Services:
     - app
     - postgres
     - redis
     - prometheus
     - grafana

2. **Production** (`production`)
   - For production deployment
   - Optimized for performance
   - Includes monitoring
   - Services:
     - app
     - postgres
     - redis
     - prometheus
     - grafana
     - nginx

3. **Testing** (`testing`)
   - For running tests
   - Includes test databases
   - Services:
     - app
     - postgres
     - redis

4. **Security** (`security`)
   - For security scanning
   - Services:
     - security-scan

### Running Different Profiles

```bash
# Development
docker-compose --profile development up

# Production
docker-compose --profile production up

# Testing
docker-compose --profile testing up

# Security scanning
docker-compose --profile security up security-scan
```

## Getting Started

### Prerequisites
- Docker
- Docker Compose
- Snyk CLI (for security scanning)

### Initial Setup

1. Set up secrets:
```bash
# Create secrets directory if it doesn't exist
mkdir -p secrets

# Copy example files
cp secrets/*.txt.example secrets/*.txt

# Update secret values
nano secrets/*.txt
```

2. Set up Snyk (optional, for security scanning):
```bash
# Install Snyk CLI
npm install -g snyk

# Login to Snyk
snyk auth

# Set SNYK_TOKEN environment variable
export SNYK_TOKEN=your_token_here
```

### Development Workflow

1. Start development environment:
```bash
docker-compose --profile development up
```

2. Run tests:
```bash
docker-compose --profile testing up
```

3. Run security scans:
```bash
docker-compose --profile security up security-scan
```

### Production Deployment

1. Build and start production environment:
```bash
docker-compose --profile production up --build
```

## Security Best Practices

1. **Secrets Management**
   - Never commit actual secret files to version control
   - Use strong, unique passwords for each secret
   - Rotate secrets regularly
   - Use different secrets for different environments

2. **Security Scanning**
   - Run security scans before deploying to production
   - Address high and critical severity issues immediately
   - Keep dependencies updated
   - Monitor security advisories for used images

3. **Container Security**
   - Use multi-stage builds to minimize attack surface
   - Run containers with non-root users
   - Implement resource limits
   - Use read-only root filesystem where possible

## Troubleshooting

### Common Issues

1. **Secret Access Issues**
   - Verify secret files exist and have correct permissions
   - Check environment variables are correctly set
   - Ensure secrets are properly mounted in containers

2. **Security Scan Failures**
   - Verify Snyk token is valid
   - Check for network connectivity
   - Review scan logs for specific issues

3. **Profile-Specific Issues**
   - Verify profile name is correct
   - Check service dependencies
   - Review environment-specific configurations

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [Snyk Documentation](https://docs.snyk.io/) 