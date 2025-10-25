package workers

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MockStorage - простая мок-реализация Storage
type MockStorage struct {
	// Поля для контроля вызовов
	GenerateUploadURLFunc   func(ctx context.Context, key string, expiresIn time.Duration) (string, error)
	GenerateDownloadURLFunc func(ctx context.Context, key string, expiresIn time.Duration) (string, error)
	DeleteFunc              func(ctx context.Context, key string) error
	UploadFileFunc          func(ctx context.Context, key string, fileData []byte) error
	DownloadFileFunc        func(ctx context.Context, key string) ([]byte, error)
}

func (m *MockStorage) GenerateUploadURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if m.GenerateUploadURLFunc != nil {
		return m.GenerateUploadURLFunc(ctx, key, expiresIn)
	}
	return "mock-upload-url", nil
}

func (m *MockStorage) GenerateDownloadURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if m.GenerateDownloadURLFunc != nil {
		return m.GenerateDownloadURLFunc(ctx, key, expiresIn)
	}
	return "mock-download-url", nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return nil
}

func (m *MockStorage) UploadFile(ctx context.Context, key string, fileData []byte) error {
	if m.UploadFileFunc != nil {
		return m.UploadFileFunc(ctx, key, fileData)
	}
	return nil
}

func (m *MockStorage) DownloadFile(ctx context.Context, key string) ([]byte, error) {
	if m.DownloadFileFunc != nil {
		return m.DownloadFileFunc(ctx, key)
	}
	return []byte("mock-data"), nil
}

// MockPreviewProducer - мок для PreviewProducer
type MockPreviewProducer struct {
	ProduceFunc func(ctx context.Context, versionID uuid.UUID) error

	// Для отслеживания вызовов
	ProduceCalls []ProduceCall
}

type ProduceCall struct {
	Ctx       context.Context
	VersionID uuid.UUID
}

func (m *MockPreviewProducer) Produce(ctx context.Context, versionID uuid.UUID) error {
	m.ProduceCalls = append(m.ProduceCalls, ProduceCall{
		Ctx:       ctx,
		VersionID: versionID,
	})

	if m.ProduceFunc != nil {
		return m.ProduceFunc(ctx, versionID)
	}
	return nil
}

// MockPreviewConsumer - мок для PreviewConsumer
type MockPreviewConsumer struct {
	ConsumeFunc func(ctx context.Context) (uuid.UUID, error)
	RemoveFunc  func(ctx context.Context, versionID uuid.UUID) error

	// Для отслеживания вызовов
	ConsumeCalls []ConsumeCall
	RemoveCalls  []RemoveCall
}

type ConsumeCall struct {
	Ctx context.Context
}

type RemoveCall struct {
	Ctx       context.Context
	VersionID uuid.UUID
}

func (m *MockPreviewConsumer) Consume(ctx context.Context) (uuid.UUID, error) {
	m.ConsumeCalls = append(m.ConsumeCalls, ConsumeCall{
		Ctx: ctx,
	})

	if m.ConsumeFunc != nil {
		return m.ConsumeFunc(ctx)
	}
	return uuid.New(), nil
}

func (m *MockPreviewConsumer) Remove(ctx context.Context, versionID uuid.UUID) error {
	m.RemoveCalls = append(m.RemoveCalls, RemoveCall{
		Ctx:       ctx,
		VersionID: versionID,
	})

	if m.RemoveFunc != nil {
		return m.RemoveFunc(ctx, versionID)
	}
	return nil
}
