package test

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoEndpoint(t *testing.T) {
	client := newTestClient()
	req, err := http.NewRequest("GET", proxyURL+"/info", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var info struct {
		Sonarr struct {
			URL       string   `json:"url"`
			Whitelist []string `json:"whitelist"`
		} `json:"sonarr"`
		Radarr struct {
			URL       string   `json:"url"`
			Whitelist []string `json:"whitelist"`
		} `json:"radarr"`
	}

	err = json.NewDecoder(resp.Body).Decode(&info)
	require.NoError(t, err)

	assert.NotEmpty(t, info.Sonarr.URL)
	assert.NotEmpty(t, info.Radarr.URL)
	assert.Contains(t, info.Sonarr.Whitelist, `^/api/v3/system/status$`)
	assert.Contains(t, info.Radarr.Whitelist, `^/api/v3/system/status$`)
}

func TestProxyEndpoint(t *testing.T) {
	client := newTestClient()

	testCases := []struct {
		name         string
		service      string
		path         string
		method       string
		body         string
		expectedCode int
	}{
		// Basic whitelist tests
		{"sonarr whitelisted", "sonarr", "/api/v3/system/status", "GET", "", http.StatusOK},
		{"sonarr not whitelisted", "sonarr", "/api/v3/nonexistent", "GET", "", http.StatusForbidden},
		{"radarr whitelisted", "radarr", "/api/v3/system/status", "GET", "", http.StatusOK},
		{"radarr not whitelisted", "radarr", "/api/v3/nonexistent", "GET", "", http.StatusForbidden},

		// Method restriction tests - GET only endpoint
		{"readonly GET allowed", "sonarr", "/api/v3/readonly", "GET", "", http.StatusOK},
		{"readonly POST blocked", "sonarr", "/api/v3/readonly", "POST", "", http.StatusForbidden},
		{"readonly DELETE blocked", "sonarr", "/api/v3/readonly", "DELETE", "", http.StatusForbidden},

		// Method restriction tests - DELETE only endpoint
		{"deleteonly DELETE allowed", "sonarr", "/api/v3/deleteonly", "DELETE", "", http.StatusOK},
		{"deleteonly GET blocked", "sonarr", "/api/v3/deleteonly", "GET", "", http.StatusForbidden},

		// Method restriction tests - GET,POST endpoint (movie)
		{"movie GET allowed", "radarr", "/api/v3/movie", "GET", "", http.StatusOK},
		{"movie POST allowed", "radarr", "/api/v3/movie", "POST", "", http.StatusOK},
		{"movie DELETE blocked", "radarr", "/api/v3/movie", "DELETE", "", http.StatusForbidden},
		{"movie PUT blocked", "radarr", "/api/v3/movie", "PUT", "", http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := proxyURL + "/" + tc.service + tc.path
			req, err := http.NewRequest(tc.method, url, strings.NewReader(tc.body))
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode, "Expected %d for %s %s", tc.expectedCode, tc.method, tc.path)

			// Only check response body for known backend endpoints
			if tc.expectedCode == http.StatusOK && tc.path == "/api/v3/system/status" {
				var result map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&result)
				assert.NoError(t, err, "Response body should be valid JSON")
				assert.NotEmpty(t, result["version"], "Response should contain a version field")
			}
		})
	}
}

func TestAuthentication(t *testing.T) {
	t.Run("no client cert", func(t *testing.T) {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: caCertPool,
				},
			},
		}
		
		_, err := client.Get(proxyURL + "/info")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "remote error: tls: certificate required")
	})

	t.Run("with client cert", func(t *testing.T) {
		client := newTestClient()
		resp, err := client.Get(proxyURL + "/info")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
