package queue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type QueueService interface {
	Produce(ctx context.Context, payload any, q string) error
	Consume(ctx context.Context, handle func(*string) error, q string) error
}

func ConstructSQSService(awsAccount, awsRegion string, waitTimeSeconds int32) *SQSService {
	if awsAccount == "" {
		panic("AWS_ACCOUNT is required for sqs provider")
	}
	if awsRegion == "" {
		panic("AWS_REGION is required for sqs provider")
	}

	s := &SQSService{
		AWSAccount:      awsAccount,
		AWSRegion:       awsRegion,
		WaitTimeSeconds: waitTimeSeconds,
		queueURLs:       make(map[string]string),
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s.AWSRegion),
	)
	if err != nil {
		panic(err)
	}
	s.client = sqs.NewFromConfig(cfg)

	return s
}

func ConstructQueueService(p QueueProvider, awsAccount, awsRegion string, waitTimeSeconds int32) QueueService {
	var svc QueueService
	switch p {
	case SQS:
		svc = ConstructSQSService(awsAccount, awsRegion, waitTimeSeconds)
	default:
		panic(fmt.Sprintf("%s is not a valid queue provider", p))
	}
	return svc
}
