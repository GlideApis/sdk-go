package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"github.com/ClearBlockchain/sdk-go/pkg/types"
	"github.com/ClearBlockchain/sdk-go/pkg/utils"
)

type TelcoFinderClient struct {
	settings types.GlideSdkSettings
	session  *types.Session
}

func NewTelcoFinderClient(settings types.GlideSdkSettings) *TelcoFinderClient {
	return &TelcoFinderClient{
		settings: settings,
	}
}

// NetworkIdForNumber resolves the network ID for a given phone number
func (c *TelcoFinderClient) NetworkIdForNumber(phoneNumber string, conf types.ApiConfig) (*types.TelcoFinderNetworkIdResponse, error) {
	if c.settings.Internal.APIBaseURL == "" {
		return nil, fmt.Errorf("[GlideClient] internal.apiBaseUrl is unset")
	}

    session, err := c.getSession(conf.Session)
    if err != nil {
        return nil, fmt.Errorf("[GlideClient] Failed to get session: %w", err)
    }
    fmt.Printf("Debug: Using session with AccessToken: %s...\n", session.AccessToken)

	body, err := json.Marshal(map[string]string{
		"phoneNumber": utils.FormatPhoneNumber(phoneNumber),
	})
	 if err != nil {
            return nil, fmt.Errorf("[GlideClient] Failed to marshal request body: %w", err)
     }

    fmt.Printf("Debug: Fetching network ID for number: %s...\n", phoneNumber)
    fmt.Printf("Debug: Request body: %s\n", body)
    fmt.Printf("Debug: APIBaseURL: %s\n", c.settings.Internal.APIBaseURL+"/telco-finder/v1/resolve-network-id")
	resp, err := utils.FetchX(c.settings.Internal.APIBaseURL+"/telco-finder/v1/resolve-network-id", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + session.AccessToken,
		},
		Body: string(body),
	})
	if err != nil {
            return nil, fmt.Errorf("[GlideClient] FetchX failed for getting Network ID not found for number: %w", err)
    }

	var result types.TelcoFinderNetworkIdResponse
	if err := resp.JSON(&result); err != nil {
            return nil, fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
    }

	return &result, nil
}

// LookupIp looks up telco information for an IP address
func (c *TelcoFinderClient) LookupIp(ip string, conf types.ApiConfig) (*types.TelcoFinderSearchResponse, error) {
	return c.lookup(fmt.Sprintf("ipport:%s", ip), conf)
}

// LookupNumber looks up telco information for a phone number
func (c *TelcoFinderClient) LookupNumber(phoneNumber string, conf types.ApiConfig) (*types.TelcoFinderSearchResponse, error) {
	return c.lookup(fmt.Sprintf("tel:%s", utils.FormatPhoneNumber(phoneNumber)), conf)
}

func (c *TelcoFinderClient) lookup(subject string, conf types.ApiConfig) (*types.TelcoFinderSearchResponse, error) {
	if c.settings.Internal.APIBaseURL == "" {
		return nil, fmt.Errorf("[GlideClient] internal.apiBaseUrl is unset")
	}

	session, err := c.getSession(conf.Session)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(map[string]string{
		"resource": subject,
	})
	if err != nil {
		return nil, err
	}

	resp, err := utils.FetchX(c.settings.Internal.APIBaseURL+"/telco-finder/v1/search", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + session.AccessToken,
		},
		Body: string(body),
	})
	if err != nil {
		if fetchErr, ok := err.(*utils.FetchError); ok && fetchErr.Response.StatusCode == 404 {
			return nil, fmt.Errorf("[GlideClient] Lookup failed for subject %s", subject)
		}
		return nil, err
	}

	var result types.TelcoFinderSearchResponse
	if err := resp.JSON(&result); err != nil {
		return nil, err
	}

	result.Properties.OperatorID = result.Properties.OperatorID // Assuming the field name is already correct in Go struct

	return &result, nil
}

func (c *TelcoFinderClient) getSession(confSession *types.Session) (*types.Session, error) {
    if confSession != nil {
        fmt.Println("Debug: Using provided session")
        return confSession, nil
    }

    if c.session != nil && c.session.ExpiresAt > time.Now().Add(time.Minute).Unix() && contains(c.session.Scopes, "telco-finder") {
        fmt.Println("Debug: Using cached session")
        return c.session, nil
    }

    fmt.Println("Debug: Generating new session")
    session, err := c.generateNewSession()
    if err != nil {
        return nil, fmt.Errorf("failed to generate new session: %w", err)
    }

    c.session = session
    return session, nil
}

func (c *TelcoFinderClient) generateNewSession() (*types.Session, error) {
	if c.settings.ClientID == "" || c.settings.ClientSecret == "" {
		return nil, fmt.Errorf("[GlideClient] Client credentials are required to generate a new session")
	}

	basicAuth := base64.StdEncoding.EncodeToString([]byte(c.settings.ClientID + ":" + c.settings.ClientSecret))

	resp, err := utils.FetchX(c.settings.Internal.AuthBaseURL+"/oauth2/token", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": "Basic " + basicAuth,
		},
		Body: url.Values{
			"grant_type": {"client_credentials"},
			"scope":      {"telco-finder"},
		}.Encode(),
	})
	if err != nil {
		if fetchErr, ok := err.(*utils.FetchError); ok {
			if fetchErr.Response.StatusCode == 401 {
				return nil, fmt.Errorf("[GlideClient] Invalid client credentials")
			} else if fetchErr.Response.StatusCode == 400 {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(fetchErr.Data), &data); err == nil {
					if data["error"] == "invalid_scope" {
						return nil, fmt.Errorf("[GlideClient] Client does not have required scopes to access this method")
					}
				}
				return nil, fmt.Errorf("[GlideClient] Invalid request")
			}
		}
		return nil, err
	}

	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := resp.JSON(&body); err != nil {
		return nil, err
	}

	return &types.Session{
		AccessToken: body.AccessToken,
		ExpiresAt:   time.Now().Unix() + body.ExpiresIn,
		Scopes:      strings.Split(body.Scope, " "),
	}, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (c *TelcoFinderClient) GetHello() (string) {
	return "Hello"
}
