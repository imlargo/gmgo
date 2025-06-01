# GMGO - Go Gmail Client

GMGO is a robust and easy-to-use Go library designed for sending emails via the Gmail API. It simplifies sending plain text, HTML emails, and handling attachments seamlessly.

## Features

* OAuth 2.0 Authentication with Gmail API
* Sending plain text and HTML emails
* Easy management of attachments
* Simple and intuitive API

## Installation

Install GMGO using Go modules. Ensure you have Go installed on your system.

```bash
go get github.com/imlargo/gmgo
```

## Configuration

Create and configure your credentials and token files:

```json
// gmgo_credentials.json
{
    "installed": {
        "client_id": "your_client_id",
        "project_id": "your_project_id",
        "auth_uri": "https://accounts.google.com/o/oauth2/auth",
        "token_uri": "https://oauth2.googleapis.com/token",
        "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
        "client_secret": "your_client_secret",
        "redirect_uris": [
            "...",
        ]
    }
}
```

Obtain OAuth token:


```go
// Use helper function to get OAuth token, which will guide you through the OAuth 2.0 authorization process.
cfg := gmgo.DefaultConfig()
gmgo.GetOauthToken(cfg)
```

Follow the instructions to complete OAuth 2.0 authorization and save the token.

## Usage/Examples

For a full example, refer to the code in the [example](example/) directory.

### Basic Email Sending

```go
package main

import (
	"fmt"
	"log"
	"github.com/imlargo/gmgo"
)

func main() {
	cfg := gmgo.DefaultConfig()
	
	client, err := gmgo.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating gmgo client: %v", err)
	}

	email := &gmgo.Email{
		To:      []string{"recipient@example.com"},
		Subject: "Test Email",
		Body:    "This is a test email from gmgo",
	}

	result, err := client.SendEmail(email)
	if err != nil {
		log.Fatalf("Error sending email: %v", err)
	}

	fmt.Printf("Email sent successfully. ID: %s\n", result.MessageID)
}
```

### Sending Email with Attachments

```go
email := &gmgo.Email{
	To:      []string{"recipient@example.com"},
	Subject: "Email with Attachment",
	Body:    "Please see the attachment",
}

email.AttachFile("document.txt", []byte("File content"), "text/plain")

result, err := client.SendEmail(email)
if err != nil {
	log.Fatalf("Error sending email with attachment: %v", err)
}

fmt.Printf("Email with attachment sent successfully. ID: %s\n", result.MessageID)
```