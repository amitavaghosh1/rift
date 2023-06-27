package main

import (
	"context"
	"log"
	"os"
	"rift/commander"
)

func main() {
	ctx := context.Background()
	log.SetFlags(log.Lshortfile)

	sub := os.Args[1]

	var cmd commander.Commander

	switch sub {
	case "cloudwatch":
		cmd = commander.NewCloudWatch()
	case "sqs":
		cmd = commander.NewSQSWatch()
	default:
		cmd = commander.NoopCommander{}
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
