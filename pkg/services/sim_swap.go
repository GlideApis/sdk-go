package services

import (
	"encoding/json"
	"fmt"
	"github.com/ClearBlockchain/sdk-go/pkg/types"
	"github.com/ClearBlockchain/sdk-go/pkg/utils"
	"encoding/base64"
    "net/url"
    "strings"
    "time"
)


type SimSwapCheckResponse struct {
	Swapped bool `json:"swapped"`
}


type SimSwapRetrieveDateResponse struct {
	LatestSimChange string `json:"latestSimChange"`
}

type SimSwapUserClient struct {
	settings         types.GlideSdkSettings
	identifier       types.UserIdentifier
	session          *types.Session
	RequiresConsent  bool
	consentURL       string
	authReqID        string
}

func NewSimSwapUserClient(settings types.GlideSdkSettings, identifier types.UserIdentifier) *SimSwapUserClient {
	return &SimSwapUserClient{
		settings:   settings,
		identifier: identifier,
	}
}

func (c *SimSwapUserClient) GetConsentURL() string {
	return c.consentURL
}

// Check performs a SIM swap check
func (c *SimSwapUserClient) Check(params types.SimSwapCheckParams, conf types.ApiConfig) (*SimSwapCheckResponse, error) {
	if c.settings.Internal.APIBaseURL == "" {
		return nil, fmt.Errorf("[GlideClient] internal.apiBaseUrl is unset")
	}
	phoneNumber := params.PhoneNumber
	if phoneNumber == "" {
		if phoneIdentifier, ok := c.identifier.(types.PhoneIdentifier); ok {
			phoneNumber = phoneIdentifier.PhoneNumber
		} else {
			return nil, fmt.Errorf("[GlideClient] phone number not provided")
		}
	}
	session, err := c.getSession(conf.Session)
	if err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to get session: %w", err)
	}
	body := map[string]interface{}{
		"phoneNumber": utils.FormatPhoneNumber(phoneNumber),
	}
	if params.MaxAge != nil {
		body["maxAge"] = *params.MaxAge
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to marshal request body: %w", err)
	}
	resp, err := utils.FetchX(c.settings.Internal.APIBaseURL+"/sim-swap/check", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + session.AccessToken,
		},
		Body: string(bodyJSON),
	})
	if err != nil {
		if fetchErr, ok := err.(*utils.FetchError); ok && fetchErr.Response.StatusCode == 404 {
			return nil, fmt.Errorf("[GlideClient] Network ID not found for number %s", phoneNumber)
		}
		return nil, fmt.Errorf("[GlideClient] FetchX failed: %w", err)
	}
	var result SimSwapCheckResponse
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
	}
	return &result, nil
}


// RetrieveDate retrieves the date of the latest SIM swap
func (c *SimSwapUserClient) RetrieveDate(params types.SimSwapRetrieveDateParams, conf types.ApiConfig) (*SimSwapRetrieveDateResponse, error) {
	if c.settings.Internal.APIBaseURL == "" {
		return nil, fmt.Errorf("[GlideClient] internal.apiBaseUrl is unset")
	}

	phoneNumber := params.PhoneNumber
	if phoneNumber == "" {
		if phoneIdentifier, ok := c.identifier.(types.PhoneIdentifier); ok {
			phoneNumber = phoneIdentifier.PhoneNumber
		} else {
			return nil, fmt.Errorf("[GlideClient] phone number not provided")
		}
	}

	session, err := c.getSession(conf.Session)
	if err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to get session: %w", err)
	}

	body := map[string]string{
		"phoneNumber": utils.FormatPhoneNumber(phoneNumber),
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to marshal request body: %w", err)
	}

	resp, err := utils.FetchX(c.settings.Internal.APIBaseURL+"/sim-swap/retrieve-date", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + session.AccessToken,
		},
		Body: string(bodyJSON),
	})

	if err != nil {
		if fetchErr, ok := err.(*utils.FetchError); ok && fetchErr.Response.StatusCode == 404 {
			return nil, fmt.Errorf("[GlideClient] Network ID not found for number %s", phoneNumber)
		}
		return nil, fmt.Errorf("[GlideClient] FetchX failed: %w", err)
	}

	var result SimSwapRetrieveDateResponse
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
	}

	return &result, nil
}

func (c *SimSwapUserClient) StartSession() error {
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
	data.Set("scope", "sim-swap")
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


func (c *SimSwapUserClient) getSession(confSession *types.Session) (*types.Session, error) {
    if confSession != nil {
        fmt.Println("Debug: Using provided session")
        return confSession, nil
    }

    if c.session != nil && c.session.ExpiresAt > time.Now().Add(time.Minute).Unix() && contains(c.session.Scopes, "sim-swap") {
        fmt.Println("Debug: Using cached session")
        return c.session, nil
    }

    fmt.Println("Debug: Generating new session")
    session, err := c.generateNewSession()
    if err != nil {
        return nil, fmt.Errorf("failed to generate new session: %w", err)
    }
    c.authReqID = ""
    c.session = session
    return session, nil
}

// PollAndWaitForSession continuously polls for a valid session
func (c *SimSwapUserClient) PollAndWaitForSession() error {
	for {
		_, err := c.getSession(nil)
		if err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
}

// generateNewSession generates a new session
func (c *SimSwapUserClient) generateNewSession() (*types.Session, error) {
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
		ExpiresIn   int64    `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := resp.JSON(&body); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
	}

	return &types.Session{
		AccessToken: body.AccessToken,
		ExpiresAt:   time.Now().Unix() + body.ExpiresIn,
		// ExpiresAt:   time.Now().Add(time.Duration(body.ExpiresIn) * time.Second),
		Scopes:      strings.Split(body.Scope, " "),
	}, nil
}

// SimSwapClient is the main client for SIM swap operations
type SimSwapClient struct {
	settings types.GlideSdkSettings
}

// NewSimSwapClient creates a new SimSwapClient
func NewSimSwapClient(settings types.GlideSdkSettings) *SimSwapClient {
	return &SimSwapClient{settings: settings}
}

// For creates a SimSwapUserClient for a specific user
func (c *SimSwapClient) For(identifier types.UserIdentifier) (*SimSwapUserClient, error) {
	client := NewSimSwapUserClient(c.settings, identifier)
	err := client.StartSession()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *SimSwapClient) GetHello() (string) {
	return "Hello"
}
