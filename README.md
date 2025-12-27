# Arr-Proxy

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A secure API gateway for Sonarr and Radarr. Enforces endpoint whitelists, injects API keys, and handles authentication.

## Quick Start

```bash
docker run -p 8080:8080 \
  -e APP_PORT=8080 \
  -e APP_API_KEY=my-secret-key \
  -e RADARR_URL=http://radarr:7878 \
  -e RADARR_API_KEY=your-radarr-key \
  barney241/arr-proxy:latest
```

```bash
curl -H "X-Api-Key: my-secret-key" http://localhost:8080/radarr/api/v3/system/status
```

## Features

- **Authentication**: API Key (default), mTLS, or Basic Auth
- **Whitelist Enforcement**: Block endpoints not in your allow-list
- **Method Restrictions**: Limit endpoints to specific HTTP methods (GET, POST, etc.)
- **Secret Injection**: Clients don't need backend API keys
- **Structured Logging**: JSON logs with request tracing

## Endpoints

| Endpoint | Description |
| :--- | :--- |
| `GET /info` | View active configuration |
| `/sonarr/*` | Proxy to Sonarr |
| `/radarr/*` | Proxy to Radarr |

## Documentation

- [Configuration](docs/configuration.md) - Environment variables, YAML config, whitelist patterns
- [Authentication](docs/authentication.md) - API Key, mTLS, Basic Auth setup
- [Certificates](docs/certificates.md) - Generating TLS certificates

## Examples

See [`examples/`](examples/) for Docker Compose configurations:

- [`docker-compose.yml`](examples/docker-compose.yml) - Basic HTTP setup
- [`docker-compose.tls.yml`](examples/docker-compose.tls.yml) - TLS enabled
- [`docker-compose.mtls.yml`](examples/docker-compose.mtls.yml) - Mutual TLS
- [`docker-compose.custom-whitelist.yml`](examples/docker-compose.custom-whitelist.yml) - Custom whitelists
- [`config/`](examples/config/) - Example service configurations

## Architecture

```
Client  ──▶  arr-proxy  ──▶  Sonarr/Radarr
              │
         ┌────┴────┐
         │Whitelist│
         │  Auth   │
         │API Keys │
         └─────────┘
```
