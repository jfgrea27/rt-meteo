package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type sqsAPI interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
}

type SQSService struct {
	AWSAccount      string
	AWSRegion       string
	WaitTimeSeconds int32
	client          sqsAPI
	ctx             context.Context
	queueURLs       map[string]string
}

func (s *SQSService) getQueueURL(ctx context.Context, q string) (string, error) {
	if url, ok := s.queueURLs[q]; ok {
		return url, nil
	}

	out, err := s.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(q),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get queue URL for %q: %w", q, err)
	}
	s.queueURLs[q] = *out.QueueUrl
	return *out.QueueUrl, nil
}

func (s *SQSService) Produce(ctx context.Context, payload any, q string) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url, err := s.getQueueURL(ctx, q)
	if err != nil {
		return err
	}

	_, err = s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(url),
		MessageBody: aws.String(string(b)),
	})
	return err
}

func (s *SQSService) Consume(ctx context.Context, handle func(*string) error, q string) error {
	url, err := s.getQueueURL(ctx, q)
	if err != nil {
		return err
	}

	for {
		output, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:        aws.String(url),
			WaitTimeSeconds: s.WaitTimeSeconds,
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
