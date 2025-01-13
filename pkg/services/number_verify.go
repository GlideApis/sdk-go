package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ClearBlockchain/sdk-go/pkg/types"
	"github.com/ClearBlockchain/sdk-go/pkg/utils"
	"github.com/google/uuid"
)

type NumberVerifyUserClient struct {
    settings types.GlideSdkSettings
	session  *types.Session
	code        string
	phoneNumber *string
}

func NewNumberVerifyUserClient(settings types.GlideSdkSettings, params types.NumberVerifyClientForParams) *NumberVerifyUserClient {
	return &NumberVerifyUserClient{
		settings:    settings,
		code:        params.Code,
		phoneNumber: params.PhoneNumber,
	}
}

func (c *NumberVerifyUserClient) StartSession() error {
	if c.settings.Internal.AuthBaseURL == "" {
		return errors.New("[GlideClient] internal.authBaseUrl is unset")
	}
	if c.settings.ClientID == "" || c.settings.ClientSecret == "" {
		return errors.New("[GlideClient] Client credentials are required to generate a new session")
	}
	if c.code == "" {
		return errors.New("[GlideClient] Code is required to start a session")
	}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", c.code)
    resp, err := utils.FetchX(c.settings.Internal.AuthBaseURL+"/oauth2/token", utils.FetchXInput{
        Method: "POST",
        Headers: map[string]string{
            "Content-Type":  "application/x-www-form-urlencoded",
            "Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(c.settings.ClientID+":"+c.settings.ClientSecret)),
        },
        Body: data.Encode(),
    })
	if err != nil {
		return fmt.Errorf("failed to generate new session: %w", err)
	}
	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64    `json:"expires_in"`
		Scope       string `json:"scope"`
	}

	if err := resp.JSON(&body); err != nil {
	        fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
    		return nil
    }

	c.session = &types.Session{
		AccessToken: body.AccessToken,
		ExpiresAt:   time.Now().Unix() + body.ExpiresIn,
		Scopes:      strings.Split(body.Scope, " "),
	}
	return nil
}

func (c *NumberVerifyUserClient) GetOperator() (string, error) {
    return utils.GetOperator(c.session)
}

func (c *NumberVerifyUserClient) VerifyNumber(number *string, conf types.ApiConfig) (*types.NumberVerifyResponse, error) {
	var wg sync.WaitGroup
	if conf.SessionIdentifier != "" {
		operator, err := utils.GetOperator(c.session)
		if err != nil {
		    fmt.Errorf("cannot report metric since failed to get operator: %w", err)
        }
		c.reportNumberVerifyMetric(&wg, conf.SessionIdentifier, "Glide numberVerify start function", operator)
	}
	if c.session == nil {
		return nil, errors.New("[GlideClient] Session is required to verify a number")
	}

	if c.settings.Internal.APIBaseURL == "" {
		return nil, errors.New("[GlideClient] internal.apiBaseUrl is unset")
	}

	var phoneNumber string
	if number != nil && *number != "" {
		phoneNumber = *number
	} else if c.phoneNumber != nil {
		phoneNumber = *c.phoneNumber
	} else {
		return nil, errors.New("[GlideClient] Phone number is required to verify a number")
	}

	body, err := json.Marshal(map[string]string{"phoneNumber": utils.FormatPhoneNumber(phoneNumber)})
	if err != nil {
		return nil, fmt.Errorf("[GlideClient] failed to marshal payload in number verify: %w", err)
	}

	resp, err := utils.FetchX(c.settings.Internal.APIBaseURL+"/number-verification/verify", utils.FetchXInput{
    		Method: "POST",
    		Headers: map[string]string{
    			"Content-Type":  "application/json",
    			"Authorization": "Bearer " + c.session.AccessToken,
    		},
    		Body: string(body),
    })

	if err != nil {
		return nil, fmt.Errorf("failed to verify number: %w", err)
	}

	var result types.NumberVerifyResponse
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("[GlideClient] Failed to parse response: %w", err)
	}
	// Metric reporting for success/failure
    if conf.SessionIdentifier != "" {
        c.reportNumberVerifyMetric(&wg, conf.SessionIdentifier, "Glide success", "")
        if result.DevicePhoneNumberVerified {
            c.reportNumberVerifyMetric(&wg, conf.SessionIdentifier, "Glide verified", "")
        } else {
            c.reportNumberVerifyMetric(&wg, conf.SessionIdentifier, "Glide unverified", "")
        }
    }
	wg.Wait()
	return &result, nil
}

type NumberVerifyClient struct {
	settings types.GlideSdkSettings
}

func NewNumberVerifyClient(settings types.GlideSdkSettings) *NumberVerifyClient {
	return &NumberVerifyClient{settings: settings}
}

func (c *NumberVerifyClient) GetAuthURL(opts ...types.NumberVerifyAuthUrlInput) (string, error) {
	if c.settings.Internal.AuthBaseURL == "" {
		return "", errors.New("[GlideClient] internal.authBaseUrl is unset")
	}
	if c.settings.ClientID == "" {
		return "", errors.New("[GlideClient] Client id is required to generate an auth url")
	}
	var state string
    if len(opts) > 0 && opts[0].State != nil {
        state = *opts[0].State
    } else {
        state = uuid.New().String()
    }
	nonce := uuid.New().String()
	params := url.Values{}
	params.Set("client_id", c.settings.ClientID)
	params.Set("response_type", "code")
	if c.settings.RedirectURI != "" {
		params.Set("redirect_uri", c.settings.RedirectURI)
	}
	params.Set("scope", "openid")
	params.Set("purpose", "dpv:FraudPreventionAndDetection:number-verification")
	params.Set("state", state)
	params.Set("nonce", nonce)
	params.Set("max_age", "0")
	if len(opts) > 0 && opts[0].UseDevNumber != "" {
		params.Set("login_hint", "tel:"+opts[0].UseDevNumber)
	}
	if len(opts) > 0 && opts[0].PrintCode {
		params.Set("dev_print", "true")
	}
	return c.settings.Internal.AuthBaseURL + "/oauth2/auth?" + params.Encode(), nil
}

func (c *NumberVerifyClient) For(params types.NumberVerifyClientForParams) (*NumberVerifyUserClient, error) {
	client := NewNumberVerifyUserClient(c.settings, params)
	err := client.StartSession()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *NumberVerifyUserClient) reportNumberVerifyMetric(wg *sync.WaitGroup, sessionId, metricName string, operator string) {
	metric := types.MetricInfo{
		Operator:   operator,
		Timestamp:  time.Now(),
		SessionId:  sessionId,
		MetricName: metricName,
		Api:        "number-verify",
		ClientId:   c.settings.ClientID,
	}
	wg.Add(1)
	go func(m types.MetricInfo) {
		defer wg.Done()
		utils.ReportMetric(m)
	}(metric)
}

func (c *NumberVerifyClient) GetHello() (string) {
	return "Hello"
}

