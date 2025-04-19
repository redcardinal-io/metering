# RedCardinal Metering Component

This repository contains the Metering component of the RedCardinal.

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Nix (optional, for development)
- Go 1.24+ (automatically managed if using Nix)

## Configuration

1. Copy the example configuration file to create your environment variables:

2. Update the configuration files to match your environment:
   - For Docker-based development: Use service names as hosts (e.g., `postgres`, `clickhouse`, `redpanda`)
   - For Nix-based development: Use `localhost` for all services

## Running the Application

### Option 1: Using Docker Compose (Recommended for first-time setup)

The simplest way to get started is with Docker Compose, which will set up all required services:

```bash
docker-compose up
```

This will start:

- ClickHouse (database for analytics)
- Postgres (relational database)
- Redpanda (Kafka-compatible event streaming)
- RedCardinal Metering API
- Migration service

The application will be accessible at http://localhost:8000.

### Option 2: Using Nix (Recommended for development)

If you prefer a more flexible development environment:

1. Update the Redpanda configuration in docker-compose.yml:

- Ensure the --advertise-kafka-addr parameter for the redpanda service is set to internal://redpanda:9092,external://localhost:19092
- This allows your locally running application to connect to Redpanda

2. Start the required infrastructure components using Docker Compose:

```bash
docker-compose up clickhouse postgres redpanda redpanda-console ch-ui
```

3. Enter the Nix development shell:

```bash
nix develop --command $SHELL
```

4. Start the application with hot reloading:

```bash
air
```

The Nix environment will automatically provide all required development tools (sqlc, goose, etc.).

## Troubleshooting

- Database connection issues: Verify the host settings in your .env files match your environment
- Port conflicts: Check if any of the required ports (8000, 5432, 8123, etc.) are already in use
- Redpanda connectivity issues: When using Nix, ensure Redpanda's external address in docker-compose.yml is set to localhost:19092 and your application is configured to connect to this address
