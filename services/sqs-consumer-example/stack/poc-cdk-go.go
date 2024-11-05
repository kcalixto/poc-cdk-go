package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	STACK_NAME = "poc-cdk-go"
)

type PocCdkGoStackProps struct {
	awscdk.StackProps
}

func bin(assetName string) *string {
	return jsii.String(fmt.Sprintf("../bin/%s", assetName))
}

func NewPocCdkGoStack(scope constructs.Construct, id string, props *PocCdkGoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	queueName := "poc-cdk-go-sample-queue"
	eventsDLQueue := awssqs.NewQueue(stack, jsii.String(fmt.Sprintf("%s-dlq", queueName)), &awssqs.QueueProps{
		QueueName: jsii.String(fmt.Sprintf("%s-dlq", queueName)),
	})
	eventsQueue := awssqs.NewQueue(stack, jsii.String(queueName), &awssqs.QueueProps{
		QueueName:         jsii.String(queueName),
		VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
		DeadLetterQueue: &awssqs.DeadLetterQueue{
			Queue:           eventsDLQueue,
			MaxReceiveCount: jsii.Number(3),
		},
	})

	sqsConsumerFunctionName := jsii.String("poc-cdk-sqs-consumer-function")
	sqsConsumer := awslambda.NewFunction(stack, sqsConsumerFunctionName, &awslambda.FunctionProps{
		FunctionName: sqsConsumerFunctionName,
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code: awslambda.Code_FromAsset(
			bin("handler.zip"),
			nil,
		),
	})

	sqsConsumerDLQFunctionName := jsii.String("poc-cdk-go-sqs-consumer-dlq")
	sqsConsumerDLQ := awslambda.NewFunction(stack, sqsConsumerDLQFunctionName, &awslambda.FunctionProps{
		FunctionName: sqsConsumerDLQFunctionName,
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Code: awslambda.Code_FromAsset(
			bin("handler-dlq.zip"),
			nil,
		),
	})

	sqsConsumer.AddEventSource(
		awslambdaeventsources.NewSqsEventSource(eventsQueue, &awslambdaeventsources.SqsEventSourceProps{
			BatchSize: jsii.Number(5),
		}),
	)
	sqsConsumerDLQ.AddEventSource(
		awslambdaeventsources.NewSqsEventSource(eventsDLQueue, &awslambdaeventsources.SqsEventSourceProps{
			BatchSize: jsii.Number(5),
		}),
	)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewPocCdkGoStack(app, STACK_NAME, &PocCdkGoStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
