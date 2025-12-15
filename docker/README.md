# Docker Configuration Structure

This directory contains all Docker-related files for the UniPilot project, organized for Docker Swarm/Stack deployments.

## Directory Structure

```
docker/
├── Dockerfile.api          # API service Dockerfile
├── Dockerfile.sse          # SSE service Dockerfile
├── Dockerfile.nodejs       # Node.js service Dockerfile
├── compose/
│   ├── stack.yml          # Main Docker Swarm stack file (production)
│   └── dev/                # Local development compose files
│       ├── api.yml         # API service for local dev
│       ├── sse.yml         # SSE service for local dev
│       └── nodejs.yml      # Node.js service for local dev
└── services/
    └── redis/
        ├── docker-compose.redis.yml  # Standalone Redis service
        └── redis.conf                # Redis configuration file
```

## File Usage

### Production Deployment (Docker Swarm)

**Main Stack File:** `docker/compose/stack.yml`
- Used for Docker Swarm stack deployments
- References pre-built images: `unipilot-api`, `unipilot-sse`, `unipilot-nodejs`
- Includes all services: API, SSE, Node.js, Redis, and Redis Insight
- Deploy with: `docker stack deploy -c docker/compose/stack.yml unipilot`

### Local Development

**Individual Service Files:** `docker/compose/dev/*.yml`
- Used for local development and testing
- Build images from Dockerfiles
- Can be used independently: `docker-compose -f docker/compose/dev/api.yml up`

### Building Images

**Dockerfiles:** `docker/Dockerfile.*`
- Used to build images for production
- Build commands:
  ```bash
  docker build -f docker/Dockerfile.api -t unipilot-api:latest .
  docker build -f docker/Dockerfile.sse -t unipilot-sse:latest .
  docker build -f docker/Dockerfile.nodejs -t unipilot-nodejs:latest .
  ```

### Redis Service

**Standalone Redis:** `docker/services/redis/`
- Redis configuration and compose file for standalone Redis testing
- Can be used independently: `docker-compose -f docker/services/redis/docker-compose.redis.yml up`

## Migration Notes

- Old root-level Dockerfiles and compose files have been moved to this structure
- All paths in compose files have been updated to reflect the new structure
- The swarm stack file now references `docker/services/redis/redis.conf` instead of `internal/services/redis/redis.conf`

