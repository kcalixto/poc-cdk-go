package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func Handler(ctx context.Context, event events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("unable to load SDK config, %s", err.Error()),
		}, nil
	}

	// Create an SQS client
	client := sqs.NewFromConfig(cfg)

	type Request struct {
		Message  string `json:"message"`
		QueueUrl string `json:"queueUrl"`
	}

	var req Request
	err = json.Unmarshal([]byte(event.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	fmt.Printf("Sending message: %s to queue: %s", req.Message, req.QueueUrl)

	// Send the message
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(req.Message),
		QueueUrl:    aws.String(req.QueueUrl),
	}

	result, err := client.SendMessage(ctx, input)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("failed to send message, %s", err.Error()),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("Message sent with ID: %s", *result.MessageId),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
