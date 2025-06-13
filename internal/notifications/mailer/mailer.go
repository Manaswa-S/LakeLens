package mailer

import (
	"bytes"
	"fmt"
	"lakelens/cmd/errpipe"
	configs "lakelens/internal/config"
	utils "lakelens/internal/utils/common"
)

type Mailer interface {
	Send(to string, sub string, body *bytes.Buffer) error
}

type Errored struct {
	Email   string
	Subject string
	Body    *bytes.Buffer
}

type EmailService struct {
	Mailer  Mailer
	ErrChan chan *Errored
}

func NewEmailService(mailer Mailer) *EmailService {
	return &EmailService{
		Mailer:  mailer,
		ErrChan: make(chan *Errored, 100),
	}
}

func (e *EmailService) RetryErrored() {

	go func(e *EmailService) {
		// TODO: setup the retry logic
		for err := range e.ErrChan {
			fmt.Println(err.Email)
		}
	}(e)

}

func (e *EmailService) SendResetPassEmail(link, email string) {

	body, err := utils.DynamicHTML(configs.Paths.ResetPassHTMLPath, struct {
		Reset_Link string
		Email      string
	}{
		Reset_Link: link,
		Email:      email,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	sub := "LakeLens Reset Account Password"

	err = e.Mailer.Send(email, sub, body)
	if err != nil {
		fmt.Println(err)

		errpipe.NewError(&errpipe.ErrorData{
			Error: "Failed to send reset pass email: " + err.Error(),
		})

		e.ErrChan <- &Errored{
			Email:   email,
			Subject: sub,
			Body:    body,
		}
	}
}
