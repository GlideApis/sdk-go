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
	LogLevel    LogLevel
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
	Session           *Session
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
	Email              string     `json:"email,omitempty"`
	PhoneNumber        string     `json:"phoneNumber,omitempty"`
	State              string     `json:"state,omitempty"`
	RedirectURL        string     `json:"redirectUrl,omitempty"`
	FallbackChannel    NoFallback `json:"fallbackChannel,omitempty"`
	DeviceIPAddress    string     `json:"deviceIpAddress,omitempty"`
	OTPConfirmationURL string     `json:"otpConfirmationUrl,omitempty"`
	RCSConfirmationURL string     `json:"rcsConfirmationUrl,omitempty"`
}

type NoFallback string
type FallbackVerificationChannel string

const (
	SMS         FallbackVerificationChannel = "SMS"
	EMAIL       FallbackVerificationChannel = "EMAIL"
	NO_FALLBACK FallbackVerificationChannel = "NO_FALLBACK"
)

type MagicAuthVerifyProps struct {
	Email           string `json:"email,omitempty"`
	PhoneNumber     string `json:"phoneNumber,omitempty"`
	Code            string `json:"code,omitempty"`
	Token           string `json:"token,omitempty"`
	DeviceIPAddress string `json:"deviceIpAddress,omitempty"`
}

type MagicAuthStartServerAuthResponse struct {
	SessionID string `json:"sessionId"`
	AuthURL   string `json:"authUrl"`
}

type MagicAuthCheckServerAuthResponse struct {
	Status   string `json:"status"` // "PENDING" or "COMPLETED"
	Verified bool   `json:"verified"`
}

// number verify

type NumberVerifyAuthUrlInput struct {
	State        *string `json:"state"`
	UseDevNumber string  `json:"useDevNumber,omitempty"`
	PrintCode    bool    `json:"printCode,omitempty"`
}

type NumberVerifyResponse struct {
	DevicePhoneNumberVerified bool
}

type NumberVerifyClientForParams struct {
	Code        string
	PhoneNumber *string
}

// sim swap
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
	Operator   string    `json:"operator,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	SessionId  string    `json:"sessionId"`
	MetricName string    `json:"metricName"`
	Api        string    `json:"api"`
	ClientId   string    `json:"clientId"`
}

type TokenData struct {
	Ext struct {
		Operator string `json:"operator"`
	} `json:"ext"`
}

// Add LogLevel type if not already defined
type LogLevel int

const (
	UNSET LogLevel = iota // 0 - represents unset value
	DEBUG                 // 1 - Most verbose
	INFO                  // 2
	WARN                  // 3
	ERROR                 // 4 - Least verbose
)

// KYC Match types
type KYCMatchProps struct {
	PhoneNumber          string `json:"phoneNumber"`
	IDDocument           string `json:"idDocument"`
	Name                 string `json:"name"`
	GivenName            string `json:"givenName"`
	FamilyName           string `json:"familyName"`
	NameKanaHankaku      string `json:"nameKanaHankaku"`
	NameKanaZenkaku      string `json:"nameKanaZenkaku"`
	MiddleNames          string `json:"middleNames"`
	FamilyNameAtBirth    string `json:"familyNameAtBirth"`
	Address              string `json:"address"`
	StreetName           string `json:"streetName"`
	StreetNumber         int    `json:"streetNumber"`
	PostalCode           int    `json:"postalCode"`
	Region               string `json:"region"`
	Locality             string `json:"locality"`
	Country              string `json:"country"`
	HouseNumberExtension string `json:"houseNumberExtension"`
	Birthdate            string `json:"birthdate"`
	Email                string `json:"email"`
	Gender               string `json:"gender"`
}

type KYCMatchResponse struct {
	IDDocumentMatch           *string `json:"idDocumentMatch"`
	NameMatch                 *string `json:"nameMatch"`
	GivenNameMatch            *string `json:"givenNameMatch"`
	FamilyNameMatch           *string `json:"familyNameMatch"`
	NameKanaHankakuMatch      *string `json:"nameKanaHankakuMatch"`
	NameKanaZenkakuMatch      *string `json:"nameKanaZenkakuMatch"`
	MiddleNamesMatch          *string `json:"middleNamesMatch"`
	FamilyNameAtBirthMatch    *string `json:"familyNameAtBirthMatch"`
	AddressMatch              *string `json:"addressMatch"`
	StreetNameMatch           *string `json:"streetNameMatch"`
	StreetNumberMatch         *string `json:"streetNumberMatch"`
	PostalCodeMatch           *string `json:"postalCodeMatch"`
	RegionMatch               *string `json:"regionMatch"`
	LocalityMatch             *string `json:"localityMatch"`
	CountryMatch              *string `json:"countryMatch"`
	HouseNumberExtensionMatch *string `json:"houseNumberExtensionMatch"`
	BirthdateMatch            *string `json:"birthdateMatch"`
	EmailMatch                *string `json:"emailMatch"`
	GenderMatch               *string `json:"genderMatch"`
}
