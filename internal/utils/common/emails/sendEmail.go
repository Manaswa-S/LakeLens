package emails

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
)


func SendEmailHTML(body *bytes.Buffer, to_Email string) {
	
	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	demo := "Subject: LakeLens" + "\n" + headers + "\n\n" + body.String()

	auth := smtp.PlainAuth(
		"",
		os.Getenv("Google_SMTP_Uname"),
		os.Getenv("Google_SMTP_Pass"),
		os.Getenv("Google_SMTP_Host"),
	)

	emails := []string{to_Email}

	


	err := smtp.SendMail(
		os.Getenv("Google_SMTP_HostAddr"),
		auth,
		os.Getenv("Google_SMTP_From"),
		emails,
		[]byte(demo),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
}