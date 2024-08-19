package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	AwsLocalEndpoint        = "http://localhost:4566"
	AwsLocalCredentialsName = "AwsLocalCredentials" //nolint:gosec
	AwsLocalDefaultRegion   = "us-east-1"
	AwsLocalAccountId       = "000000000000" //nolint:gosec
	AwsLocalAccessKey       = "test"         //nolint:gosec
	AwsLocalSecret          = "test"         //nolint:gosec
)

var (
	ErrorAwsLocalCredentialsEmpty = "awslocal credentials are empty" //nolint:gosec
)

// NewAwsLocalConfig returns a new [aws.Config] object configured to connect to LocalStack.
func NewAwsLocalConfig(ctx context.Context, accountId, endpoint, region, key, secret string, optFns ...func(*config.LoadOptions) error) aws.Config {
	// load the opts
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithCredentialsProvider(NewAwsLocalCredentialsProvider(key, secret, accountId)),
	}
	opts = append(opts, optFns...)

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		panic(err)
	}

	cfg.BaseEndpoint = aws.String(endpoint)
	return cfg
}

// NewDefaultAwsLocalConfig returns a new default [aws.Config] object configured to connect to the default LocalStack.
func NewDefaultAwsLocalConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(AwsLocalDefaultRegion),
		config.WithCredentialsProvider(NewDefaultAwsLocalCredentialsProvider()))
	if err != nil {
		panic(err)
	}
	cfg.BaseEndpoint = aws.String(AwsLocalEndpoint)
	return cfg
}

// NewAwsLocalSQS returns a new [sqs.Client] object configured to connect to LocalStack
func NewAwsLocalSQS(cfg aws.Config) *sqs.Client {
	return sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(AwsLocalEndpoint)
	})
}

// to ensure AwsLocalCredentialsProvider implements the [aws.CredentialsProvider] interface
var _ aws.CredentialsProvider = (*AwsLocalCredentialsProvider)(nil)

// A AwsLocalCredentialsProvider is a static credentials provider that returns the same credentials,
// designed for use with LocalStack
type AwsLocalCredentialsProvider struct {
	Value aws.Credentials
}

// NewDefaultAwsLocalCredentialsProvider returns an AwsLocalCredentialsProvider
// initialized with the default AwsLocal credentials.
func NewDefaultAwsLocalCredentialsProvider() AwsLocalCredentialsProvider {
	return NewAwsLocalCredentialsProvider(AwsLocalAccessKey, AwsLocalSecret, AwsLocalAccountId)
}

// NewAwsLocalCredentialsProvider return a StaticCredentialsProvider initialized with the AWS credentials passed in.
func NewAwsLocalCredentialsProvider(key, secret, accountId string) AwsLocalCredentialsProvider {
	return AwsLocalCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     key,
			SecretAccessKey: secret,
			AccountID:       accountId,
			SessionToken:    "",
			CanExpire:       false,
		},
	}
}

// Retrieve returns the credentials or error if the credentials are invalid.
func (s AwsLocalCredentialsProvider) Retrieve(_ context.Context) (aws.Credentials, error) {
	v := s.Value
	if v.AccessKeyID == "" || v.SecretAccessKey == "" {
		return aws.Credentials{
			Source: AwsLocalCredentialsName,
		}, &AwsLocalCredentialsEmptyError{}
	}

	if len(v.Source) == 0 {
		v.Source = AwsLocalCredentialsName
	}

	return v, nil
}

func (s AwsLocalCredentialsProvider) IsExpired() bool {
	return false
}

// AwsLocalCredentialsEmptyError is emitted when the AwsLocal credentials are empty.
type AwsLocalCredentialsEmptyError struct{}

func (*AwsLocalCredentialsEmptyError) Error() string {
	return ErrorAwsLocalCredentialsEmpty
}

func main() {
	// build the client
	sqsClient := NewAwsLocalSQS(NewDefaultAwsLocalConfig())

	// create a new queue
	queueResp, err := sqsClient.CreateQueue(context.TODO(), &sqs.CreateQueueInput{
		QueueName: aws.String("test-queue"),
	})
	if err != nil {
		panic(err)
	}

	// print the queue URL
	println("Queue URL:", *queueResp.QueueUrl)

	// send a message to the queue
	sendResp, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    queueResp.QueueUrl,
		MessageBody: aws.String("Hello, world!"),
	})
	if err != nil {
		panic(err)
	}

	// print the message ID
	println("Message ID:", *sendResp.MessageId)

	// receive a message from the queue
	receiveResp, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
		QueueUrl: queueResp.QueueUrl,
	})
	if err != nil {
		panic(err)
	}

	// print the message body
	for _, message := range receiveResp.Messages {
		println("Message Body:", *message.Body)
	}

	// delete the queue
	_, err = sqsClient.DeleteQueue(context.TODO(), &sqs.DeleteQueueInput{
		QueueUrl: queueResp.QueueUrl,
	})
	if err != nil {
		panic(err)
	}
}
