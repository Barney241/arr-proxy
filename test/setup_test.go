package test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"arr-proxy/cmd/proxy"
	"arr-proxy/internal/config"
)

var (
	clientTLSConfig *tls.Config
	proxyURL        string
	caCertPool      *x509.CertPool
)

func TestMain(m *testing.M) {
	// 1. Setup temporary directory for all test assets
	tempDir, err := os.MkdirTemp("", "e2etests_assets")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 2. Generate Certificates
	log.Println("Generating TLS certificates...")
	certs, err := generateCerts(tempDir)
	if err != nil {
		log.Fatalf("Failed to generate certificates: %v", err)
	}

	// 3. Setup Client TLS Config
	clientTLSConfig, err = setupClientTLSConfig(certs.caCertPath, certs.clientCertPath, certs.clientKeyPath)
	if err != nil {
		log.Fatalf("Failed to set up client TLS config: %v", err)
	}

	// 4. Start Mock backend services
	log.Println("Starting mock backend services...")
	sonarrAPIKey := generateRandomString(32)
	sonarrMock := startMockService(sonarrAPIKey)
	defer sonarrMock.Close()

	radarrAPIKey := generateRandomString(32)
	radarrMock := startMockService(radarrAPIKey)
	defer radarrMock.Close()

	// 5. Create temporary config files
	log.Println("Creating temporary configuration files...")
	configDir := filepath.Join(tempDir, "config")
	if err := os.Mkdir(configDir, 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}
	if err := writeConfigFile(configDir, "sonarr.yaml", sonarrMock.URL, sonarrAPIKey); err != nil {
		log.Fatalf("Failed to write sonarr config: %v", err)
	}
	if err := writeConfigFile(configDir, "radarr.yaml", radarrMock.URL, radarrAPIKey); err != nil {
		log.Fatalf("Failed to write radarr config: %v", err)
	}

	// 6. Start the proxy application
	log.Println("Starting proxy application...")
	proxyPort := getFreePortForTest()

	os.Setenv("APP_TLS_CERT", certs.serverCertPath)
	os.Setenv("APP_TLS_KEY", certs.serverKeyPath)
	os.Setenv("APP_CA_CERT", certs.caCertPath)
	os.Setenv("APP_PORT", proxyPort)
	os.Setenv("APP_CONFIG_DIR", configDir)
	os.Setenv("APP_AUTH_MODE", "mtls")

	proxyCfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	url, stop, err := StartProxy(&proxyCfg)
	if err != nil {
		log.Fatalf("Failed to start proxy: %v", err)
	}
	proxyURL = url
	defer stop()
	
	// Run tests
	code := m.Run()

	// Teardown
	log.Println("Tearing down e2e tests...")
	os.Exit(code)
}

func StartProxy(cfg *config.Config) (string, func(), error) {
	stop, err := proxy.Run(cfg)
	if err != nil {
		return "", nil, err
	}

	url := fmt.Sprintf("https://localhost:%s", cfg.Port)
	log.Printf("Proxy process started at: %s", url)

	// Give server a moment to start
	time.Sleep(1 * time.Second)

	// Wrap stop to match expected signature
	return url, func() { stop(context.Background()) }, nil
}

func startMockService(apiKey string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey := r.Header.Get("X-Api-Key")
		log.Printf("Mock received request: %s %s, X-Api-Key: %s", r.Method, r.URL.Path, gotKey)
		if gotKey != apiKey {
			log.Printf("Mock rejecting request due to API key mismatch. Expected: %s, Got: %s", apiKey, gotKey)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v3/system/status") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"version": "mock-version"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

type certPaths struct {
	caCertPath     string
	serverCertPath string
	serverKeyPath  string
	clientCertPath string
	clientKeyPath  string
}

func generateCerts(tempDir string) (*certPaths, error) {
	// CA
	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(2021),
		Subject:               pkix.Name{Organization: []string{"arr-proxy test ca"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * time.Minute),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caPrivKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	caBytes, _ := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	caCertPath := filepath.Join(tempDir, "ca.pem")
	os.WriteFile(caCertPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caBytes}), 0644)

	// Server
	serverCertTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		Subject:      pkix.Name{Organization: []string{"arr-proxy test server"}},
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * time.Minute),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverPrivKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	serverBytes, _ := x509.CreateCertificate(rand.Reader, serverCertTmpl, ca, &serverPrivKey.PublicKey, caPrivKey)
	serverCertPath := filepath.Join(tempDir, "server.pem")
	serverKeyPath := filepath.Join(tempDir, "server.key")
	os.WriteFile(serverCertPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverBytes}), 0644)
	os.WriteFile(serverKeyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey)}), 0644)

	// Client
	clientCertTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2023),
		Subject:      pkix.Name{Organization: []string{"arr-proxy test client"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * time.Minute),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientPrivKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	clientBytes, _ := x509.CreateCertificate(rand.Reader, clientCertTmpl, ca, &clientPrivKey.PublicKey, caPrivKey)
	clientCertPath := filepath.Join(tempDir, "client.pem")
	clientKeyPath := filepath.Join(tempDir, "client.key")
	os.WriteFile(clientCertPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientBytes}), 0644)
	os.WriteFile(clientKeyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientPrivKey)}), 0644)

	return &certPaths{caCertPath, serverCertPath, serverKeyPath, clientCertPath, clientKeyPath}, nil
}

func setupClientTLSConfig(caCertPath, clientCertPath, clientKeyPath string) (*tls.Config, error) {
	caCert, err := os.ReadFile(caCertPath)
	if err != nil { return nil, err }
	caCertPool = x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil { return nil, err }

	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{clientCert},
	}, nil
}

func writeConfigFile(configDir, fileName, url, apiKey string) error {
	content := fmt.Sprintf(`url: "%s"
api_key: "%s"
whitelist:
  - '^/api/v3/system/status$'
  - '^/api/v3/series(?:/.*)?$'
  - '^/api/v3/movie(?:/.*)?$'
  - '^/api/v3/queue$'
`, url, apiKey)
	return os.WriteFile(filepath.Join(configDir, fileName), []byte(content), 0644)
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

func getFreePortForTest() string {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to get free port: %v", err)
	}
	defer listener.Close()
	return fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
}
