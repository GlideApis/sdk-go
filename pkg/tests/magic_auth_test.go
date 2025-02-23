package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/GlideApis/sdk-go/pkg/glide"
	"github.com/GlideApis/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestMagicAuth(t *testing.T) {
	settings := SetupTestEnvironment(t)
	glideClient, err := glide.NewGlideClient(settings)
	assert.NoError(t, err)
	t.Run("should start magic auth", func(t *testing.T) {
		magicRes, err := glideClient.MagicAuth.StartAuth(types.MagicAuthStartProps{
			PhoneNumber: "+555123456789",
		}, types.ApiConfig{SessionIdentifier: "magic_auth_test_55"})
		assert.NoError(t, err)
		assert.NotNil(t, magicRes)
		assert.Equal(t, "MAGIC", magicRes.Type)
		assert.NotEmpty(t, magicRes.AuthURL)
		t.Logf("Magic auth StartAuth response: %+v", magicRes)
		tokenRes, err := http.Get(magicRes.AuthURL)
		if err != nil {
			// handle error
			t.Errorf("Error making request: %v", err)
		}
		t.Logf("Magic auth token request res: %+v", tokenRes)
		defer tokenRes.Body.Close()
		token := tokenRes.Header.Get("token")
		if token == "" {
			t.Errorf("Token not found in response")
		}
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		verifyRes, err := glideClient.MagicAuth.VerifyAuth(types.MagicAuthVerifyProps{
			PhoneNumber: "+555123456789",
			Token:       token,
		}, types.ApiConfig{SessionIdentifier: "magic_auth_test_55"})
		assert.NoError(t, err)
		t.Logf("Check verify: %+v", verifyRes)
		assert.True(t, verifyRes.Verified)
	})

	t.Run("should start and check server auth with polling", func(t *testing.T) {
		// Start server auth process
		serverAuthRes, err := glideClient.MagicAuth.StartServerAuth(types.MagicAuthStartProps{
			PhoneNumber: "+555123456789",
		}, types.ApiConfig{SessionIdentifier: "magic_auth_server_test_55"})
		assert.NoError(t, err)
		assert.NotNil(t, serverAuthRes)
		assert.NotEmpty(t, serverAuthRes.SessionID)
		assert.NotEmpty(t, serverAuthRes.AuthURL)
		t.Logf("Magic auth StartServerAuth response: %+v", serverAuthRes)

		// Make request to auth URL (simulating user interaction)
		res, err := MakeRawHttpRequestFollowRedirectChain(serverAuthRes.AuthURL)
		assert.NoError(t, err)
		t.Logf("Auth URL response: %+v", res)

		// Poll until verification is complete or timeout
		timeout := time.After(120 * time.Second)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		var lastCheckRes *types.MagicAuthCheckServerAuthResponse
		var checkErr error

		for {
			select {
			case <-timeout:
				t.Fatal("Auth timed out after 120 seconds")
				return
			case <-ticker.C:
				lastCheckRes, checkErr = glideClient.MagicAuth.CheckServerAuth(
					serverAuthRes.SessionID,
					types.ApiConfig{SessionIdentifier: "magic_auth_server_test_55"},
				)
				assert.NoError(t, checkErr)
				assert.NotNil(t, lastCheckRes)
				t.Logf("Magic auth CheckServerAuth response: %+v", lastCheckRes)

				if lastCheckRes.Verified {
					t.Log("Authentication successful!")
					return
				}
				if lastCheckRes.Status == "COMPLETED" && !lastCheckRes.Verified {
					t.Fatal("Auth completed but not verified")
					return
				}
			}
		}
	})
}
