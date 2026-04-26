package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
)

// EmailService handles sending emails
type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("SMTP_FROM_EMAIL"),
		fromName:     os.Getenv("SMTP_FROM_NAME"),
	}
}

// CertificateEmailData contains data for certificate email
type CertificateEmailData struct {
	StudentName    string
	UniversityName string
	CourseName     string
	IssueDate      string
	CertificateID  string
	VerifyURL      string
	TransactionURL string
	TransactionID  string
}

// SendCertificateEmail sends a certificate to a student
func (s *EmailService) SendCertificateEmail(
	toEmail string,
	studentName string,
	data CertificateEmailData,
	pdfPath string,
) error {

	log.Printf("📧 Sending certificate to: %s (%s)", studentName, toEmail)

	// Create email body
	htmlBody, err := s.generateCertificateHTML(data)
	if err != nil {
		return fmt.Errorf("failed to generate email HTML: %v", err)
	}

	// Read PDF file
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to read PDF: %v", err)
	}

	// Send email with attachment
	subject := fmt.Sprintf("Your Certificate from %s", data.UniversityName)

	if err := s.sendEmailWithAttachment(
		toEmail,
		subject,
		htmlBody,
		"certificate.pdf",
		pdfData,
	); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("✅ Certificate sent to: %s", toEmail)
	return nil
}

// generateCertificateHTML generates the HTML email body
func (s *EmailService) generateCertificateHTML(data CertificateEmailData) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: 'Arial', sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background: linear-gradient(135deg, #2563eb 0%, #7c3aed 100%);
            color: white;
            padding: 30px;
            text-align: center;
            border-radius: 10px 10px 0 0;
        }
        .content {
            background: #f9fafb;
            padding: 30px;
            border-radius: 0 0 10px 10px;
        }
        .cert-info {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
            border-left: 4px solid #2563eb;
        }
        .cert-info h3 {
            margin-top: 0;
            color: #2563eb;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background: #2563eb;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            margin: 10px 5px;
            font-weight: bold;
        }
        .button:hover {
            background: #1d4ed8;
        }
        .footer {
            text-align: center;
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #e5e7eb;
            color: #6b7280;
            font-size: 14px;
        }
        .blockchain-badge {
            background: #10b981;
            color: white;
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: bold;
            display: inline-block;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎓 Congratulations, {{.StudentName}}!</h1>
        <p>Your certificate has been issued and secured on the blockchain</p>
    </div>
    
    <div class="content">
        <p>Dear {{.StudentName}},</p>
        
        <p>We are pleased to inform you that your certificate has been successfully issued by <strong>{{.UniversityName}}</strong> and secured on the Cardano blockchain.</p>
        
        <div class="cert-info">
            <h3>📜 Certificate Details</h3>
            <p><strong>Course:</strong> {{.CourseName}}</p>
            <p><strong>Issue Date:</strong> {{.IssueDate}}</p>
            <p><strong>Certificate ID:</strong> {{.CertificateID}}</p>
            <div class="blockchain-badge">⛓️ Blockchain Verified</div>
        </div>
        
        <div class="cert-info">
            <h3>🔗 Blockchain Verification</h3>
            <p>Your certificate has been permanently recorded on the Cardano blockchain:</p>
            <p><strong>Transaction ID:</strong><br>
            <code style="background: #f3f4f6; padding: 4px 8px; border-radius: 4px; font-size: 12px;">{{.TransactionID}}</code></p>
        </div>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.VerifyURL}}" class="button">🔍 Verify Certificate</a>
            <a href="{{.TransactionURL}}" class="button" style="background: #7c3aed;">⛓️ View on Blockchain</a>
        </div>
        
        <p><strong>📎 Your certificate PDF is attached to this email.</strong></p>
        
        <p>You can verify the authenticity of your certificate at any time using the verification link above or by visiting our verification portal and entering your certificate ID.</p>
        
        <p>The blockchain ensures that your certificate is:</p>
        <ul>
            <li>✅ Tamper-proof and immutable</li>
            <li>✅ Independently verifiable</li>
            <li>✅ Permanently stored</li>
            <li>✅ Globally accessible</li>
        </ul>
    </div>
    
    <div class="footer">
        <p>This is an automated message from {{.UniversityName}}</p>
        <p>© {{.UniversityName}} - Powered by Blockchain Technology</p>
    </div>
</body>
</html>
`

	t, err := template.New("certificate").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// sendEmailWithAttachment sends an email with a PDF attachment
func (s *EmailService) sendEmailWithAttachment(
	to string,
	subject string,
	htmlBody string,
	attachmentName string,
	attachmentData []byte,
) error {

	// Prepare email headers and body
	boundary := "boundary-certificate-email"

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%s", boundary)

	// Build message
	var message bytes.Buffer

	// Headers
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")

	// HTML body
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	// PDF attachment
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: application/pdf\r\n")
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", attachmentName))

	// Encode PDF to base64
	encoded := make([]byte, len(attachmentData)*2)
	base64Encode(encoded, attachmentData)
	message.Write(encoded)
	message.WriteString("\r\n\r\n")

	// End boundary
	message.WriteString(fmt.Sprintf("--%s--", boundary))

	// Send via SMTP
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	return smtp.SendMail(addr, auth, s.fromEmail, []string{to}, message.Bytes())
}

// base64Encode encodes data to base64
func base64Encode(dst, src []byte) {
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	di := 0
	for si := 0; si < len(src); si += 3 {
		// Read 3 bytes
		b1, b2, b3 := src[si], byte(0), byte(0)

		n := len(src) - si
		if n > 1 {
			b2 = src[si+1]
		}
		if n > 2 {
			b3 = src[si+2]
		}

		// Encode to 4 base64 characters
		dst[di] = base64Table[b1>>2]
		dst[di+1] = base64Table[((b1&0x03)<<4)|(b2>>4)]

		if n > 1 {
			dst[di+2] = base64Table[((b2&0x0f)<<2)|(b3>>6)]
		} else {
			dst[di+2] = '='
		}

		if n > 2 {
			dst[di+3] = base64Table[b3&0x3f]
		} else {
			dst[di+3] = '='
		}

		di += 4
	}
}

// SendBulkCertificates sends certificates to multiple students
func (s *EmailService) SendBulkCertificates(
	certificates []struct {
		Email       string
		StudentName string
		PDFPath     string
		EmailData   CertificateEmailData
	},
) (int, []error) {

	successCount := 0
	var errors []error

	log.Printf("📧 Starting bulk email send for %d certificates...", len(certificates))

	for i, cert := range certificates {
		log.Printf("📨 Sending %d/%d...", i+1, len(certificates))

		err := s.SendCertificateEmail(
			cert.Email,
			cert.StudentName,
			cert.EmailData,
			cert.PDFPath,
		)

		if err != nil {
			log.Printf("❌ Failed to send to %s: %v", cert.Email, err)
			errors = append(errors, fmt.Errorf("%s: %v", cert.Email, err))
		} else {
			successCount++
		}
	}

	log.Printf("✅ Bulk send complete: %d/%d successful", successCount, len(certificates))

	return successCount, errors
}

// TestEmailConnection tests the SMTP connection
func (s *EmailService) TestEmailConnection() error {
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Try to send a test email to the sender
	testMessage := []byte("Subject: Email Service Test\r\n\r\nEmail service is working!")

	return smtp.SendMail(addr, auth, s.fromEmail, []string{s.fromEmail}, testMessage)
}
