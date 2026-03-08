package blob

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockS3Client struct {
	putObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	lastInput     *s3.PutObjectInput
}

func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.lastInput = params
	if m.putObjectFunc != nil {
		return m.putObjectFunc(ctx, params, optFns...)
	}
	return &s3.PutObjectOutput{}, nil
}

func TestS3Store_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockS3Client{}
		store := &S3Store{bucket: "test-bucket", client: mock, log: slog.Default()}

		data := []byte(`{"temp": 18.0}`)
		err := store.Save("openweather/London/1700000000.json", data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mock.lastInput == nil {
			t.Fatal("PutObject was not called")
		}
		if *mock.lastInput.Bucket != "test-bucket" {
			t.Errorf("Bucket = %q, want %q", *mock.lastInput.Bucket, "test-bucket")
		}
		if *mock.lastInput.Key != "openweather/London/1700000000.json" {
			t.Errorf("Key = %q, want %q", *mock.lastInput.Key, "openweather/London/1700000000.json")
		}
		if *mock.lastInput.ContentType != "application/json" {
			t.Errorf("ContentType = %q, want %q", *mock.lastInput.ContentType, "application/json")
		}

		body, _ := io.ReadAll(mock.lastInput.Body)
		if string(body) != `{"temp": 18.0}` {
			t.Errorf("Body = %q, want %q", string(body), `{"temp": 18.0}`)
		}
	})

	t.Run("put object error", func(t *testing.T) {
		putErr := errors.New("access denied")
		mock := &mockS3Client{
			putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				return nil, putErr
			},
		}
		store := &S3Store{bucket: "test-bucket", client: mock, log: slog.Default()}

		err := store.Save("key.json", []byte(`{}`))
		if !errors.Is(err, putErr) {
			t.Fatalf("expected put error, got %v", err)
		}
	})
}
