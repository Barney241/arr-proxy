# Authentication Modes

## API Key (Default)

Designed for compatibility with standard Sonarr/Radarr clients.

```yaml
environment:
  - APP_AUTH_MODE=apikey  # or omit (default)
  - APP_API_KEY=your-secret-key
```

**Usage:** Send the key via `X-Api-Key` header or `apikey` query parameter.

```bash
# Header
curl -H "X-Api-Key: your-secret-key" http://localhost:8080/radarr/api/v3/movie

# Query parameter
curl "http://localhost:8080/radarr/api/v3/movie?apikey=your-secret-key"
```

## mTLS (Mutual TLS)

**Recommended for maximum security.** Requires client certificates signed by your CA.

```yaml
environment:
  - APP_AUTH_MODE=mtls
  - APP_TLS_CERT=/certs/server.crt
  - APP_TLS_KEY=/certs/server.key
  - APP_CA_CERT=/certs/ca.crt
```

**Usage:** Configure your client with the client certificate.

```bash
curl --cert client.crt --key client.key --cacert ca.crt \
  https://localhost:8443/radarr/api/v3/movie
```

**Why mTLS?** Unlike API keys, mTLS uses cryptographic proof of identity. A stolen key grants immediate access; a stolen certificate still requires its private key.

## Basic Auth

Standard HTTP Basic Authentication.

```yaml
environment:
  - APP_AUTH_MODE=basic
  - APP_BASIC_AUTH_USER=admin
  - APP_BASIC_AUTH_PASS=secret
```

**Usage:**

```bash
curl -u admin:secret http://localhost:8080/radarr/api/v3/movie
```

See [certificates.md](certificates.md) for generating TLS certificates.
