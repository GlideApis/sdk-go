package types

import "time"

// GlideSdkSettings represents the settings for the Glide SDK
type GlideSdkSettings struct {
    ClientID     string
    ClientSecret string
    RedirectURI  string
    UseEnv       bool
    Internal     InternalSettings
}

// InternalSettings represents internal settings for the SDK
type InternalSettings struct {
    AuthBaseURL string
    APIBaseURL  string
}

// Session represents an authentication session
type Session struct {
    AccessToken string
    ExpiresAt   int64
    Scopes      []string
}

// ApiConfig represents the configuration for API calls
type ApiConfig struct {
	SessionIdentifier string
    Session *Session
}

// TelcoFinderNetworkIdResponse represents the response for network ID lookup
type TelcoFinderNetworkIdResponse struct {
    NetworkID string `json:"networkId"`
}

// TelcoFinderSearchResponse represents the response for telco finder search
type TelcoFinderSearchResponse struct {
    Subject    string `json:"subject"`
    Properties struct {
        OperatorID string `json:"operator_Id"`
    } `json:"properties"`
    Links []struct {
        Rel  string `json:"rel"`
        Href string `json:"href"`
    } `json:"links"`
}

// PhoneIdentifier represents a phone number identifier
type PhoneIdentifier struct {
    PhoneNumber string `json:"phoneNumber"`
}

// IpIdentifier represents an IP address identifier
type IpIdentifier struct {
    IPAddress string `json:"ipAddress"`
}

// UserIdIdentifier represents a user ID identifier
type UserIdIdentifier struct {
    UserID string `json:"userId"`
}

// UserIdentifier is an interface that can be satisfied by any of the identifier types
type UserIdentifier interface {
    isUserIdentifier()
}

//magic auth

type MagicAuthStartProps struct {
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
    State      string `json:"state,omitempty"`
    RedirectURL string `json:"redirectURL,omitempty"`
    FallbackChannel NoFallback `json:"fallbackChannel,omitempty"`
}

type NoFallback string
type FallbackVerificationChannel string

const (
    SMS         FallbackVerificationChannel = "SMS"
    EMAIL       FallbackVerificationChannel = "EMAIL"
    NO_FALLBACK FallbackVerificationChannel = "NO_FALLBACK"
)

type MagicAuthVerifyProps struct {
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Code        string `json:"code,omitempty"`
	Token       string `json:"token,omitempty"`
}

// number verify

type NumberVerifyAuthUrlInput struct {
    State *string `json:"state"`
    UseDevNumber string `json:"useDevNumber,omitempty"`
    PrintCode bool `json:"printCode,omitempty"`
}


type NumberVerifyResponse struct {
	DevicePhoneNumberVerified bool
}


type NumberVerifyClientForParams struct {
    Code        string
    PhoneNumber *string
}

//sim swap
type SimSwapCheckParams struct {
	PhoneNumber string
	MaxAge      *int // Pointer to allow nil for undefined
}

type SimSwapRetrieveDateParams struct {
	PhoneNumber string
}

// Implement the UserIdentifier interface for each identifier type
func (PhoneIdentifier) isUserIdentifier()  {}
func (IpIdentifier) isUserIdentifier()     {}
func (UserIdIdentifier) isUserIdentifier() {}


type MetricInfo struct {
	Operator    string    `json:"operator,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	SessionId   string    `json:"sessionId"`
	MetricName  string    `json:"metricName"`
	Api         string    `json:"api"`
	ClientId    string    `json:"clientId"`
}

type TokenData struct {
	Ext struct {
		Operator string `json:"operator"`
	} `json:"ext"`
}

