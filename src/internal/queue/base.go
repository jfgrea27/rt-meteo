package queue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jfgrea27/rt-meteo/internal/utils"
)

type QueueService interface {
	Produce(ctx context.Context, payload any, q string) error
	Consume(ctx context.Context, handle func(*string) error, q string) error
}

func ConstructSQSService() *SQSService {
	awsAccount := utils.GetEnvVar("AWS_ACCOUNT", true)
	awsRegion := utils.GetEnvVar("AWS_REGION", true)

	s := &SQSService{
		AWSAccount: awsAccount,
		AWSRegion:  awsRegion,
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

func ConstructQueueService(p QueueProvider) QueueService {
	var svc QueueService
	switch p {
	case SQS:
		svc = ConstructSQSService()
	default:
		panic(fmt.Sprintf("%s is not a valid queue provider", p))
	}
	return svc
}
