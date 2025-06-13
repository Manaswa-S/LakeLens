package errpipe

import (
	"fmt"
	"net/http"
	"net/url"

)

// TelegramErrorHandler is responsible for sending error reports to a specified Telegram chat.
type TelegramErrorHandler struct {
	BotToken string
	ChatID string
}

func NewTelegramErrorHandler(botToken, chatID string) (*TelegramErrorHandler, error) {


	// TODO: check if the requested bot is available and accessible.


	return &TelegramErrorHandler{
		BotToken: botToken,
		ChatID: chatID,
	}, nil
}

// HandleError processes the given error report and sends it to Telegram.
// If the first attempt fails, it retries once with a failure notification.
func (t *TelegramErrorHandler) HandleError(report *ErrorData) {
	
	textToSend := t.reportTextMsg(report)

	result, err := t.postTelegramMsg(textToSend)
	if err != nil {
		textToSend = t.telegramFailedMsg(result, err.Error())
	} 
	if result != nil && result.StatusCode != http.StatusOK {
		textToSend += t.telegramFailedMsg(result, "nil")
	}

	result, err = t.postTelegramMsg(textToSend)
	if err != nil || (result != nil && result.StatusCode != http.StatusOK) {
		fmt.Println("Error: Failed to send message via Telegram. Check errors log file for more details.")
		return
	}
}
// reportTextMsg formats the error report into a readable text message and returns it.
func (t *TelegramErrorHandler) reportTextMsg(report *ErrorData) (string) {

	textToSend := ""

	if report.Debug != "" {
		textToSend += fmt.Sprintf("🔵 Debug: %s\n", report.Debug)
	}
	if report.Info != "" {
		textToSend += fmt.Sprintf("ℹ️ Info: %s\n", report.Info)
	}
	if report.Warn != "" {
		textToSend += fmt.Sprintf("⚠️ Warn: %s\n", report.Warn)
	}
	if report.Error != "" {
		textToSend += fmt.Sprintf("🔴 Error: %s\n", report.Error)
	}
	if report.Critical != "" {
		textToSend += fmt.Sprintf("🔥 Critical: %s\n", report.Critical)
	}
	if report.Fatal != "" {
		textToSend += fmt.Sprintf("☠️ Fatal: %s\n", report.Fatal)
	}

	if report.LogData != nil {
		textToSend += fmt.Sprintf(
			"\n" +
			"Start Time: %s\n" + 
			"Client IP: %s\n" +
			"Method: %s\n" + 
			"Path: %s\n" + 
			"Status Code: %d\n" +
			"Internal Error: %s\n" + 
			"Latency: %d\n", 

			report.LogData.StartTime, report.LogData.ClientIP, report.LogData.Method, report.LogData.Path,
			report.LogData.StatusCode, report.LogData.InternalError, report.LogData.Latency,
		)
	}

	return textToSend
}
// telegramFailedMsg formats a failure notification message if earlier message delivery fails and returns it.
func (t *TelegramErrorHandler) telegramFailedMsg(result *http.Response, errStr string) (string) {

	textToSend := fmt.Sprintf("\n" + 
	"Failed to report message via Telegram. \n" + 
	"Internal Error: %s\n" + 
	"Status Code: %d \n" + 
	"Content Length: %d \n" +
	"Check errors log file for full details." +
	"\n", errStr , result.StatusCode, result.ContentLength)

	return textToSend
}
// postTelegramMsg sends a message to the Telegram API using an HTTP POST request.
// Returns (nil, error) if internal error occurred, (result, nil) otherwise.
func (t *TelegramErrorHandler) postTelegramMsg(textToSend string) (*http.Response, error) {

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	data := url.Values{}
	data.Set("chat_id", t.ChatID)
	data.Set("text", textToSend)
	
	result, err := http.PostForm(apiURL, data)
	if err != nil {
		return nil, err
	}

	return result, nil
}

