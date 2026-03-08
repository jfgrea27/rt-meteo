package queue

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type mockSQSClient struct {
	sendMessageFunc    func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	receiveMessageFunc func(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	deleteMessageFunc  func(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	getQueueUrlFunc    func(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
}

func (m *mockSQSClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return m.sendMessageFunc(ctx, params, optFns...)
}

func (m *mockSQSClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	return m.receiveMessageFunc(ctx, params, optFns...)
}

func (m *mockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	return m.deleteMessageFunc(ctx, params, optFns...)
}

func (m *mockSQSClient) GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
	return m.getQueueUrlFunc(ctx, params, optFns...)
}

func defaultGetQueueUrlFunc() func(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
	return func(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
		url := "http://localhost:4566/000000000000/" + *params.QueueName
		return &sqs.GetQueueUrlOutput{QueueUrl: aws.String(url)}, nil
	}
}

func newTestSQSService(mock *mockSQSClient) *SQSService {
	return &SQSService{
		AWSAccount: "123456789012",
		AWSRegion:  "us-east-1",
		client:     mock,
		queueURLs:  make(map[string]string),
	}
}

func TestProduce_Success(t *testing.T) {
	payload := map[string]string{"key": "value"}
	var capturedInput *sqs.SendMessageInput

	mock := &mockSQSClient{
		sendMessageFunc: func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
			capturedInput = params
			return &sqs.SendMessageOutput{}, nil
		},
		getQueueUrlFunc: defaultGetQueueUrlFunc(),
	}

	svc := newTestSQSService(mock)
	err := svc.Produce(context.Background(), payload, "test-queue")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if capturedInput == nil {
		t.Fatal("SendMessage was not called")
	}

	// Verify the message body is valid JSON matching the payload
	var got map[string]string
	if err := json.Unmarshal([]byte(*capturedInput.MessageBody), &got); err != nil {
		t.Fatalf("message body is not valid JSON: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", got["key"])
	}

	// Verify queue URL was resolved and passed through
	if capturedInput.QueueUrl == nil {
		t.Fatal("expected QueueUrl to be set")
	}
	expectedURL := "http://localhost:4566/000000000000/test-queue"
	if *capturedInput.QueueUrl != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, *capturedInput.QueueUrl)
	}
}

func TestProduce_MarshalError(t *testing.T) {
	mock := &mockSQSClient{
		getQueueUrlFunc: defaultGetQueueUrlFunc(),
	}
	svc := newTestSQSService(mock)

	// channels cannot be marshaled to JSON
	err := svc.Produce(context.Background(), make(chan int), "test-queue")
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

func TestProduce_SendMessageError(t *testing.T) {
	sendErr := errors.New("send failed")
	mock := &mockSQSClient{
		sendMessageFunc: func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
			return nil, sendErr
		},
		getQueueUrlFunc: defaultGetQueueUrlFunc(),
	}

	svc := newTestSQSService(mock)
	err := svc.Produce(context.Background(), "hello", "test-queue")

	if !errors.Is(err, sendErr) {
		t.Fatalf("expected send error, got %v", err)
	}
}

func TestConsume_ProcessesAndDeletesMessages(t *testing.T) {
	callCount := 0
	var deletedReceipt *string

	mock := &mockSQSClient{
		getQueueUrlFunc: defaultGetQueueUrlFunc(),
		receiveMessageFunc: func(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
			callCount++
			if callCount == 1 {
				return &sqs.ReceiveMessageOutput{
					Messages: []types.Message{
						{
							Body:          aws.String("message-body"),
							ReceiptHandle: aws.String("receipt-1"),
						},
					},
				}, nil
			}
			// Return context canceled on second call to exit the loop
			return nil, context.Canceled
		},
		deleteMessageFunc: func(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
			deletedReceipt = params.ReceiptHandle
			return &sqs.DeleteMessageOutput{}, nil
		},
	}

	svc := newTestSQSService(mock)

	var handledBody *string
	handler := func(body *string) error {
		handledBody = body
		return nil
	}

	err := svc.Consume(context.Background(), handler, "test-queue")

	// We expect context.Canceled from the second ReceiveMessage call
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if handledBody == nil || *handledBody != "message-body" {
		t.Error("handler was not called with the correct message body")
	}

	if deletedReceipt == nil || *deletedReceipt != "receipt-1" {
		t.Error("message was not deleted with the correct receipt handle")
	}
}

func TestConsume_ReceiveMessageError(t *testing.T) {
	recvErr := errors.New("receive failed")
	mock := &mockSQSClient{
		getQueueUrlFunc: defaultGetQueueUrlFunc(),
		receiveMessageFunc: func(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
			return nil, recvErr
		},
	}

	svc := newTestSQSService(mock)
	err := svc.Consume(context.Background(), func(s *string) error { return nil }, "test-queue")

	if !errors.Is(err, recvErr) {
		t.Fatalf("expected receive error, got %v", err)
	}
}

func TestConsume_HandlerError(t *testing.T) {
	handleErr := errors.New("handle failed")
	mock := &mockSQSClient{
		getQueueUrlFunc: defaultGetQueueUrlFunc(),
		receiveMessageFunc: func(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
			return &sqs.ReceiveMessageOutput{
				Messages: []types.Message{
					{Body: aws.String("msg"), ReceiptHandle: aws.String("r1")},
				},
			}, nil
		},
	}

	svc := newTestSQSService(mock)
	err := svc.Consume(context.Background(), func(s *string) error { return handleErr }, "test-queue")

	if !errors.Is(err, handleErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func TestConsume_DeleteMessageError(t *testing.T) {
	deleteErr := errors.New("delete failed")
	callCount := 0

	mock := &mockSQSClient{
		getQueueUrlFunc: defaultGetQueueUrlFunc(),
		receiveMessageFunc: func(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
			callCount++
			if callCount == 1 {
				return &sqs.ReceiveMessageOutput{
					Messages: []types.Message{
						{Body: aws.String("msg"), ReceiptHandle: aws.String("r1")},
					},
				}, nil
			}
			return nil, context.Canceled
		},
		deleteMessageFunc: func(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
			return nil, deleteErr
		},
	}

	svc := newTestSQSService(mock)
	err := svc.Consume(context.Background(), func(s *string) error { return nil }, "test-queue")

	if !errors.Is(err, deleteErr) {
		t.Fatalf("expected delete error, got %v", err)
	}
}

func TestGetQueueURL_Success(t *testing.T) {
	var capturedName *string
	mock := &mockSQSClient{
		getQueueUrlFunc: func(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
			capturedName = params.QueueName
			return &sqs.GetQueueUrlOutput{
				QueueUrl: aws.String("http://localhost:4566/000000000000/weather"),
			}, nil
		},
	}

	svc := newTestSQSService(mock)
	url, err := svc.getQueueURL(context.Background(), "weather")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if url != "http://localhost:4566/000000000000/weather" {
		t.Errorf("expected http://localhost:4566/000000000000/weather, got %s", url)
	}
	if capturedName == nil || *capturedName != "weather" {
		t.Error("expected GetQueueUrl to be called with queue name 'weather'")
	}
}

func TestGetQueueURL_CachesResult(t *testing.T) {
	callCount := 0
	mock := &mockSQSClient{
		getQueueUrlFunc: func(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
			callCount++
			return &sqs.GetQueueUrlOutput{
				QueueUrl: aws.String("http://localhost:4566/000000000000/" + *params.QueueName),
			}, nil
		},
	}

	svc := newTestSQSService(mock)
	ctx := context.Background()

	url1, err := svc.getQueueURL(ctx, "weather")
	if err != nil {
		t.Fatalf("first call: expected no error, got %v", err)
	}

	url2, err := svc.getQueueURL(ctx, "weather")
	if err != nil {
		t.Fatalf("second call: expected no error, got %v", err)
	}

	if url1 != url2 {
		t.Errorf("expected same URL, got %s and %s", url1, url2)
	}
	if callCount != 1 {
		t.Errorf("expected GetQueueUrl to be called once, got %d", callCount)
	}
}

func TestGetQueueURL_DifferentQueuesNotCached(t *testing.T) {
	callCount := 0
	mock := &mockSQSClient{
		getQueueUrlFunc: func(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
			callCount++
			return &sqs.GetQueueUrlOutput{
				QueueUrl: aws.String("http://localhost:4566/000000000000/" + *params.QueueName),
			}, nil
		},
	}

	svc := newTestSQSService(mock)
	ctx := context.Background()

	url1, _ := svc.getQueueURL(ctx, "queue-a")
	url2, _ := svc.getQueueURL(ctx, "queue-b")

	if url1 == url2 {
		t.Error("expected different URLs for different queues")
	}
	if callCount != 2 {
		t.Errorf("expected GetQueueUrl to be called twice, got %d", callCount)
	}
}

func TestGetQueueURL_Error(t *testing.T) {
	apiErr := errors.New("queue not found")
	mock := &mockSQSClient{
		getQueueUrlFunc: func(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
			return nil, apiErr
		},
	}

	svc := newTestSQSService(mock)
	_, err := svc.getQueueURL(context.Background(), "nonexistent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, apiErr) {
		t.Fatalf("expected wrapped api error, got %v", err)
	}
}

func TestGetQueueURL_ErrorNotCached(t *testing.T) {
	callCount := 0
	mock := &mockSQSClient{
		getQueueUrlFunc: func(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) {
			callCount++
			if callCount == 1 {
				return nil, errors.New("temporary failure")
			}
			return &sqs.GetQueueUrlOutput{
				QueueUrl: aws.String("http://localhost:4566/000000000000/" + *params.QueueName),
			}, nil
		},
	}

	svc := newTestSQSService(mock)
	ctx := context.Background()

	_, err := svc.getQueueURL(ctx, "weather")
	if err == nil {
		t.Fatal("first call: expected error")
	}

	url, err := svc.getQueueURL(ctx, "weather")
	if err != nil {
		t.Fatalf("second call: expected no error, got %v", err)
	}
	if url != "http://localhost:4566/000000000000/weather" {
		t.Errorf("expected resolved URL, got %s", url)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (error should not be cached), got %d", callCount)
	}
}
