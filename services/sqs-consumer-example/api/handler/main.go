package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(ctx context.Context, event events.SQSEvent) (res events.SQSBatchItemFailure, err error) {
	for _, record := range event.Records {
		fmt.Printf("Processing message: %s\n", record.Body)
		time.Sleep(1 * time.Second)
	}

	return
}

func main() {
	lambda.Start(Handler)
}
