package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type MailgunService struct {
	APIKey  string
	Domain  string
	From    string
	BaseURL string
}

var EmailService *MailgunService

// InitializeEmailService initializes the Mailgun email service
func InitializeEmailService() error {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")
	fromEmail := os.Getenv("MAILGUN_FROM_EMAIL")
	baseURL := os.Getenv("APP_BASE_URL")

	if apiKey == "" || domain == "" || fromEmail == "" {
		return fmt.Errorf("MAILGUN_API_KEY, MAILGUN_DOMAIN, and MAILGUN_FROM_EMAIL must be set")
	}

	if baseURL == "" {
		baseURL = "http://localhost:5173" // Default for development
	}

	EmailService = &MailgunService{
		APIKey:  apiKey,
		Domain:  domain,
		From:    fromEmail,
		BaseURL: baseURL,
	}

	return nil
}

// SendEmail sends an email using Mailgun API
func (mg *MailgunService) SendEmail(to, subject, textBody, htmlBody string) error {
	url := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", mg.Domain)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	writer.WriteField("from", mg.From)
	writer.WriteField("to", to)
	writer.WriteField("subject", subject)
	if textBody != "" {
		writer.WriteField("text", textBody)
	}
	if htmlBody != "" {
		writer.WriteField("html", htmlBody)
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", mg.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailgun API error: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

// SendPasswordResetEmail sends a password reset email
func (mg *MailgunService) SendPasswordResetEmail(to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", mg.BaseURL, token)
	
	subject := "Reset Your Password"
	textBody := fmt.Sprintf(`Hello,

You requested to reset your password. Click the link below to reset it:

%s

This link will expire in 1 hour.

If you didn't request this, please ignore this email.`, resetURL)

	htmlBody := fmt.Sprintf(`<html>
<body>
	<h2>Reset Your Password</h2>
	<p>Hello,</p>
	<p>You requested to reset your password. Click the link below to reset it:</p>
	<p><a href="%s">Reset Password</a></p>
	<p>This link will expire in 1 hour.</p>
	<p>If you didn't request this, please ignore this email.</p>
</body>
</html>`, resetURL)

	return mg.SendEmail(to, subject, textBody, htmlBody)
}

// SendWeeklyReport sends a weekly workout report email
func (mg *MailgunService) SendWeeklyReport(to, reportHTML string) error {
	subject := "Your Weekly Workout Report"
	textBody := "Please view this email in HTML format to see your weekly workout report."
	
	return mg.SendEmail(to, subject, textBody, reportHTML)
}

