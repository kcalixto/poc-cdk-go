package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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

	dirName, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	lambdaZipS3Asset := awss3assets.NewAsset(stack, jsii.String("ProducerZippedDirAsset"), &awss3assets.AssetProps{
		Path: jsii.String(filepath.Join(dirName, "../bin/producer.zip")),
	})

	// new cloudfromation output
	awscdk.NewCfnOutput(stack, jsii.String("ProducerZippedDirAssetS3Bucket"), &awscdk.CfnOutputProps{
		Value: lambdaZipS3Asset.S3ObjectUrl(),
	})

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
		Environment: &map[string]*string{
			"TEST": getSSM("/poc-cdk-go/test"),
		},
	})
	// newLambdaRole := func(scope constructs.Construct, id string) awsiam.Role {
	// 	// Create a new IAM role for the Lambda function
	// 	role := awsiam.NewRole(scope, jsii.String(id), &awsiam.RoleProps{
	// 		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	// 	})
	//
	// 	// Attach a policy to the role
	// 	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
	// 		Actions:   jsii.Strings("lambda:InvokeFunction"),
	// 		Resources: jsii.Strings("*"),
	// 	}))
	//
	// 	return role
	// }
	// // lambda permissions
	// producerRole := newLambdaRole(stack, "ProducerLambdaRole")
	sqsProducerFunction.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("lambda:InvokeFunction"),
		Resources: jsii.Strings("*"),
	}))

	// api
	apiName := name("producer-api")
	api := awsapigateway.NewRestApi(stack, apiName, &awsapigateway.RestApiProps{
		RestApiName: apiName,
		// DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
		// 	AllowOrigins: jsii.PtrSlice("*"),
		// 	AllowMethods: jsii.PtrSlice(http.MethodGet, http.MethodPost),
		// 	AllowHeaders: jsii.PtrSlice("*"),
		// },
		DeployOptions: &awsapigateway.StageOptions{
			StageName: jsii.String(ENV),
		},
	})

	// Add a '/produce' resource with a GET method
	produceResource := api.Root().ResourceForPath(jsii.String("/produce/message"))
	produceResource.AddMethod(
		jsii.String(http.MethodPost),
		awsapigateway.NewLambdaIntegration(
			sqsProducerFunction,
			&awsapigateway.LambdaIntegrationOptions{
				AllowTestInvoke: jsii.Bool(false),
			},
		),
		&awsapigateway.MethodOptions{},
	)

	// anther lambda
	sqsProducer2FunctionName := name("producer-2-function")
	sqsProducer2Function := awslambda.NewFunction(stack, sqsProducer2FunctionName, &awslambda.FunctionProps{
		FunctionName: sqsProducer2FunctionName,
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Architecture: awslambda.Architecture_ARM_64(),
		Handler:      jsii.String("bootstrap"),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		MemorySize:   jsii.Number(256),
		Code: awslambda.Code_FromAsset(
			bin("producer.zip"),
			nil,
		),
		Environment: &map[string]*string{
			"TEST": getSSM("/poc-cdk-go/test"),
		},
	})

	producer2Resource := api.Root().ResourceForPath(jsii.String("/produce/message2"))
	producer2Resource.AddMethod(
		jsii.String(http.MethodPost),
		awsapigateway.NewLambdaIntegration(
			sqsProducer2Function,
			&awsapigateway.LambdaIntegrationOptions{
				AllowTestInvoke: jsii.Bool(false),
			},
		),
		&awsapigateway.MethodOptions{},
	)

	return stack
}

func getSSM(key string) *string {
	awsconfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	ssmClient := ssm.NewFromConfig(awsconfig)

	param, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           jsii.String(key),
		WithDecryption: jsii.Bool(true),
	})
	if err != nil {
		panic(err)
	}

	return param.Parameter.Value
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
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
