package main

import (
	"context"
	"log"
	"os"
	"rift/commander"
)

func main() {
	ctx := context.Background()
	log.SetFlags(0)

	if os.Getenv("DEBUG") == "debug" {
		log.SetFlags(log.Lshortfile)
	}

	sub := os.Args[1]

	var cmd commander.Commander

	switch sub {
	case "cloudwatch":
		cmd = commander.NewCloudWatch()
	case "sqs":
		cmd = commander.NewSQSWatch()
	default:
		Help()
		return
	}

	err := cmd.Parse(ctx, os.Args[2:]...)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	err = cmd.Run(ctx)
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func Help() {
	helpString := `Usage of rift:

cloudwatch [options]
	get logs from aws cloudwatch

sqs [options]
	get logs from sqs queue. this might change message visibility
	`
	log.Println(helpString)
}
