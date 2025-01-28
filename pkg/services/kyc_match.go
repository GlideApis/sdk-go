package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/GlideApis/sdk-go/pkg/types"
	"github.com/GlideApis/sdk-go/pkg/utils"
)

type KYCMatchUserClient struct {
	settings        types.GlideSdkSettings
	identifier      types.UserIdentifier
	session         *types.Session
	RequiresConsent bool
	consentURL      string
	authReqID       string
}

func NewKYCMatchUserClient(settings types.GlideSdkSettings, identifier types.UserIdentifier) *KYCMatchUserClient {
	return &KYCMatchUserClient{
		settings:   settings,
		identifier: identifier,
	}
}

func (c *KYCMatchUserClient) GetConsentURL() string {
	return c.consentURL
}

func (c *KYCMatchUserClient) Match(props types.KYCMatchProps, conf types.ApiConfig) (*types.KYCMatchResponse, error) {
	var wg sync.WaitGroup
	if c.settings.Internal.APIBaseURL == "" {
		return nil, fmt.Errorf("[GlideClient] internal.apiBaseUrl is unset")
	}
	if conf.SessionIdentifier != "" {
		c.reportKYCMatchMetric(&wg, conf.SessionIdentifier, "Glide start", "")
	}

	session, err := c.getSession(conf.Session)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"phoneNumber":          props.PhoneNumber,
		"idDocument":           props.IDDocument,
		"name":                 props.Name,
		"givenName":            props.GivenName,
		"familyName":           props.FamilyName,
		"nameKanaHankaku":      props.NameKanaHankaku,
		"nameKanaZenkaku":      props.NameKanaZenkaku,
		"middleNames":          props.MiddleNames,
		"familyNameAtBirth":    props.FamilyNameAtBirth,
		"address":              props.Address,
		"streetName":           props.StreetName,
		"streetNumber":         props.StreetNumber,
		"postalCode":           props.PostalCode,
		"region":               props.Region,
		"locality":             props.Locality,
		"country":              props.Country,
		"houseNumberExtension": props.HouseNumberExtension,
		"birthdate":            props.Birthdate,
		"email":                props.Email,
		"gender":               props.Gender,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := utils.FetchX(c.settings.Internal.APIBaseURL+"/kyc-match/match", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + session.AccessToken,
		},
		Body: string(jsonData),
	})

	if err != nil {
		return nil, fmt.Errorf("[GlideClient]: [kyc-match] FetchX failed for match: %w", err)
	}

	// Add debug logging
	var rawResponse map[string]interface{}
	if err := resp.JSON(&rawResponse); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse raw response: %w", err)
	}
	fmt.Printf("Debug: Raw KYC Match response: %+v\n", rawResponse)

	var result types.KYCMatchResponse
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
	}

	// Helper function to set default value
	setDefault := func(field **string) {
		if *field == nil {
			notAvailable := "not_available"
			*field = &notAvailable
		}
	}

	// Set defaults for nil fields
	setDefault(&result.IDDocumentMatch)
	setDefault(&result.NameMatch)
	setDefault(&result.GivenNameMatch)
	setDefault(&result.FamilyNameMatch)
	setDefault(&result.NameKanaHankakuMatch)
	setDefault(&result.NameKanaZenkakuMatch)
	setDefault(&result.MiddleNamesMatch)
	setDefault(&result.FamilyNameAtBirthMatch)
	setDefault(&result.AddressMatch)
	setDefault(&result.StreetNameMatch)
	setDefault(&result.StreetNumberMatch)
	setDefault(&result.PostalCodeMatch)
	setDefault(&result.RegionMatch)
	setDefault(&result.LocalityMatch)
	setDefault(&result.CountryMatch)
	setDefault(&result.HouseNumberExtensionMatch)
	setDefault(&result.BirthdateMatch)
	setDefault(&result.EmailMatch)
	setDefault(&result.GenderMatch)

	if conf.SessionIdentifier != "" {
		c.reportKYCMatchMetric(&wg, conf.SessionIdentifier, "Glide match complete", "")
	}
	wg.Wait()
	return &result, nil
}

func (c *KYCMatchUserClient) StartSession() error {
	if c.settings.ClientID == "" || c.settings.ClientSecret == "" {
		return fmt.Errorf("[GlideClient] Client credentials are required to generate a new session")
	}
	var loginHint string
	switch identifier := c.identifier.(type) {
	case types.PhoneIdentifier:
		loginHint = "tel:" + utils.FormatPhoneNumber(identifier.PhoneNumber)
	case types.IpIdentifier:
		loginHint = "ipport:" + identifier.IPAddress
	}
	data := url.Values{}
	data.Set("scope", "kyc-match")
	if loginHint != "" {
		data.Set("login_hint", loginHint)
	}
	resp, err := utils.FetchX(c.settings.Internal.AuthBaseURL+"/oauth2/backchannel-authentication", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(c.settings.ClientID+":"+c.settings.ClientSecret)),
		},
		Body: data.Encode(),
	})
	if err != nil {
		return fmt.Errorf("[GlideClient] FetchX failed: %w", err)
	}
	var body struct {
		ConsentURL string `json:"consentUrl"`
		AuthReqID  string `json:"auth_req_id"`
	}
	if err := resp.JSON(&body); err != nil {
		return fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
	}
	if body.ConsentURL != "" {
		c.RequiresConsent = true
		c.consentURL = body.ConsentURL
	}
	c.authReqID = body.AuthReqID

	return nil
}

func (c *KYCMatchUserClient) getSession(confSession *types.Session) (*types.Session, error) {
	if confSession != nil {
		return confSession, nil
	}

	if c.session != nil && c.session.ExpiresAt > time.Now().Add(time.Minute).Unix() && contains(c.session.Scopes, "kyc-match") {
		return c.session, nil
	}

	session, err := c.generateNewSession()
	if err != nil {
		return nil, err
	}

	c.session = session
	return session, nil
}

func (c *KYCMatchUserClient) PollAndWaitForSession() error {
	for {
		_, err := c.getSession(nil)
		if err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

func (c *KYCMatchUserClient) generateNewSession() (*types.Session, error) {
	if c.settings.ClientID == "" || c.settings.ClientSecret == "" {
		return nil, fmt.Errorf("[GlideClient] Client credentials are required to generate a new session")
	}

	if c.authReqID == "" {
		if err := c.StartSession(); err != nil {
			return nil, err
		}
	}

	if c.authReqID == "" {
		return nil, fmt.Errorf("[GlideClient] Failed to start session")
	}

	data := url.Values{}
	data.Set("grant_type", "urn:openid:params:grant-type:ciba")
	data.Set("auth_req_id", c.authReqID)

	resp, err := utils.FetchX(c.settings.Internal.AuthBaseURL+"/oauth2/token", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(c.settings.ClientID+":"+c.settings.ClientSecret)),
		},
		Body: data.Encode(),
	})

	if err != nil {
		c.authReqID = ""
		return nil, fmt.Errorf("[GlideClient] FetchX failed: %w", err)
	}

	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := resp.JSON(&body); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
	}

	return &types.Session{
		AccessToken: body.AccessToken,
		ExpiresAt:   time.Now().Unix() + body.ExpiresIn,
		Scopes:      strings.Split(body.Scope, " "),
	}, nil
}

func (c *KYCMatchUserClient) reportKYCMatchMetric(wg *sync.WaitGroup, sessionId, metricName string, operator string) {
	metric := types.MetricInfo{
		Operator:   operator,
		Timestamp:  time.Now(),
		SessionId:  sessionId,
		MetricName: metricName,
		Api:        "kyc-match",
		ClientId:   c.settings.ClientID,
	}
	wg.Add(1)
	go func(m types.MetricInfo) {
		defer wg.Done()
		utils.ReportMetric(m)
	}(metric)
}

// Main client for KYC match operations
type KYCMatchClient struct {
	settings types.GlideSdkSettings
}

func NewKYCMatchClient(settings types.GlideSdkSettings) *KYCMatchClient {
	return &KYCMatchClient{settings: settings}
}

func (c *KYCMatchClient) For(identifier types.UserIdentifier) (*KYCMatchUserClient, error) {
	client := NewKYCMatchUserClient(c.settings, identifier)
	err := client.StartSession()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *KYCMatchClient) GetHello() string {
	return "Hello"
}
