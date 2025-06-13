package errpipe

import (
	"encoding/json"
	"fmt"
	configs "lakelens/internal/config"
	"lakelens/internal/dto"
	"os"
	"sync"
)

type ErrorData struct {
	Debug    string // found in ctx.Get("debug")
	Info     string // found in ctx.Get("info")
	Warn     string // found in ctx.Get("warn")
	Error    string // found in ctx.Get("error")
	Critical string // found in ctx.Get("critical")
	Fatal    string // found in ctx.Get("fatal")

	LogData *dto.RequestLogData
}

type ErrorHandler interface {
	HandleError(*ErrorData)
}

type ErrorProcessor struct {
	Handler ErrorHandler
	errChan chan *ErrorData
}

func NewErrorProcessor(handler ErrorHandler) *ErrorProcessor {
	return &ErrorProcessor{
		Handler: handler,
		errChan: make(chan *ErrorData, 50),
	}
}

func (p *ErrorProcessor) ProcessErrors() error {

	errorsLoggerFile, err := os.OpenFile(configs.Paths.ErrorLoggerFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	go func(errorsLoggerFile *os.File) {

		for report := range p.errChan {

			if report == nil {
				continue
			}

			jsonLog, err := json.Marshal(report)
			if err != nil {
				fmt.Printf("Failed to marshal report struct to json : %v \n Report : %+v\n", err.Error(), report)
				continue
			}

			_, err = errorsLoggerFile.WriteString(string(jsonLog) + "\n")
			if err != nil {
				fmt.Printf("Failed to write error log in logger file : %v \n Report : %+v\n", err.Error(), report)
				continue
			}

			p.Handler.HandleError(report)
		}

		defer errorsLoggerFile.Close()
	}(errorsLoggerFile)

	once.Do(func() {
		processor = p
	})

	return nil
}

// TODO: implement the retry mechanism here too.

var (
	processor *ErrorProcessor
	once      sync.Once
)

// NewError sends a new error of format dto.ErrorData to the ErrorsChan
func NewError(report *ErrorData) {
	if report == nil {
		fmt.Println("Tried sending empty report(ErrorData) sent to the report handler.")
		return
	}

	select {
	case processor.errChan <- report:
	default:
		fmt.Printf("Error channel is full, dropping error : %+v", report)
	}
}
