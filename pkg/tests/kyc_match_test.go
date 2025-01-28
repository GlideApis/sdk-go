package tests

import (
	"testing"

	"github.com/GlideApis/sdk-go/pkg/services"
	"github.com/GlideApis/sdk-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestKYCMatch(t *testing.T) {
	settings := SetupTestEnvironment(t)

	t.Run("creates client", func(t *testing.T) {
		client := services.NewKYCMatchClient(settings)
		assert.NotNil(t, client)
		assert.Equal(t, "Hello", client.GetHello())
	})

	t.Run("creates user client", func(t *testing.T) {
		client := services.NewKYCMatchClient(settings)
		identifier := types.PhoneIdentifier{PhoneNumber: "+555123456789"}
		userClient, err := client.For(identifier)
		assert.NoError(t, err)
		assert.NotNil(t, userClient)
	})

	t.Run("performs KYC match", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		client := services.NewKYCMatchClient(settings)
		identifier := types.PhoneIdentifier{PhoneNumber: "+555123456789"}
		userClient, err := client.For(identifier)
		assert.NoError(t, err)

		props := types.KYCMatchProps{
			PhoneNumber:          "+555123456789",
			IDDocument:           "66666666q",
			Name:                 "Federica Sanchez Arjona",
			GivenName:            "Federica",
			FamilyName:           "Sanchez Arjona",
			NameKanaHankaku:      "federica",
			NameKanaZenkaku:      "Ｆｅｄｅｒｉｃａ",
			MiddleNames:          "Sanchez",
			FamilyNameAtBirth:    "YYYY",
			Address:              "Tokyo-to Chiyoda-ku Iidabashi 3-10-10",
			StreetName:           "Nicolas Salmeron",
			StreetNumber:         4,
			PostalCode:           1028460,
			Region:               "Tokyo",
			Locality:             "ZZZZ",
			Country:              "Japan",
			HouseNumberExtension: "VVVV",
			Birthdate:            "1978-08-22",
			Email:                "abc@example.com",
			Gender:               "male",
		}

		response, err := userClient.Match(props, types.ApiConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, response)

		// Helper function to check field value
		checkField := func(field *string, fieldName string) {
			if field == nil {
				t.Errorf("Field %s should not be nil", fieldName)
				return
			}
			validValues := []string{"true", "false", "not_available"}
			assert.Contains(t, validValues, *field, "Field %s has invalid value", fieldName)
		}

		// Check each field
		checkField(response.IDDocumentMatch, "IDDocumentMatch")
		checkField(response.NameMatch, "NameMatch")
		checkField(response.GivenNameMatch, "GivenNameMatch")
		checkField(response.FamilyNameMatch, "FamilyNameMatch")
		checkField(response.NameKanaHankakuMatch, "NameKanaHankakuMatch")
		checkField(response.NameKanaZenkakuMatch, "NameKanaZenkakuMatch")
		checkField(response.MiddleNamesMatch, "MiddleNamesMatch")
		checkField(response.FamilyNameAtBirthMatch, "FamilyNameAtBirthMatch")
		checkField(response.AddressMatch, "AddressMatch")
		checkField(response.StreetNameMatch, "StreetNameMatch")
		checkField(response.StreetNumberMatch, "StreetNumberMatch")
		checkField(response.PostalCodeMatch, "PostalCodeMatch")
		checkField(response.RegionMatch, "RegionMatch")
		checkField(response.LocalityMatch, "LocalityMatch")
		checkField(response.CountryMatch, "CountryMatch")
		checkField(response.HouseNumberExtensionMatch, "HouseNumberExtensionMatch")
		checkField(response.BirthdateMatch, "BirthdateMatch")
		checkField(response.EmailMatch, "EmailMatch")
		checkField(response.GenderMatch, "GenderMatch")
	})

	t.Run("handles consent required", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		client := services.NewKYCMatchClient(settings)
		identifier := types.PhoneIdentifier{PhoneNumber: "+555123456789"}
		userClient, err := client.For(identifier)
		assert.NoError(t, err)

		if userClient.RequiresConsent {
			assert.NotEmpty(t, userClient.GetConsentURL())
			err = userClient.PollAndWaitForSession()
			assert.NoError(t, err)
		}
	})

	t.Run("handles errors", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		client := services.NewKYCMatchClient(settings)

		// Test with invalid identifier
		identifier := types.PhoneIdentifier{PhoneNumber: "invalid-phone"}
		userClient, err := client.For(identifier)
		if err == nil {
			// If For() doesn't return error, Match should fail
			props := types.KYCMatchProps{
				PhoneNumber: "invalid-phone",
			}
			_, err = userClient.Match(props, types.ApiConfig{})
		}
		assert.Error(t, err, "Expected an error with invalid phone number")

		// Test with missing required fields
		identifier = types.PhoneIdentifier{PhoneNumber: "+555123456789"}
		userClient, err = client.For(identifier)
		assert.NoError(t, err)

		props := types.KYCMatchProps{
			// Missing required fields
		}

		_, err = userClient.Match(props, types.ApiConfig{})
		assert.Error(t, err, "Expected an error with missing required fields")
	})
}
