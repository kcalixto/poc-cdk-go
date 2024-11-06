package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	// "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

var (
	STACK_NAME = "poc-cdk-go"
	ENV        = os.Getenv("ENV")
	name       = func(s string) *string {
		return jsii.Sprintf("%s-%s-%s", STACK_NAME, ENV, s)
	}
	bin = func(assetName string) *string {
		return jsii.Sprintf("../bin/%s", assetName)
	}
)

type PocCdkGoStackProps struct {
	awscdk.StackProps
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

	sqsConsumerFunctionName := name("consumer-function")
	sqsConsumerFunction := awslambda.NewFunction(stack, sqsConsumerFunctionName, &awslambda.FunctionProps{
		FunctionName: sqsConsumerFunctionName,
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:   jsii.Number(256),
		Code: awslambda.Code_FromAsset(
			bin("handler.zip"),
			nil,
		),
	})

	sqsConsumerDLQFunctionName := name("consumer-dlq")
	sqsConsumerDLQFunction := awslambda.NewFunction(stack, sqsConsumerDLQFunctionName, &awslambda.FunctionProps{
		FunctionName: sqsConsumerDLQFunctionName,
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:   jsii.Number(256),
		Code: awslambda.Code_FromAsset(
			bin("handler-dlq.zip"),
			nil,
		),
	})

	sqsConsumerFunction.AddEventSource(
		awslambdaeventsources.NewSqsEventSource(eventsQueue, &awslambdaeventsources.SqsEventSourceProps{
			BatchSize: jsii.Number(5),
		}),
	)
	sqsConsumerDLQFunction.AddEventSource(
		awslambdaeventsources.NewSqsEventSource(eventsDLQueue, &awslambdaeventsources.SqsEventSourceProps{
			BatchSize: jsii.Number(5),
		}),
	)

	sqsProducerFunctionName := name("producer-function")
	sqsProducerFunction := awslambda.NewFunction(stack, sqsProducerFunctionName, &awslambda.FunctionProps{
		FunctionName: sqsProducerFunctionName,
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:   jsii.Number(256),
		Code: awslambda.Code_FromAsset(
			bin("producer.zip"),
			nil,
		),
	})

	// producerPrincipal := awsiam.NewServicePrincipal(jsii.String("apigateway.amazonaws.com"), &awsiam.ServicePrincipalOpts{})
	// sqsProducerFunction.GrantInvoke(producerPrincipal)

	apiName := name("producer-api")
	api := awsapigateway.NewLambdaRestApi(stack, apiName, &awsapigateway.LambdaRestApiProps{
		RestApiName: apiName,
		Handler:     sqsProducerFunction,
		Proxy:       jsii.Bool(false),
	})

	// Add a '/produce' resource with a GET method
	produceResource := api.Root().AddResource(
		jsii.String("produce"),
		&awsapigateway.ResourceOptions{},
	)
	produceResource.AddMethod(
		jsii.String(http.MethodPost),
		awsapigateway.NewLambdaIntegration(
			sqsProducerFunction,
			&awsapigateway.LambdaIntegrationOptions{},
		),
		&awsapigateway.MethodOptions{},
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
	if ENV == "" {
		panic("ENV is required")
	}
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_ACCOUNT_ID")),
		Region:  jsii.String("sa-east-1"),
	}
}
