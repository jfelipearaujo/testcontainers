package example

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/cucumber/godog"
	"github.com/jfelipearaujo/testcontainers/pkg/container"
	"github.com/jfelipearaujo/testcontainers/pkg/container/localstack"
	"github.com/jfelipearaujo/testcontainers/pkg/state"
	"github.com/jfelipearaujo/testcontainers/pkg/testsuite"
)

type test struct {
	awsEndpoint string

	topicArn       string
	topicMessageId string

	queueUrl       string
	queueMessageId string

	totalMessages int
}

var testState = state.NewState[test]()

var containers = container.NewGroup()

func TestFeatures(t *testing.T) {
	testsuite.NewTestSuite(t,
		initializeScenario,
		testsuite.WithPaths("features"),
		testsuite.WithConcurrency(0),
	)
}

func initializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		os.Setenv("AWS_ACCESS_KEY_ID", "test")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
		os.Setenv("AWS_SESSION_TOKEN", "test")

		definition := container.NewContainerDefinition(
			localstack.WithLocalStackContainer(),
			container.WithEnvVars(map[string]string{
				"SERVICES": "sns,sqs",
			}),
			container.WithExecutableFiles(
				localstack.BasePath,
				"./testdata/init-sns.sh",
				"./testdata/init-sqs.sh",
				"./testdata/z-init.sh",
			),
			container.WithForceWaitDuration(5*time.Second),
		)

		localStackContainer, err := definition.BuildContainer(ctx)
		if err != nil {
			return ctx, err
		}

		awsEndpoint, err := localstack.BuildEndpoint(ctx, localStackContainer)
		if err != nil {
			return ctx, err
		}

		currentState := testState.Retrieve(ctx)
		currentState.awsEndpoint = awsEndpoint

		containers[sc.Id] = container.BuildGroupContainer(
			container.WithDockerContainer(localStackContainer),
		)

		return testState.Enrich(ctx, currentState), nil
	})

	ctx.Step(`^I have an SNS topic named "([^"]*)"$`, iHaveAnSNSTopicNamed)
	ctx.Step(`^I have an SQS queue named "([^"]*)"$`, iHaveAnSQSQueueNamed)
	ctx.Step(`^I publish a message into the topic$`, iPublishAMessageIntoTheTopic)
	ctx.Step(`^I publish a message into the queue$`, iPublishAMessageIntoTheQueue)
	ctx.Step(`^I read the messages from the queue$`, iReadTheMessagesFromTheQueue)
	ctx.Step(`^I should read "([^"]*)" message$`, iShouldReadMessage)
	ctx.Step(`^the message should be published into the topic$`, theMessageShouldBePublishedIntoTheTopic)
	ctx.Step(`^the message should be published into the queue$`, theMessageShouldBePublishedIntoTheQueue)

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		group := containers[sc.Id]

		return container.DestroyGroup(ctx, group)
	})
}

func iHaveAnSNSTopicNamed(ctx context.Context, topicName string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	cloudConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to load AWS default config: %w", err)
	}

	cloudConfig.BaseEndpoint = &currentState.awsEndpoint

	client := sns.NewFromConfig(cloudConfig)

	output, err := client.ListTopics(ctx, &sns.ListTopicsInput{})
	if err != nil {
		return ctx, fmt.Errorf("failed to list topics: %w", err)
	}

	expectedArn := fmt.Sprintf("arn:aws:sns:us-east-1:000000000000:%s", topicName)

	topicFound := false
	for _, topic := range output.Topics {
		if *topic.TopicArn == expectedArn {
			topicFound = true
			break
		}
	}

	if !topicFound {
		return ctx, fmt.Errorf("topic '%s' not found", topicName)
	}

	currentState.topicArn = expectedArn

	return testState.Enrich(ctx, currentState), nil
}

func iHaveAnSQSQueueNamed(ctx context.Context, queueName string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	cloudConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to load AWS default config: %w", err)
	}

	cloudConfig.BaseEndpoint = &currentState.awsEndpoint

	client := sqs.NewFromConfig(cloudConfig)

	output, err := client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		return ctx, fmt.Errorf("failed to list queues: %w", err)
	}

	expectedUrl := fmt.Sprintf("http://sqs.us-east-1.localhost.localstack.cloud:4566/000000000000/%s", queueName)

	queueFound := false
	for _, queueUrl := range output.QueueUrls {
		if queueUrl == expectedUrl {
			queueFound = true
			break
		}
	}

	if !queueFound {
		return ctx, fmt.Errorf("queue '%s' not found", queueName)
	}

	currentState.queueUrl = expectedUrl

	return testState.Enrich(ctx, currentState), nil
}

func iPublishAMessageIntoTheTopic(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	cloudConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to load AWS default config: %w", err)
	}

	cloudConfig.BaseEndpoint = &currentState.awsEndpoint

	client := sns.NewFromConfig(cloudConfig)

	output, err := client.Publish(ctx, &sns.PublishInput{
		TopicArn: &currentState.topicArn,
		Message:  aws.String("Hello World!"),
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to publish message: %w", err)
	}

	currentState.topicMessageId = *output.MessageId

	return testState.Enrich(ctx, currentState), nil
}

func iPublishAMessageIntoTheQueue(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	cloudConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to load AWS default config: %w", err)
	}

	cloudConfig.BaseEndpoint = &currentState.awsEndpoint

	client := sqs.NewFromConfig(cloudConfig)

	output, err := client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &currentState.queueUrl,
		MessageBody: aws.String("Hello World!"),
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to publish message: %w", err)
	}

	currentState.queueMessageId = *output.MessageId

	return testState.Enrich(ctx, currentState), nil
}

func iReadTheMessagesFromTheQueue(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	cloudConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to load AWS default config: %w", err)
	}

	cloudConfig.BaseEndpoint = &currentState.awsEndpoint

	client := sqs.NewFromConfig(cloudConfig)

	output, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &currentState.queueUrl,
		MaxNumberOfMessages: 10,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to read messages: %w", err)
	}

	currentState.totalMessages = len(output.Messages)

	return testState.Enrich(ctx, currentState), nil
}

func iShouldReadMessage(ctx context.Context, numOfMessages int) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	if currentState.totalMessages != numOfMessages {
		return ctx, fmt.Errorf("expected %d messages on the queue, but got %d", numOfMessages, currentState.totalMessages)
	}

	return ctx, nil
}

func theMessageShouldBePublishedIntoTheTopic(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	if currentState.topicMessageId == "" {
		return ctx, fmt.Errorf("topic message id not found, but expected to be found")
	}

	return ctx, nil
}

func theMessageShouldBePublishedIntoTheQueue(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	if currentState.queueMessageId == "" {
		return ctx, fmt.Errorf("queue message id not found, but expected to be found")
	}

	return ctx, nil
}
