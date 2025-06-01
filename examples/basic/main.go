package main

import (
	"fmt"
	"log"

	"github.com/imlargo/gmgo"
)

func main() {

	cfg := &gmgo.Config{
		CredentialsFile: "gmgo_credentials.json",
		TokenFile:       "gmgo_token.json",
	}

	client, err := gmgo.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating gmgo client: %v", err)
	}

	// Example 1: Basic email sending
	email := &gmgo.Email{
		To:      []string{"recipient@example.com"},
		Subject: "Test from Go",
		Body:    "This is a test message from gmgo",
	}

	result, err := client.SendEmail(email)
	if err != nil {
		log.Printf("Error sending email: %v", err)
	} else {
		fmt.Printf("Email sent successfully. ID: %s\n", result.MessageID)
	}

	// Example 2: HTML email with attachments
	emailWithAttachment := &gmgo.Email{
		To:      []string{"recipient@example.com"},
		Subject: "Email with Attachment",
		Body:    "This is a test message from gmgo",
	}

	emailWithAttachment.AttachFile("document.txt", []byte("File content"), "text/plain")

	result, err = client.SendEmail(emailWithAttachment)
	if err != nil {
		log.Printf("Error sending HTML email: %v", err)
	} else {
		fmt.Printf("HTML email sent. ID: %s\n", result.MessageID)
	}
}
