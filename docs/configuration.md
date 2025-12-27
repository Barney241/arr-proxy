# Configuration

The proxy uses **environment variables** for app settings and **YAML files** for service definitions.

## Environment Variables

### App Settings

| Variable | Description | Default |
| :--- | :--- | :--- |
| `APP_PORT` | Port to listen on | `8443` |
| `APP_CONFIG_DIR` | Directory containing `sonarr.yaml` and `radarr.yaml` | `./config` |
| `APP_TLS_CERT` | Path to server TLS certificate | - |
| `APP_TLS_KEY` | Path to server TLS private key | - |
| `APP_CA_CERT` | Path to CA certificate (for mTLS) | - |
| `APP_AUTH_MODE` | Authentication mode: `apikey`, `mtls`, or `basic` | `apikey` |
| `APP_API_KEY` | Proxy API key (required for `apikey` mode) | - |
| `APP_BASIC_AUTH_USER` | Username for `basic` mode | - |
| `APP_BASIC_AUTH_PASS` | Password for `basic` mode | - |

### Server Tuning

| Variable | Description | Default |
| :--- | :--- | :--- |
| `APP_READ_TIMEOUT` | HTTP read timeout | `30s` |
| `APP_WRITE_TIMEOUT` | HTTP write timeout | `30s` |
| `APP_IDLE_TIMEOUT` | HTTP idle timeout | `120s` |
| `APP_READ_HEADER_TIMEOUT` | HTTP read header timeout | `20s` |
| `APP_MAX_BODY_SIZE` | Max request body size in bytes | `10485760` (10MB) |
| `APP_TLS_MIN_VERSION` | Minimum TLS version (`1.2` or `1.3`) | `1.2` |
| `APP_LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |

### Service Overrides

Environment variables override YAML config values:

| Variable | Overrides |
| :--- | :--- |
| `SONARR_URL` | `sonarr.yaml: url` |
| `SONARR_API_KEY` | `sonarr.yaml: api_key` |
| `RADARR_URL` | `radarr.yaml: url` |
| `RADARR_API_KEY` | `radarr.yaml: api_key` |

## Service Configuration (YAML)

Place `sonarr.yaml` and `radarr.yaml` in your `APP_CONFIG_DIR`.

```yaml
url: "http://radarr:7878"
api_key: "YOUR_API_KEY"
whitelist:
  - '^/api/v3/movie(?:/.*)?$'
  - '^/api/v3/system/status$'
```

## Whitelist Patterns

Patterns are regex expressions matching API paths. Supports optional method restrictions:

```yaml
whitelist:
  # All methods allowed
  - '^/api/v3/system/status$'

  # GET only
  - 'GET:^/api/v3/movie(?:/.*)?$'

  # Multiple methods
  - 'GET,POST:^/api/v3/queue(?:/.*)?$'

  # Full CRUD
  - 'GET,POST,PUT,DELETE:^/api/v3/series(?:/.*)?$'
```

**Supported methods:** `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`

See [examples/config/](../examples/config/) for complete examples.
