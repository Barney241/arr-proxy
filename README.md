# Arr-Proxy

A secure, permission-limited API gateway for Sonarr and Radarr. It sits between external clients and your media servers, enforcing whitelists, injecting API keys, and handling authentication.

## Quick Start

```bash
# 1. Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
services:
  proxy:
    image: barney241/arr-proxy:latest
    ports:
      - "8080:8080"
    environment:
      - APP_PORT=8080
      - APP_API_KEY=my-secure-proxy-key
      # Configure the services you use (one or both)
      - RADARR_URL=http://radarr:7878
      - RADARR_API_KEY=your-radarr-api-key
      # - SONARR_URL=http://sonarr:8989
      # - SONARR_API_KEY=your-sonarr-api-key
EOF

# 2. Start the proxy
docker-compose up -d

# 3. Test it
curl -H "X-Api-Key: my-secure-proxy-key" http://localhost:8080/info
curl -H "X-Api-Key: my-secure-proxy-key" http://localhost:8080/radarr/api/v3/system/status
```

### With HTTPS (Recommended for Production)

For production deployments, enable TLS:

```bash
# 1. Generate TLS certificates
mkdir -p certs && cd certs
openssl req -x509 -newkey rsa:4096 -keyout ca.key -out ca.crt -days 365 -nodes -subj "/CN=ArrProxyCA"
openssl req -newkey rsa:4096 -keyout server.key -out server.csr -nodes -subj "/CN=localhost"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365
cd ..

# 2. Create docker-compose.yml with TLS
cat > docker-compose.yml << 'EOF'
services:
  proxy:
    image: barney241/arr-proxy:latest
    ports:
      - "8443:8443"
    environment:
      - APP_API_KEY=my-secure-proxy-key
      - APP_TLS_CERT=/certs/server.crt
      - APP_TLS_KEY=/certs/server.key
      - APP_CA_CERT=/certs/ca.crt
      - RADARR_URL=http://radarr:7878
      - RADARR_API_KEY=your-radarr-api-key
    volumes:
      - ./certs:/certs
EOF

# 3. Start and test
docker-compose up -d
curl -k -H "X-Api-Key: my-secure-proxy-key" https://localhost:8443/info
```

### Custom Whitelists

To customize allowed endpoints, create config files and mount them:

```bash
mkdir -p config

# Create radarr.yaml with custom whitelist
cat > config/radarr.yaml << 'EOF'
url: ""  # Set via RADARR_URL env var
api_key: ""  # Set via RADARR_API_KEY env var
whitelist:
  - '^/api/v3/movie(?:/.*)?$'
  - '^/api/v3/queue$'
EOF

# Add volume mount to docker-compose.yml:
#   - ./config:/config
```

## Features

- **Flexible Authentication**: Supports API Key (Default), mTLS (Recommended), or Basic Auth.
- **Plug & Play**: Default API Key mode works identically to standard Sonarr/Radarr clients.
- **Whitelist Enforcement**: Only allows specific API paths defined in YAML configuration.
- **Secret Management**: Injects backend API keys so clients don't need them.
- **Observability**: Structured JSON logging and an authenticated `/info` endpoint to view active configurations.

## Configuration

The proxy is configured using a combination of **Environment Variables** (for app settings) and **YAML files** (for service definitions). Environment variables can override YAML config values.

### 1. Environment Variables

#### App Settings

| Variable | Description | Default |
| :--- | :--- | :--- |
| `APP_PORT` | Port to listen on | `8443` |
| `APP_CONFIG_DIR` | Directory containing `sonarr.yaml` and `radarr.yaml` | `./config` |
| `APP_TLS_CERT` | Path to server TLS certificate | Required |
| `APP_TLS_KEY` | Path to server TLS private key | Required |
| `APP_CA_CERT` | Path to CA certificate (Used for mTLS) | Required |
| `APP_AUTH_MODE` | Authentication mode: `apikey`, `mtls`, or `basic` | `apikey` |
| `APP_API_KEY` | Proxy Auth Key (Required for `apikey` mode) | - |
| `APP_BASIC_AUTH_USER` | Username for `basic` mode | - |
| `APP_BASIC_AUTH_PASS` | Password for `basic` mode | - |

#### Server Tuning

These settings can also be configured in `config/server.yaml`:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `APP_READ_TIMEOUT` | HTTP read timeout | `30s` |
| `APP_WRITE_TIMEOUT` | HTTP write timeout | `30s` |
| `APP_IDLE_TIMEOUT` | HTTP idle timeout | `120s` |
| `APP_READ_HEADER_TIMEOUT` | HTTP read header timeout | `20s` |
| `APP_MAX_BODY_SIZE` | Max request body size in bytes | `10485760` (10MB) |
| `APP_TLS_MIN_VERSION` | Minimum TLS version (`1.2` or `1.3`) | `1.2` |
| `APP_LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |

#### Service Overrides

These environment variables override values from YAML config files:

| Variable | Description | Overrides |
| :--- | :--- | :--- |
| `SONARR_URL` | Sonarr instance URL | `sonarr.yaml: url` |
| `SONARR_API_KEY` | Sonarr API key | `sonarr.yaml: api_key` |
| `RADARR_URL` | Radarr instance URL | `radarr.yaml: url` |
| `RADARR_API_KEY` | Radarr API key | `radarr.yaml: api_key` |

### 2. Service Configuration (YAML)

Place these files in your `APP_CONFIG_DIR`. The default configs are included in the Docker image at `/config`.

**`sonarr.yaml`**
```yaml
url: "http://sonarr:8989"
api_key: "YOUR_SONARR_API_KEY"
whitelist:
  - '^/api/v3/series(?:/.*)?$'
  - '^/api/v3/episode(?:/.*)?$'
  - '^/api/v3/queue(?:/.*)?$'
  - '^/api/v3/system/status$'
  - '^/api/v3/tag(?:/.*)?$'
```

**`radarr.yaml`**
```yaml
url: "http://radarr:7878"
api_key: "YOUR_RADARR_API_KEY"
whitelist:
  - '^/api/v3/movie(?:/.*)?$'
  - '^/api/v3/queue(?:/.*)?$'
  - '^/api/v3/system/status$'
  - '^/api/v3/tag(?:/.*)?$'
```

### 3. Customizing Whitelists

The whitelist patterns are regex expressions that match API paths. To customize which endpoints are allowed:

**Option A: Mount custom config files**
```yaml
volumes:
  - ./my-config:/config  # Your custom sonarr.yaml and radarr.yaml
```

**Option B: Override just the service URLs/keys via environment**
```yaml
environment:
  - SONARR_URL=http://my-sonarr:8989
  - SONARR_API_KEY=my-api-key
  # Whitelist still comes from the mounted config file
```

**Common whitelist patterns:**
```yaml
# Allow all series endpoints
- '^/api/v3/series(?:/.*)?$'

# Allow only GET on movies (read-only)
- '^/api/v3/movie$'
- '^/api/v3/movie/\d+$'

# Allow specific endpoints
- '^/api/v3/queue$'
- '^/api/v3/system/status$'

# Allow calendar endpoint
- '^/api/v3/calendar(?:/.*)?$'
```

## Authentication Modes

### API Key (*Default / Plug & Play*)
This mode is designed for maximum compatibility with standard Sonarr/Radarr clients (like mobile apps).
- **Setup**: Set `APP_API_KEY=your-secret-key`.
- **Usage**: Send the key in the `X-Api-Key` header **OR** as an `apikey` query parameter.
- **Why**: It allows your favorite apps to work through the proxy without any code changes.

### mTLS (Mutual TLS) - *Highly Recommended*
Requires the client to present a valid certificate signed by your trusted CA.
- **Setup**: Set `APP_AUTH_MODE=mtls`.
- **Why**: **Best Security**. Unlike static keys, mTLS uses cryptographic proof of identity. A stolen API key is enough to gain access, but a stolen client certificate usually requires the associated private key (often stored in a hardware secure module or protected by a passphrase). It also eliminates the risk of "sharing" secrets.

### Basic Auth
Standard HTTP Basic Authentication.
- **Setup**: Set `APP_AUTH_MODE=basic`, `APP_BASIC_AUTH_USER`, and `APP_BASIC_AUTH_PASS`.
- **Why**: Good for quick browser-based access where certificate installation is not feasible.

## Generating Certificates

You can use `openssl` to generate the necessary certificates.

1. **Generate CA:**
   ```bash
   openssl req -x509 -newkey rsa:4096 -keyout ca.key -out ca.crt -days 365 -nodes -subj "/CN=MyProxyCA"
   ```

2. **Generate Server Cert (for the proxy):**
   ```bash
   openssl req -newkey rsa:4096 -keyout server.key -out server.csr -nodes -subj "/CN=localhost"
   openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365
   ```

3. **Generate Client Cert (for mTLS mode):**
   ```bash
   openssl req -newkey rsa:4096 -keyout client.key -out client.csr -nodes -subj "/CN=my-client"
   openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365
   ```

## Docker Deployment

### Docker Hub
Official images are available at: `barney241/arr-proxy:latest`

### Quick Start

```yaml
services:
  proxy:
    image: barney241/arr-proxy:latest
    ports:
      - "8443:8443"
    environment:
      - APP_API_KEY=my-secure-proxy-key
      - APP_TLS_CERT=/certs/server.crt
      - APP_TLS_KEY=/certs/server.key
      - APP_CA_CERT=/certs/ca.crt
      - SONARR_API_KEY=your-sonarr-key
      - RADARR_API_KEY=your-radarr-key
    volumes:
      - ./certs:/certs
```

### With Custom Whitelists

```yaml
services:
  proxy:
    image: barney241/arr-proxy:latest
    ports:
      - "8443:8443"
    environment:
      - APP_API_KEY=my-secure-proxy-key
      - APP_TLS_CERT=/certs/server.crt
      - APP_TLS_KEY=/certs/server.key
      - APP_CA_CERT=/certs/ca.crt
    volumes:
      - ./certs:/certs
      - ./my-config:/config  # Custom sonarr.yaml and radarr.yaml
```

### Full Stack Example

See `docker-compose.yml` for a complete example with Sonarr and Radarr services.

## API Endpoints

| Endpoint | Description |
| :--- | :--- |
| `GET /info` | Returns current proxy configuration (URLs, whitelists) |
| `/sonarr/*` | Proxied to Sonarr (whitelist enforced) |
| `/radarr/*` | Proxied to Radarr (whitelist enforced) |

### Accessing Services Through the Proxy

Once the proxy is running, your services are available at:

| Service | Proxy URL | Original API Path |
| :--- | :--- | :--- |
| Sonarr | `http://localhost:8080/sonarr/api/v3/...` | `/api/v3/...` |
| Radarr | `http://localhost:8080/radarr/api/v3/...` | `/api/v3/...` |

**Examples:**

```bash
# Get Radarr system status
curl -H "X-Api-Key: my-secure-proxy-key" \
  http://localhost:8080/radarr/api/v3/system/status

# Get all movies from Radarr
curl -H "X-Api-Key: my-secure-proxy-key" \
  http://localhost:8080/radarr/api/v3/movie

# Get Sonarr series list
curl -H "X-Api-Key: my-secure-proxy-key" \
  http://localhost:8080/sonarr/api/v3/series

# Get Sonarr queue
curl -H "X-Api-Key: my-secure-proxy-key" \
  http://localhost:8080/sonarr/api/v3/queue

# Using query parameter instead of header
curl "http://localhost:8080/radarr/api/v3/movie?apikey=my-secure-proxy-key"
```

**Configure your apps:**

When configuring mobile apps or other clients to use the proxy:
- **URL**: `http://your-proxy-host:8080/radarr` (or `/sonarr`)
- **API Key**: Your proxy API key (`APP_API_KEY`), not the backend Radarr/Sonarr key

The proxy automatically injects the real API keys when forwarding requests to the backend services.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  arr-proxy  │────▶│   Sonarr    │
│  (App/CLI)  │     │  (Gateway)  │     │   Radarr    │
└─────────────┘     └─────────────┘     └─────────────┘
                          │
                    ┌─────┴─────┐
                    │ Whitelist │
                    │ Auth Mode │
                    │ API Key   │
                    │ Injection │
                    └───────────┘
```
