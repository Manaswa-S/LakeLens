package utils

import (
	"bytes"
	"fmt"
	"html/template"
)

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
