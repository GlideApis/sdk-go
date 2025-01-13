package main

import (
	"fmt"
	"log"
	"os"
	"github.com/ClearBlockchain/sdk-go/pkg/glide"
	"github.com/ClearBlockchain/sdk-go/pkg/types"
)

func SetupTestEnvironment() types.GlideSdkSettings {
    if os.Getenv("GLIDE_CLIENT_ID") == "" {
		log.Fatal("GLIDE_CLIENT_ID environment variable is not set")
	}
    if os.Getenv("GLIDE_CLIENT_SECRET") == "" {
        log.Fatal("GLIDE_CLIENT_SECRET environment variable is not set")
    }
    if os.Getenv("GLIDE_REDIRECT_URI") == "" {
        fmt.Printf("GLIDE_REDIRECT_URI environment variable is not set")
    }
    if os.Getenv("GLIDE_AUTH_BASE_URL") == "" {
        fmt.Printf("GLIDE_AUTH_BASE_URL environment variable is not set")
    }
    if os.Getenv("GLIDE_API_BASE_URL") == "" {
        fmt.Printf("GLIDE_API_BASE_URL environment variable is not set")
    }
    if os.Getenv("REPORT_METRIC_URL") == "" {
        fmt.Printf("REPORT_METRIC_URL environment variable is not set")
    }
    return types.GlideSdkSettings{
        ClientID:     os.Getenv("GLIDE_CLIENT_ID"),
        ClientSecret: os.Getenv("GLIDE_CLIENT_SECRET"),
        RedirectURI:  os.Getenv("GLIDE_REDIRECT_URI"),
        Internal: types.InternalSettings{
            AuthBaseURL: os.Getenv("GLIDE_AUTH_BASE_URL"),
            APIBaseURL:  os.Getenv("GLIDE_API_BASE_URL"),
        },
    }
}

func main() {
	// Example of how to use the SDK
    fmt.Println("Hello from Glide SDK")
	settings := SetupTestEnvironment()
	glideClient, err := glide.NewGlideClient(settings)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Example SDK usage
	magicHelloRes := glideClient.MagicAuth.GetHello()
	magicHelloRes = magicHelloRes + " from MagicAuth Glide SDK"
	fmt.Printf("Magic Auth Says: %v\n", magicHelloRes)

    telcoFinderRes := glideClient.TelcoFinder.GetHello()
	telcoFinderRes = telcoFinderRes + " from TelcoFinder Glide SDK"
	fmt.Printf("Telco Finder Says: %v\n", telcoFinderRes)

    simSwapRes := glideClient.SimSwap.GetHello()
	simSwapRes = simSwapRes + " from SimSwap Glide SDK"
	fmt.Printf("Sim Swap Says: %v\n", simSwapRes)

    numberVerifyRes := glideClient.NumberVerify.GetHello()
    numberVerifyRes = numberVerifyRes + " from NumberVerify Glide SDK"
    fmt.Printf("Number Verify Says: %v\n", numberVerifyRes)

}
