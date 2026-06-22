package services

import (
	"fmt"

	"github.com/goravel/framework/facades"
	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client      *resend.Client
	fromAddress string
	fromName    string
	frontendURL string
}

func NewEmailService() *EmailService {
	apiKey := facades.Config().GetString("mail.resend_api_key", "")
	fromAddress := facades.Config().GetString("mail.from.address", "noreply@jobbin.app")
	fromName := facades.Config().GetString("mail.from.name", "Jobbin")
	frontendURL := facades.Config().GetString("app.frontend_url", "http://localhost:5173")

	return &EmailService{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
		fromName:    fromName,
		frontendURL: frontendURL,
	}
}

// SendVerificationEmail kirim email verifikasi ke user baru
func (s *EmailService) SendVerificationEmail(toEmail, toName, token string) error {
	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", s.frontendURL, token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: 'Space Grotesk', Arial, sans-serif; background: #FFD600; margin: 0; padding: 40px 20px; }
    .card { background: #fff; border: 3px solid #000; box-shadow: 6px 6px 0 #000; max-width: 480px; margin: 0 auto; padding: 32px; }
    h1 { font-size: 28px; font-weight: 900; margin: 0 0 8px; }
    p { font-size: 15px; line-height: 1.6; color: #1a1a1a; }
    .btn { display: inline-block; background: #FFD600; color: #000; font-weight: 700; font-size: 14px; padding: 12px 24px; border: 2px solid #000; box-shadow: 4px 4px 0 #000; text-decoration: none; margin: 16px 0; }
    .footer { font-size: 12px; color: #6b6b6b; margin-top: 24px; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Welcome to Jobbin! 🎯</h1>
    <p>Hi %s,</p>
    <p>Thanks for signing up! Please verify your email address to start tracking your job applications.</p>
    <a href="%s" class="btn">VERIFY EMAIL →</a>
    <p>Or copy this link:</p>
    <p style="word-break:break-all; font-size:13px;">%s</p>
    <div class="footer">
      <p>This link expires in 24 hours. If you didn't create an account, you can ignore this email.</p>
    </div>
  </div>
</body>
</html>`, toName, verifyLink, verifyLink)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromAddress),
		To:      []string{toEmail},
		Subject: "Verify your Jobbin account",
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	return err
}

// SendReminderEmail kirim email reminder lamaran
func (s *EmailService) SendReminderEmail(toEmail, toName, jobTitle, company, reminderType string) error {
	subject := fmt.Sprintf("Reminder: Follow up on %s at %s", jobTitle, company)
	urgency := "tomorrow"
	if reminderType == "day_of" {
		urgency = "TODAY"
		subject = fmt.Sprintf("TODAY: Follow up on %s at %s", jobTitle, company)
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: 'Space Grotesk', Arial, sans-serif; background: #FFD600; margin: 0; padding: 40px 20px; }
    .card { background: #fff; border: 3px solid #000; box-shadow: 6px 6px 0 #000; max-width: 480px; margin: 0 auto; padding: 32px; }
    h1 { font-size: 24px; font-weight: 900; margin: 0 0 8px; }
    .badge { display: inline-block; background: %s; color: #000; font-weight: 700; font-size: 11px; padding: 3px 10px; border: 1.5px solid #000; margin-bottom: 16px; }
    p { font-size: 15px; line-height: 1.6; color: #1a1a1a; }
    .job-card { background: #FFF176; border: 2px solid #000; box-shadow: 4px 4px 0 #000; padding: 16px; margin: 16px 0; }
    .btn { display: inline-block; background: #FFD600; color: #000; font-weight: 700; font-size: 14px; padding: 12px 24px; border: 2px solid #000; box-shadow: 4px 4px 0 #000; text-decoration: none; margin: 16px 0; }
    .footer { font-size: 12px; color: #6b6b6b; margin-top: 24px; }
  </style>
</head>
<body>
  <div class="card">
    <span class="badge">⏰ %s</span>
    <h1>Don't forget to follow up!</h1>
    <p>Hi %s,</p>
    <p>You have a job application reminder for %s:</p>
    <div class="job-card">
      <strong style="font-size:16px;">%s</strong><br>
      <span style="font-size:14px;">%s</span>
    </div>
    <a href="%s/board" class="btn">VIEW ON JOBBIN →</a>
    <div class="footer">
      <p>You're receiving this because you set a reminder in Jobbin. Good luck! 🚀</p>
    </div>
  </div>
</body>
</html>`,
		map[string]string{"day_of": "#EF9A9A", "day_before": "#FFD600"}[reminderType],
		urgency, toName, urgency, jobTitle, company, s.frontendURL,
	)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromAddress),
		To:      []string{toEmail},
		Subject: subject,
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	return err
}
