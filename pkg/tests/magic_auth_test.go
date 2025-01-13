package tests

import (
	"net/http"
	"testing"
	"github.com/ClearBlockchain/sdk-go/pkg/glide"
	"github.com/ClearBlockchain/sdk-go/pkg/types"
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
}
