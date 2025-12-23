package test

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"

	"arr-proxy/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getFreePort() string {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	return strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
}

func TestDefaultAuthMode(t *testing.T) {
	// Verify that the default auth mode is API Key
	// Temporarily unset the env var set by TestMain
	origMode := os.Getenv("APP_AUTH_MODE")
	os.Unsetenv("APP_AUTH_MODE")
	defer os.Setenv("APP_AUTH_MODE", origMode)

	// Set a dummy API key to pass validation
	origKey := os.Getenv("APP_API_KEY")
	os.Setenv("APP_API_KEY", "dummy")
	defer func() {
		if origKey == "" {
			os.Unsetenv("APP_API_KEY")
		} else {
			os.Setenv("APP_API_KEY", origKey)
		}
	}()

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, config.AuthModeAPIKey, cfg.Auth.Mode)
}

func TestBasicAuth(t *testing.T) {
	// Load base config (inherits TLS paths from env vars set in TestMain)
	cfg, err := config.Load()
	require.NoError(t, err)
	cfg.Port = getFreePort()
	cfg.Auth.Mode = config.AuthModeBasic
	cfg.Auth.BasicAuth.User = "testuser"
	cfg.Auth.BasicAuth.Password = "testpass"

	url, stop, err := StartProxy(&cfg)
	require.NoError(t, err)
	defer stop()

	// Client that trusts the CA but doesn't present a client cert (Basic Auth shouldn't require it)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	t.Run("valid credentials", func(t *testing.T) {
		req, _ := http.NewRequest("GET", url+"/info", nil)
		req.SetBasicAuth("testuser", "testpass")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		req, _ := http.NewRequest("GET", url+"/info", nil)
		req.SetBasicAuth("testuser", "wrongpass")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("missing credentials", func(t *testing.T) {
		req, _ := http.NewRequest("GET", url+"/info", nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestAPIKeyAuth(t *testing.T) {
	cfg, err := config.Load()
	require.NoError(t, err)
	cfg.Port = getFreePort()
	cfg.Auth.Mode = config.AuthModeAPIKey
	cfg.Auth.APIKey = "my-secret-proxy-key"

	url, stop, err := StartProxy(&cfg)
	require.NoError(t, err)
	defer stop()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	t.Run("valid api key via header X-Api-Key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", url+"/info", nil)
		req.Header.Set("X-Api-Key", "my-secret-proxy-key")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("valid api key via query param apikey", func(t *testing.T) {
		req, _ := http.NewRequest("GET", url+"/info?apikey=my-secret-proxy-key", nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid api key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", url+"/info", nil)
		req.Header.Set("X-Api-Key", "wrong-key")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("missing api key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", url+"/info", nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}