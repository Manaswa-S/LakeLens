package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"main.go/cmd/server"
)		


func main() {

	flowChan := make(chan os.Signal, 1)
	signal.Notify(flowChan, syscall.SIGINT, syscall.SIGTERM)
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = server.InitHTTPServer()
	if err != nil {
		fmt.Println(err)
		return
	}

	<-flowChan
	fmt.Println("shutting down")
}



