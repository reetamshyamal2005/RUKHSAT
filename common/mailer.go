package common

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends an HTML email via Gmail SMTP using an App Password
func SendEmail(to, subject, htmlBody string) error {
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	if smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP credentials not configured in environment variables (SMTP_USER/SMTP_PASS)")
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587" // Port 587 is standard for STARTTLS

	// SMTP authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// Compose the MIME email header and body
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, htmlBody))

	// Send mail via STARTTLS
	addr := smtpHost + ":" + smtpPort
	return smtp.SendMail(addr, auth, smtpUser, []string{to}, msg)
}

// SendEmailWithAttachment sends an HTML email with a PDF/Image attachment via Gmail SMTP
func SendEmailWithAttachment(to, subject, htmlBody string, attachmentBytes []byte, attachmentName string) error {
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	if smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP credentials not configured in environment variables (SMTP_USER/SMTP_PASS)")
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "465" // Port 465 is standard for SSL/TLS connections

	// Setup SMTP Auth
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// Define boundary for multipart message
	boundary := "RUKHSAT_BOUNDARY_STRING_12345"

	// Compose MIME headers
	header := ""
	header += fmt.Sprintf("To: %s\r\n", to)
	header += fmt.Sprintf("Subject: %s\r\n", subject)
	header += "MIME-Version: 1.0\r\n"
	header += fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary)
	header += "\r\n"

	// Compose HTML body part
	body := fmt.Sprintf("--%s\r\n", boundary)
	body += "Content-Type: text/html; charset=UTF-8\r\n"
	body += "\r\n"
	body += htmlBody + "\r\n"
	body += "\r\n"

	// Compose Attachment part
	body += fmt.Sprintf("--%s\r\n", boundary)
	body += "Content-Type: application/octet-stream\r\n"
	body += "Content-Transfer-Encoding: base64\r\n"
	body += fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachmentName)
	body += "\r\n"

	// Base64 encode the attachment
	base64EncodedAttachment := encodeBase64(attachmentBytes)
	
	// Add base64 lines (rfc 2045 specifies 76 character line length limit)
	for i := 0; i < len(base64EncodedAttachment); i += 76 {
		end := i + 76
		if end > len(base64EncodedAttachment) {
			end = len(base64EncodedAttachment)
		}
		body += base64EncodedAttachment[i:end] + "\r\n"
	}
	body += "\r\n"
	body += fmt.Sprintf("--%s--\r\n", boundary)

	// Assemble message
	msg := []byte(header + body)

	// Since we are using port 465, we must use a TLS Dial connection
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpHost,
	}

	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, tlsconfig)
	if err != nil {
		return fmt.Errorf("failed to dial TLS SMTP server: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %v", err)
	}

	if err = client.Mail(smtpUser); err != nil {
		return fmt.Errorf("SMTP MAIL command failed: %v", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT command failed: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA command failed: %v", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message body: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %v", err)
	}

	return nil
}

// Simple base64 encoder
func encodeBase64(data []byte) string {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	
	lenData := len(data)
	if lenData == 0 {
		return ""
	}
	
	encoded := make([]byte, ((lenData+2)/3)*4)
	var di, si int
	for lenData-si >= 3 {
		val := uint32(data[si])<<16 | uint32(data[si+1])<<8 | uint32(data[si+2])
		encoded[di] = encodeStd[val>>18&0x3F]
		encoded[di+1] = encodeStd[val>>12&0x3F]
		encoded[di+2] = encodeStd[val>>6&0x3F]
		encoded[di+3] = encodeStd[val&0x3F]
		si += 3
		di += 4
	}
	
	if lenData-si == 2 {
		val := uint32(data[si])<<16 | uint32(data[si+1])<<8
		encoded[di] = encodeStd[val>>18&0x3F]
		encoded[di+1] = encodeStd[val>>12&0x3F]
		encoded[di+2] = encodeStd[val>>6&0x3F]
		encoded[di+3] = '='
	} else if lenData-si == 1 {
		val := uint32(data[si])<<16
		encoded[di] = encodeStd[val>>18&0x3F]
		encoded[di+1] = encodeStd[val>>12&0x3F]
		encoded[di+2] = '='
		encoded[di+3] = '='
	}
	
	return string(encoded)
}
