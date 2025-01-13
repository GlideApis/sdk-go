# Glide Go SDK

### Introduction
The Glide Go SDK (sdk-go) provides a convenient way to integrate Glide's services into your Go applications. It supports various Glide services, including:

Magic Authentication: Implement easy safe authentication via magic links sent to users' devices.
SIM Swap Detection: Detect recent SIM swaps to prevent fraud.
Number Verification: Verify phone numbers and retrieve operator information.

### Installation
To install the Glide Go SDK, use the go get command:

```bash
go get github.com/ClearBlockchain/sdk-go
```
## Getting Started

### Prerequisites

* Go 1.18 or later.
* Glide API Credentials: You'll need your Glide Client ID and Client Secret.
* Environment Variables: Set up your Glide credentials and configuration.

### Setting Up Environment Variables

Create a .env file in your project root or set environment variables in your system.

```bash
GLIDE_CLIENT_ID=your-client-id
GLIDE_CLIENT_SECRET=your-client-secret
GLIDE_REDIRECT_URI=your-redirect-uri
GLIDE_AUTH_BASE_URL=glide auth base url
GLIDE_API_BASE_URL=glide api base url
```

### Initializing the Glide Client

```go
package main

import (
    "log"
    "os"
    "github.com/joho/godotenv"
    "github.com/ClearBlockchain/sdk-go/pkg/glide"
    "github.com/ClearBlockchain/sdk-go/pkg/types"
)

func main() {
    // Load environment variables from .env file if exists
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found or error reading .env file")
    }
    settings := types.GlideSdkSettings{
        ClientID:     os.Getenv("GLIDE_CLIENT_ID"),
        ClientSecret: os.Getenv("GLIDE_CLIENT_SECRET"),
        RedirectURI:  os.Getenv("GLIDE_REDIRECT_URI"),
        Internal: types.InternalSettings{
            AuthBaseURL: os.Getenv("GLIDE_AUTH_BASE_URL"),
            APIBaseURL:  os.Getenv("GLIDE_API_BASE_URL"),
        },
    }
    glideClient, err := glide.NewGlideClient(settings)
    if err != nil {
        log.Fatalf("Failed to create Glide client: %v", err)
    }
    // Use glideClient to interact with Glide services
}
```


**To view the documents and usage examples please vist: https://docs.glideapi.com/**


