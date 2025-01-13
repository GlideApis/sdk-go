package tests

import (
	"testing"
	"github.com/ClearBlockchain/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
    "github.com/ClearBlockchain/sdk-go/pkg/glide"
)


func TestTelcoFinderClient(t *testing.T) {
    settings := SetupTestEnvironment(t)
	glideClient, err := glide.NewGlideClient(settings)
    assert.NoError(t, err)
	// Test NetworkIdForNumber
	t.Run("NetworkIdForNumber", func(t *testing.T) {
        conf := types.ApiConfig{}
        if conf.Session != nil {
            t.Logf("Debug: Provided session in conf: %+v", conf.Session)
        } else {
            t.Log("Debug: No session provided in conf")
        }
        response, err := glideClient.TelcoFinder.NetworkIdForNumber("+34630844671", conf)
        if err != nil {
            t.Fatalf("NetworkIdForNumber failed: %v", err)
        }
        if response.NetworkID == "" {
            t.Errorf("NetworkIdForNumber returned empty network ID")
        }
        t.Logf("Debug: NetworkIdForNumber response: %+v", response)
        assert.NotEmpty(t, response.NetworkID, "NetworkID should not be empty")
        assert.Equal(t, "21407", response.NetworkID, "NetworkID should be 21407")
    })

	t.Run("LookupIp", func(t *testing.T) {
            response, err := glideClient.TelcoFinder.LookupIp("80.58.0.0", types.ApiConfig{})
            assert.NoError(t, err, "LookupIp should not return an error")
            assert.NotNil(t, response, "Response should not be nil")
            t.Logf("LookupIp response: %+v", response)
            assert.Equal(t, "ipport:80.58.0.0", response.Subject, "Subject should match the input IP")
            assert.NotEmpty(t, response.Properties.OperatorID, "OperatorID should not be empty")
            assert.Equal(t, "Movistar", response.Properties.OperatorID, "OperatorID should be Movistar")
            assert.NotEmpty(t, response.Links, "Links should not be empty")
    })

    t.Run("LookupNumber", func(t *testing.T) {
            response, err := glideClient.TelcoFinder.LookupNumber("+555123456789", types.ApiConfig{})
            assert.NoError(t, err, "LookupNumber should not return an error")
            assert.NotNil(t, response, "Response should not be nil")
            t.Logf("LookupNumber response: %+v", response)
            assert.Equal(t, "tel:+555123456789", response.Subject, "Subject should match the input phone number")
            assert.Equal(t, "Glide Test Lab", response.Properties.OperatorID, "OperatorID should be Glide Test Lab")
            assert.NotEmpty(t, response.Links, "Links should not be empty")
    })
}
