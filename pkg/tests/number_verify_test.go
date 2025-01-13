package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/ClearBlockchain/sdk-go/pkg/glide"
	"github.com/ClearBlockchain/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNumberVerify(t *testing.T) {
	settings := SetupTestEnvironment(t)
	glideClient, err := glide.NewGlideClient(settings)
	assert.NoError(t, err)
	t.Run("should work", func(t *testing.T) {
		phoneNumber := "+555123456789"
	    authUrl, err := glideClient.NumberVerify.GetAuthURL(types.NumberVerifyAuthUrlInput{UseDevNumber: phoneNumber})
		fmt.Println("Open this URL on the user's device: ", authUrl)
		assert.NoError(t, err)
		assert.NotNil(t, authUrl)
		assert.NotEmpty(t, authUrl)
		t.Logf("authUrl response: %+v", authUrl)
		baseURL, err := url.Parse(authUrl)
		if err != nil {
			t.Fatalf("Failed to parse authUrl: %v", err)
		}
		query := baseURL.Query()
		// this should be used if not using UseDevNumber in GetAuthURL
		// query.Set("login_hint", "tel:+555123456789")
		baseURL.RawQuery = query.Encode()
		res, _ := MakeRawHttpRequestFollowRedirectChain(baseURL.String())
		location := res.Headers.Get("Location")
		parsedLocation, err := url.Parse(location)
		assert.NoError(t, err)
        code := parsedLocation.Query().Get("code")
		t.Logf("res: %s", res)
		t.Logf("Code: %s", code)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		client, err := glideClient.NumberVerify.For(types.NumberVerifyClientForParams{PhoneNumber: &phoneNumber, Code: code})
		assert.NoError(t, err)
		verify, err := client.VerifyNumber(nil, types.ApiConfig{SessionIdentifier: "session77"})
		assert.NoError(t, err)
		t.Logf("Check verify: %+v", verify)
		assert.Equal(t, true, verify.DevicePhoneNumberVerified, "Response should have devicePhoneNumberVerified=true")
	})

	t.Run("should work with print code", func(t *testing.T) {
		phoneNumber := "+555123456789"
	    authUrl, err := glideClient.NumberVerify.GetAuthURL(types.NumberVerifyAuthUrlInput{UseDevNumber: phoneNumber, PrintCode: true})
		assert.NoError(t, err)
		assert.NotNil(t, authUrl)
		assert.NotEmpty(t, authUrl)
		codeRes, err := http.Get(authUrl)
		if err != nil {
			// handle error
			t.Errorf("Error making request: %v", err)
		}
		defer codeRes.Body.Close()
		codeResParsed, err := io.ReadAll(codeRes.Body)
		if err != nil {
			t.Errorf("Error codeResParsed: %v", err)
		}
		var codeResStruct struct {
			Code string `json:"code"`
		}
		err = json.Unmarshal(codeResParsed, &codeResStruct)
		if err != nil {
			t.Errorf("Error unmarshaling code response: %v", err)
		}
		code := codeResStruct.Code
		assert.NoError(t, err)
		t.Logf("Code: %s", code)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)
		client, err := glideClient.NumberVerify.For(types.NumberVerifyClientForParams{PhoneNumber: &phoneNumber, Code: code})
		assert.NoError(t, err)
		verify, err := client.VerifyNumber(nil, types.ApiConfig{SessionIdentifier: "session77"})
		assert.NoError(t, err)
		t.Logf("Check verify: %+v", verify)
		assert.Equal(t, true, verify.DevicePhoneNumberVerified, "Response should have devicePhoneNumberVerified=true")	
	})
}
