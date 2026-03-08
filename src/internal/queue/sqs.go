package queue

import (
	"context"
	"encoding/json"
	"log"

	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const SQS_QUEUE_URL = "https://sqs.%s.amazonaws.com/%s/%s"

type sqsAPI interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type SQSService struct {
	AWSAccount string
	AWSRegion  string
	client     sqsAPI
	ctx        context.Context
}

func (s *SQSService) Produce(ctx context.Context, payload any, q string) error {

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(SQS_QUEUE_URL, s.AWSRegion, s.AWSAccount, q)

	_, err = s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(url),
		MessageBody: aws.String(string(b)),
	})
	return err
}

func (s *SQSService) Consume(ctx context.Context, handle func(*string) error, q string) error {
	url := fmt.Sprintf(SQS_QUEUE_URL, s.AWSRegion, s.AWSAccount, q)

	for {
		output, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl: aws.String(q),
		})

		if err != nil {
			log.Printf("error receiving messages: %v", err)
			return err
		}

		for _, msg := range output.Messages {
			if err := handle(msg.Body); err != nil {
				log.Printf("error handling message: %v", err)
				return err
			}

			// delete after successful processing
			_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(url),
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Printf("error deleting message: %v", err)
				return err
			}
		}
	}

}
