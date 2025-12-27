# Generating TLS Certificates

## Quick Setup

```bash
mkdir -p certs && cd certs

# 1. Generate CA
openssl req -x509 -newkey rsa:4096 -keyout ca.key -out ca.crt \
  -days 365 -nodes -subj "/CN=ArrProxyCA"

# 2. Generate Server Certificate
openssl req -newkey rsa:4096 -keyout server.key -out server.csr \
  -nodes -subj "/CN=localhost"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out server.crt -days 365

# 3. Generate Client Certificate (for mTLS)
openssl req -newkey rsa:4096 -keyout client.key -out client.csr \
  -nodes -subj "/CN=my-client"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key \
  -CAcreateserial -out client.crt -days 365

cd ..
```

## Files Created

| File | Purpose |
| :--- | :--- |
| `ca.crt` | Certificate Authority - used to verify client/server certs |
| `ca.key` | CA private key - keep secure, used to sign new certs |
| `server.crt` | Server certificate - presented to clients |
| `server.key` | Server private key - keep secure |
| `client.crt` | Client certificate - for mTLS authentication |
| `client.key` | Client private key - keep secure |

## Docker Volume Mount

```yaml
volumes:
  - ./certs:/certs
environment:
  - APP_TLS_CERT=/certs/server.crt
  - APP_TLS_KEY=/certs/server.key
  - APP_CA_CERT=/certs/ca.crt
```

## Production Notes

- Use longer validity periods for production (e.g., `-days 3650`)
- Consider using a proper PKI or ACME (Let's Encrypt) for public-facing deployments
- Store private keys securely (restricted file permissions, secrets management)
- For SAN (Subject Alternative Names), add `-addext "subjectAltName=DNS:yourdomain.com"`
