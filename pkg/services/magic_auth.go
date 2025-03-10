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

type MagicAuthStartResponse struct {
	Type        string `json:"type"`
	AuthURL     string `json:"authUrl,omitempty"`
	FlatAuthURL string `json:"flatAuthUrl,omitempty"`
	OperatorId  string `json:"operatorId,omitempty"`
}

type MagicAuthCheckResponse struct {
	Verified bool `json:"verified"`
}

type MagicAuthClient struct {
	settings types.GlideSdkSettings
	session  *types.Session
}

func NewMagicAuthClient(settings types.GlideSdkSettings) *MagicAuthClient {
	return &MagicAuthClient{
		settings: settings,
	}
}

func (c *MagicAuthClient) StartAuth(props types.MagicAuthStartProps, conf types.ApiConfig) (*MagicAuthStartResponse, error) {
	var wg sync.WaitGroup
	if c.settings.Internal.APIBaseURL == "" {
		utils.Logger.Error("internal.apiBaseUrl is unset")
		return nil, fmt.Errorf("[GlideClient] internal.apiBaseUrl is unset")
	}
	if conf.SessionIdentifier != "" {
		c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide start", "")
	}

	session, err := c.getSession(conf.Session)
	if err != nil {
		return nil, err
	}

	data := map[string]string{}
	if props.PhoneNumber != "" {
		data["phoneNumber"] = props.PhoneNumber
	} else {
		data["email"] = props.Email
	}
	if props.RedirectURL != "" {
		data["redirectUrl"] = props.RedirectURL
	}
	if props.State != "" {
		data["state"] = props.State
	}
	if props.FallbackChannel != "" {
		data["fallbackChannel"] = string(props.FallbackChannel)
	}
	if props.DeviceIPAddress != "" {
		data["deviceIpAddress"] = props.DeviceIPAddress
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := utils.FetchX(c.settings.Internal.APIBaseURL+"/magic-auth/verification/start", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + session.AccessToken,
		},
		Body: string(jsonData),
	})

	if err != nil {
		utils.Logger.Error("FetchX failed for startAuth: %v", err)
		return nil, fmt.Errorf("[GlideClient]: [magic-auth] FetchX failed for startAuth : %w", err)
	}

	var result MagicAuthStartResponse
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
	}

	if conf.SessionIdentifier != "" && result.OperatorId != "" {
		c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide verificationStartRes", result.OperatorId)
	}
	wg.Wait()
	return &result, nil
}

type MagicAuthVerifyRes struct {
	Verified bool `json:"verified"`
}

func (c *MagicAuthClient) VerifyAuth(props types.MagicAuthVerifyProps, conf types.ApiConfig) (*MagicAuthVerifyRes, error) {
	var wg sync.WaitGroup
	if c.settings.Internal.APIBaseURL == "" {
		return nil, fmt.Errorf("[GlideClient] internal.apiBaseUrl is unset")
	}

	session, err := c.getSession(conf.Session)
	if err != nil {
		return nil, err
	}

	data := map[string]string{}
	if props.PhoneNumber != "" {
		data["phoneNumber"] = props.PhoneNumber
	} else {
		data["email"] = props.Email
	}
	if props.Code != "" {
		data["code"] = props.Code
	} else {
		data["token"] = props.Token
	}
	if props.DeviceIPAddress != "" {
		data["deviceIpAddress"] = props.DeviceIPAddress
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := utils.FetchX(c.settings.Internal.APIBaseURL+"/magic-auth/verification/check", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + session.AccessToken,
		},
		Body: string(jsonData),
	})

	if err != nil {
		return nil, fmt.Errorf("[GlideClient]: [magic-auth] FetchX failed for VerifyAuth : %w", err)
	}

	var result MagicAuthVerifyRes
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response in VerifyAuth: %w", err)
	}

	if conf.SessionIdentifier != "" {
		c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide success", "")
		if result.Verified {
			c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide verified", "")
		} else if !result.Verified {
			c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide unverified", "")
		}
	}
	wg.Wait()
	return &result, nil
}

func (c *MagicAuthClient) getSession(confSession *types.Session) (*types.Session, error) {
	if confSession != nil {
		return confSession, nil
	}

	if c.session != nil && c.session.ExpiresAt > time.Now().Add(time.Minute).Unix() && contains(c.session.Scopes, "magic-auth") {
		utils.Logger.Debug("Using cached session")
		return c.session, nil
	}

	session, err := c.generateNewSession()
	if err != nil {
		return nil, err
	}

	c.session = session
	return session, nil
}

func (c *MagicAuthClient) generateNewSession() (*types.Session, error) {
	if c.settings.ClientID == "" || c.settings.ClientSecret == "" {
		return nil, fmt.Errorf("[GlideClient] Client credentials are required to generate a new session")
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "magic-auth")

	resp, err := utils.FetchX(c.settings.Internal.AuthBaseURL+"/oauth2/token", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(c.settings.ClientID+":"+c.settings.ClientSecret)),
		},
		Body: data.Encode(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate new session: %w", err)
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

func (c *MagicAuthClient) reportMagicAuthMetric(wg *sync.WaitGroup, sessionId, metricName string, operator string) {
	utils.Logger.Debug("reportMagicAuthMetric: %s", metricName)
	metric := types.MetricInfo{
		Operator:   operator,
		Timestamp:  time.Now(),
		SessionId:  sessionId,
		MetricName: metricName,
		Api:        "magic-auth",
		ClientId:   c.settings.ClientID,
	}
	wg.Add(1)
	go func(m types.MetricInfo) {
		defer wg.Done()
		utils.ReportMetric(m)
	}(metric)
}

func (c *MagicAuthClient) GetHello() string {
	return "Hello"
}

func (c *MagicAuthClient) StartServerAuth(props types.MagicAuthStartProps, conf types.ApiConfig) (*types.MagicAuthStartServerAuthResponse, error) {
	var wg sync.WaitGroup
	if c.settings.Internal.APIBaseURL == "" {
		utils.Logger.Error("internal.apiBaseUrl is unset")
		return nil, fmt.Errorf("[GlideClient] internal.apiBaseUrl is unset")
	}
	if conf.SessionIdentifier != "" {
		c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide start server auth", "")
	}

	session, err := c.getSession(conf.Session)
	if err != nil {
		return nil, err
	}

	data := map[string]string{}
	if props.PhoneNumber != "" {
		data["phoneNumber"] = props.PhoneNumber
	} else {
		data["email"] = props.Email
	}
	if props.RedirectURL != "" {
		data["redirectUrl"] = props.RedirectURL
	}
	if props.State != "" {
		data["state"] = props.State
	}
	if props.FallbackChannel != "" {
		data["fallbackChannel"] = string(props.FallbackChannel)
	}
	if props.DeviceIPAddress != "" {
		data["deviceIpAddress"] = props.DeviceIPAddress
	}
	if props.OTPConfirmationURL != "" {
		data["otpConfirmationUrl"] = props.OTPConfirmationURL
	}
	if props.RCSConfirmationURL != "" {
		data["rcsConfirmationUrl"] = props.RCSConfirmationURL
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := utils.FetchX(c.settings.Internal.APIBaseURL+"/magic-auth/verification/start-server-auth", utils.FetchXInput{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + session.AccessToken,
		},
		Body: string(jsonData),
	})

	if err != nil {
		utils.Logger.Error("FetchX failed for startServerAuth: %v", err)
		return nil, fmt.Errorf("[GlideClient]: [magic-auth] FetchX failed for startServerAuth : %w", err)
	}

	var result types.MagicAuthStartServerAuthResponse
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
	}

	if conf.SessionIdentifier != "" {
		c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide serverAuthStartRes", "")
	}
	wg.Wait()
	return &result, nil
}

func (c *MagicAuthClient) CheckServerAuth(sessionID string, conf types.ApiConfig) (*types.MagicAuthCheckServerAuthResponse, error) {
	var wg sync.WaitGroup
	if c.settings.Internal.APIBaseURL == "" {
		return nil, fmt.Errorf("[GlideClient] internal.apiBaseUrl is unset")
	}

	session, err := c.getSession(conf.Session)
	if err != nil {
		return nil, err
	}

	resp, err := utils.FetchX(fmt.Sprintf("%s/magic-auth/verification/check-server-auth?sessionId=%s",
		c.settings.Internal.APIBaseURL, sessionID), utils.FetchXInput{
		Method: "GET",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + session.AccessToken,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("[GlideClient]: [magic-auth] FetchX failed for CheckServerAuth : %w", err)
	}

	var result types.MagicAuthCheckServerAuthResponse
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response in CheckServerAuth: %w", err)
	}

	if conf.SessionIdentifier != "" {
		c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide serverAuthCheck", "")
		if result.Verified {
			c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide serverAuthVerified", "")
		} else {
			c.reportMagicAuthMetric(&wg, conf.SessionIdentifier, "Glide serverAuthUnverified", "")
		}
	}
	wg.Wait()
	return &result, nil
}
