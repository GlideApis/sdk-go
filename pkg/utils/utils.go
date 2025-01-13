package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ClearBlockchain/sdk-go/pkg/types"
)

// HTTPResponseError represents an HTTP error response
type HTTPResponseError struct {
    Response *http.Response
}

func (e *HTTPResponseError) Error() string {
    return fmt.Sprintf("HTTP Error Response: %d %s", e.Response.StatusCode, e.Response.Status)
}

// InsufficientSessionError represents an error due to insufficient session
type InsufficientSessionError struct {
    Have    *int
    Need    *int
    Message string
}

func (e *InsufficientSessionError) Error() string {
    if e.Message != "" {
        return e.Message
    }
    return "Session is required for this request"
}

// FormatPhoneNumber formats a phone number string
func FormatPhoneNumber(phoneNumber string) string {
    re := regexp.MustCompile("[^0-9]")
    return "+" + re.ReplaceAllString(phoneNumber, "")
}

// FetchError represents an error during fetch operation
type FetchError struct {
    Response *http.Response
    Data     string
}

func (e *FetchError) Error() string {
    return fmt.Sprintf("Fetch Error: %d %s", e.Response.StatusCode, e.Response.Status)
}

// FetchXInput represents input for FetchX function
type FetchXInput struct {
    Method  string
    Headers map[string]string
    Body    string
}

// FetchXResponse represents the response from FetchX function
type FetchXResponse struct {
    Data []byte
    Response *http.Response
}

func (r *FetchXResponse) JSON(v interface{}) error {
    return json.Unmarshal(r.Data, v)
}

func (r *FetchXResponse) Text() string {
    return string(r.Data)
}

func (r *FetchXResponse) OK() bool {
    return r.Response.StatusCode < 400
}

// FetchX performs an HTTP request
func FetchX(url string, input FetchXInput) (*FetchXResponse, error) {
    client := &http.Client{}

    req, err := http.NewRequest(input.Method, url, strings.NewReader(input.Body))
    if err != nil {
        return nil, err
    }

    for k, v := range input.Headers {
        req.Header.Set(k, v)
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode >= 400 {
        return nil, &FetchError{Response: resp, Data: string(data)}
    }

    return &FetchXResponse{Data: data, Response: resp}, nil
}


func GetOperator(session *types.Session) (string, error) {
	if session == nil {
		return "", errors.New("[GlideClient] Session is required to get operator")
	}
	tokenParts := strings.Split(session.AccessToken, ".")
	if len(tokenParts) < 2 {
		return "", errors.New("invalid access token format")
	}
	decodedToken, err := base64.RawStdEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return "unknown", fmt.Errorf("failed to decode token: %v", err)
	}
	var tokenData types.TokenData
	if err := json.Unmarshal(decodedToken, &tokenData); err != nil {
		return "unknown", fmt.Errorf("failed to unmarshal token data: %v", err)
	}
	return tokenData.Ext.Operator, nil
}


func ReportMetric(report types.MetricInfo) {
	reportToServer := map[string]interface{}{
		"sessionId":  report.SessionId,
		"metricName": report.MetricName,
		"timestamp":  report.Timestamp.Format(time.RFC3339), // ISO 8601 format
		"api":        report.Api,
		"clientId":   report.ClientId,
		"operator":   report.Operator,
	}
	url := os.Getenv("REPORT_METRIC_URL")
	if url == "" {
		fmt.Println("missing process env REPORT_METRIC_URL")
		return
	}
	const maxRetries = 3
	attempt := 0
	retryDelay := func(attempt int) time.Duration {
		return time.Duration(1<<attempt) * time.Second // Exponential backoff: 1s, 2s, 4s
	}
	for attempt < maxRetries {
		err := sendMetric(url, reportToServer)
		if err == nil {
			return // Successfully sent the metric
		}
		fmt.Printf("Error reporting to metric server (attempt %d): %v\n", attempt+1, err)
		attempt++
		if attempt < maxRetries {
			time.Sleep(retryDelay(attempt))
		}
	}
	fmt.Println("Failed to report metric after multiple attempts")
}

func sendMetric(url string, data map[string]interface{}) error {
	fmt.Println("Sending metric to: ", url)
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal report data: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	fmt.Println("Response status: ", resp.Status)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK HTTP status: %s", resp.Status)
	}
	return nil
}
