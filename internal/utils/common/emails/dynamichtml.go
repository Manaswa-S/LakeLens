package emails

import (
	"bytes"
	"fmt"
	"html/template"
	configs "lakelens/internal/config"
)

type resetPass struct {
	Reset_Link string `json:"Reset_Link"`
	Email      string `json:"Email"`
}

// TODO: errors should be handled internally.
// TODO: should run in a go routine, async
func ResetPassEmail(link, email string) {

	body, err := DynamicHTML(configs.Internal.Paths.ResetPassHTMLPath, resetPass{
		Reset_Link: link,
		Email:      email,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	SendEmailHTML(body, email)
}

func DynamicHTML(pathToHTML string, data any) (*bytes.Buffer, error) {

	body := new(bytes.Buffer)

	bodytemplate, err := template.ParseFiles(pathToHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html template : %v", err)
	}

	err = bodytemplate.Execute(body, data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute html template : %v", err)
	}

	return body, nil
}
